package main

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"strings"
	"time"
)

type KeyBindings map[string][]tcell.Key

var keyBindings = KeyBindings{
	"search":           {tcell.KeyF2, tcell.KeyCtrlS},
	"keys":             {tcell.KeyF3, tcell.KeyCtrlK},
	"key_list_value":   {tcell.KeyF6, tcell.KeyCtrlY},
	"key_string_value": {tcell.KeyF7, tcell.KeyCtrlA},
	"key_hash":         {tcell.KeyF6, tcell.KeyCtrlY},
	"output":           {tcell.KeyF9, tcell.KeyCtrlO},
	"command":          {tcell.KeyF1, tcell.KeyCtrlN},
	"command_focus":    {tcell.KeyF4, tcell.KeyCtrlF},
	"command_result":   {tcell.KeyF5, tcell.KeyCtrlR},
	"quit":             {tcell.KeyEsc, tcell.KeyCtrlQ},
	"switch_focus":     {tcell.KeyTab},
}

func (kb KeyBindings) SearchKey(k tcell.Key) string {
	for name, bind := range kb {
		for _, b := range bind {
			if b == k {
				return name
			}
		}
	}

	return ""
}

func (kb KeyBindings) KeyID(key string) string {
	return key
}

func (kb KeyBindings) Keys(key string) []tcell.Key {
	return kb[key]
}

func (kb KeyBindings) Name(key string) string {
	keyNames := make([]string, 0)
	for _, k := range kb[key] {
		keyNames = append(keyNames, tcell.KeyNames[k])
	}

	return strings.Join(keyNames, ", ")
}

type primitiveKey struct {
	Primitive tview.Primitive
	Key       string
}

type OutputMessage struct {
	Color   tcell.Color
	Message string
}

// RedisTUI is a redis gui object
type RedisTUI struct {
	metaPanel           *tview.TextView
	mainPanel           *tview.Flex
	outputPanel         *tview.List
	keyItemsPanel       *tview.List
	summaryPanel        *tview.TextView
	searchPanel         *tview.InputField
	welcomeScreen       tview.Primitive
	helpPanel           *tview.Flex
	helpMessagePanel    *tview.TextView
	helpServerInfoPanel *tview.TextView

	commandPanel       *tview.Flex
	commandFormPanel   *tview.InputField
	commandResultPanel *tview.TextView
	commandMode        bool

	leftPanel  *tview.Flex
	rightPanel *tview.Flex

	layout *tview.Flex
	pages  *tview.Pages
	app    *tview.Application

	redisClient RedisClient
	outputChan  chan OutputMessage

	itemSelectedHandler func(index int, key string) func()

	maxKeyLimit int64
	version     string
	gitCommit   string

	focusPrimitives   []primitiveKey
	currentFocusIndex int

	config Config
}

// NewRedisTUI create a RedisTUI object
func NewRedisTUI(redisClient RedisClient, maxKeyLimit int64, version string, gitCommit string, outputChan chan OutputMessage, config Config) *RedisTUI {
	tui := &RedisTUI{
		redisClient:       redisClient,
		maxKeyLimit:       maxKeyLimit,
		version:           version,
		gitCommit:         gitCommit[0:8],
		focusPrimitives:   make([]primitiveKey, 0),
		currentFocusIndex: 0,
		outputChan:        outputChan,
		config:            config,
	}

	tui.welcomeScreen = tview.NewTextView().SetTitle("Hello, world!")

	tui.metaPanel = tui.createMetaPanel()
	tui.mainPanel = tui.createMainPanel()
	tui.outputPanel = tui.createOutputPanel()
	tui.summaryPanel = tui.createSummaryPanel()
	tui.keyItemsPanel = tui.createKeyItemsPanel()
	tui.itemSelectedHandler = tui.createKeySelectedHandler()
	tui.searchPanel = tui.createSearchPanel()
	tui.helpPanel = tui.createHelpPanel()

	tui.commandPanel = tui.createCommandPanel()

	tui.leftPanel = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tui.searchPanel, 3, 0, false).
		AddItem(tui.keyItemsPanel, 0, 1, false).
		AddItem(tui.summaryPanel, 3, 1, false)

	tui.rightPanel = tview.NewFlex().SetDirection(tview.FlexRow)
	tui.redrawRightPanel(tui.mainPanel)

	tui.app = tview.NewApplication()
	tui.layout = tview.NewFlex().
		AddItem(tui.leftPanel, 0, 3, false).
		AddItem(tui.rightPanel, 0, 8, false)

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.searchPanel, Key: keyBindings.KeyID("search")})
	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.keyItemsPanel, Key: keyBindings.KeyID("keys")})

	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if tui.config.Debug {
			tui.outputChan <- OutputMessage{Message: fmt.Sprintf("Key %s pressed", tcell.KeyNames[event.Key()])}
		}

		name := keyBindings.SearchKey(event.Key())
		switch name {
		case "switch_focus":
			nextFocusIndex := tui.currentFocusIndex + 1
			if nextFocusIndex > len(tui.focusPrimitives)-1 {
				nextFocusIndex = 0
			}

			tui.app.SetFocus(tui.focusPrimitives[nextFocusIndex].Primitive)
			tui.currentFocusIndex = nextFocusIndex

			return nil
		case "quit":
			tui.app.Stop()
		case "command":
			if tui.commandMode {
				tui.commandMode = false
				tui.redrawRightPanel(tui.mainPanel)
				tui.app.SetFocus(tui.searchPanel)
				tui.currentFocusIndex = 0
			} else {
				tui.commandMode = true
				tui.redrawRightPanel(tui.commandPanel)

				for i, p := range tui.focusPrimitives {
					if p.Primitive == tui.commandFormPanel {
						tui.app.SetFocus(tui.commandFormPanel)
						tui.currentFocusIndex = i
					}
				}
			}
		default:
			for i, pv := range tui.focusPrimitives {
				if pv.Key == name {
					tui.app.SetFocus(pv.Primitive)
					tui.currentFocusIndex = i
					break
				}
			}
		}

		return event
	})

	return tui
}

// Start create the ui and start the program
func (tui *RedisTUI) Start() error {
	go tui.app.QueueUpdateDraw(func() {
		info, err := RedisServerInfo(tui.config, tui.redisClient)
		if err != nil {
			tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
		}

		tui.helpServerInfoPanel.SetText(info)

		keys, _, err := tui.redisClient.Scan(0, "*", tui.maxKeyLimit).Result()
		if err != nil {
			tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
			return
		}

		tui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))

		for i, k := range keys {
			tui.keyItemsPanel.AddItem(tui.keyItemsFormat(i, k), "", 0, tui.itemSelectedHandler(i, k))
		}

		tui.app.SetFocus(tui.keyItemsPanel)
	})

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case out := <-tui.outputChan:
				(func(out OutputMessage) {
					tui.app.QueueUpdateDraw(func() {
						// tui.outputPanel.SetTextColor(out.Color).SetText(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), out.Message))
						tui.outputPanel.AddItem(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), out.Message), "", 0, nil)
						tui.outputPanel.SetCurrentItem(-1)
					})
				})(out)
			case <-ticker.C:
				tui.app.QueueUpdateDraw(func() {
					info, err := RedisServerInfo(tui.config, tui.redisClient)
					if err != nil {
						tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					}

					tui.helpServerInfoPanel.SetText(info)
				})
			}
		}
	}()

	tui.pages = tview.NewPages()
	tui.pages.AddPage("base", tui.layout, true, true)

	return tui.app.SetRoot(tui.pages, true).Run()
}

func (tui *RedisTUI) redrawRightPanel(center tview.Primitive) {
	tui.rightPanel.RemoveItem(tui.metaPanel).
		RemoveItem(tui.outputPanel).
		RemoveItem(tui.mainPanel).
		RemoveItem(tui.commandPanel).
		RemoveItem(tui.helpPanel)

	tui.rightPanel.AddItem(tui.helpPanel, 5, 1, false).
		AddItem(tui.metaPanel, 4, 1, false).
		AddItem(center, 0, 7, false).
		AddItem(tui.outputPanel, 8, 1, false)
}

func (tui *RedisTUI) createSummaryPanel() *tview.TextView {
	panel := tview.NewTextView()
	panel.SetBorder(true).SetTitle(" Info ")
	return panel
}

func (tui *RedisTUI) keyItemsFormat(index int, key string) string {
	return fmt.Sprintf("%3d | %s", index+1, key)
}

func (tui *RedisTUI) createCommandPanel() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	resultPanel := tview.NewTextView()
	resultPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Results (%s) ", keyBindings.Name("command_result")))

	formPanel := tview.NewInputField().SetLabel("Command ")
	var locked bool
	formPanel.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		if locked {
			pageID := "alert"
			tui.pages.AddPage(
				pageID,
				tview.NewModal().
					SetText("Other command is processing, please wait...").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						tui.pages.HidePage(pageID).RemovePage(pageID)
						tui.app.SetFocus(formPanel)
					}),
				false,
				true,
			)

			return
		}

		cmdText := formPanel.GetText()
		tui.outputChan <- OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("Command %s is processing...", cmdText)}
		locked = true

		go func(cmdText string) {
			tui.app.QueueUpdateDraw(func() {
				defer func() {
					locked = false
				}()

				res, err := RedisExecute(tui.redisClient, cmdText)
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				resultPanel.SetText(fmt.Sprintf("%v", res))
				tui.outputChan <- OutputMessage{tcell.ColorGreen, fmt.Sprintf("Command %s succeed", cmdText)}
			})
		}(cmdText)

		formPanel.SetText("")
	})
	// formPanel.SetBackgroundColor(tcell.ColorOrange)
	formPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Commands (%s) ", keyBindings.Name("command_focus")))

	tui.commandFormPanel = formPanel
	tui.commandResultPanel = resultPanel

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.commandFormPanel, Key: keyBindings.KeyID("command_focus")})
	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.commandResultPanel, Key: keyBindings.KeyID("command_result")})

	flex.AddItem(formPanel, 4, 0, false).
		AddItem(resultPanel, 0, 1, false)

	return flex
}

// createSearchPanel create search panel
func (tui *RedisTUI) createSearchPanel() *tview.InputField {
	searchArea := tview.NewInputField().SetLabel(" KeyID ")
	searchArea.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		var text = searchArea.GetText()

		var keys []string
		var err error

		if text == "" || text == "*" {
			keys, _, err = tui.redisClient.Scan(0, text, tui.maxKeyLimit).Result()
		} else {
			keys, err = tui.redisClient.Keys(text).Result()
		}

		if err != nil {
			tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
			return
		}

		tui.keyItemsPanel.Clear()

		tui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))
		for i, k := range keys {
			tui.keyItemsPanel.AddItem(tui.keyItemsFormat(i, k), "", 0, tui.itemSelectedHandler(i, k))
		}
	})
	searchArea.SetBorder(true).SetTitle(fmt.Sprintf(" Search (%s) ", keyBindings.Name("search")))
	return searchArea
}

// createKeyItemsPanel create key items panel
func (tui *RedisTUI) createKeyItemsPanel() *tview.List {
	keyItemsList := tview.NewList().ShowSecondaryText(false)
	keyItemsList.SetBorder(true).SetTitle(fmt.Sprintf(" Keys (%s) ", keyBindings.Name("keys")))
	return keyItemsList
}

// primitivesFilter is a filter for primitives
func (tui *RedisTUI) primitivesFilter(items []primitiveKey, filter func(item primitiveKey) bool) []primitiveKey {
	res := make([]primitiveKey, 0)
	for _, item := range items {
		if filter(item) {
			res = append(res, item)
		}
	}

	return res
}

// createMetaPanel create a panel for meta info
func (tui *RedisTUI) createMetaPanel() *tview.TextView {
	metaInfoArea := tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	metaInfoArea.SetBorder(true).SetTitle(" Meta ")

	return metaInfoArea
}

// createMainPanel create main panel
func (tui *RedisTUI) createMainPanel() *tview.Flex {
	mainArea := tview.NewFlex()
	mainArea.SetBorder(true).SetTitle(" Value ")

	mainArea.AddItem(tui.welcomeScreen, 0, 1, false)

	return mainArea
}

// createOutputPanel create a panel for outputFunc
func (tui *RedisTUI) createOutputPanel() *tview.List {
	outputArea := tview.NewList().ShowSecondaryText(false)
	outputArea.SetBorder(true).SetTitle(fmt.Sprintf(" Output (%s) ", keyBindings.Name("output")))

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: outputArea, Key: keyBindings.KeyID("output")})

	return outputArea
}

// createHelpPanel create a panel for help message display
func (tui *RedisTUI) createHelpPanel() *tview.Flex {
	helpPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	helpPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Version: %s (%s) ", tui.version, tui.gitCommit))

	tui.helpServerInfoPanel = tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	helpPanel.AddItem(tui.helpServerInfoPanel, 2, 1, false)

	tui.helpMessagePanel = tview.NewTextView()
	tui.helpMessagePanel.SetTextColor(tcell.ColorOrange).SetText(fmt.Sprintf(
		" ❈ %s - open command panel, %s - switch focus, %s - quit",
		keyBindings.Name("command"),
		keyBindings.Name("switch_focus"),
		keyBindings.Name("quit"),
	))

	helpPanel.AddItem(tui.helpMessagePanel, 1, 1, false)

	return helpPanel
}

// createKeySelectedHandler create a handler for item selected event
func (tui *RedisTUI) createKeySelectedHandler() func(index int, key string) func() {

	// 用于KV展示的视图
	mainStringView := tview.NewTextView()
	mainStringView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", keyBindings.Name("key_string_value")))

	mainHashView := tview.NewList().ShowSecondaryText(false)
	mainHashView.SetBorder(true).SetTitle(fmt.Sprintf(" Hash KeyID (%s) ", keyBindings.Name("key_hash")))

	mainListView := tview.NewList().ShowSecondaryText(false).SetSecondaryTextColor(tcell.ColorOrangeRed)
	mainListView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", keyBindings.Name("key_list_value")))

	return func(index int, key string) func() {
		return func() {
			keyType, err := tui.redisClient.Type(key).Result()
			if err != nil {
				tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
				return
			}

			ttl, err := tui.redisClient.TTL(key).Result()
			if err != nil {
				tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
				return
			}

			// 移除主区域的边框，因为展示区域已经带有边框了
			tui.mainPanel.RemoveItem(tui.welcomeScreen).SetBorder(false)

			// 重置展示视图
			mainHashView.Clear()
			mainStringView.Clear()
			mainListView.Clear().ShowSecondaryText(false)

			tui.focusPrimitives = tui.primitivesFilter(tui.focusPrimitives, func(item primitiveKey) bool {
				return item.Primitive != mainHashView && item.Primitive != mainListView && item.Primitive != mainStringView
			})

			tui.mainPanel.RemoveItem(mainStringView)
			tui.mainPanel.RemoveItem(mainHashView)
			tui.mainPanel.RemoveItem(mainListView)

			// 根据不同的kv类型，展示不同的视图
			switch keyType {
			case "string":
				result, err := tui.redisClient.Get(key).Result()
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				tui.mainPanel.AddItem(mainStringView.SetText(fmt.Sprintf(" %s", result)), 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: keyBindings.KeyID("key_string_value")})
			case "list":
				values, err := tui.redisClient.LRange(key, 0, 1000).Result()
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: keyBindings.KeyID("key_list_value")})

			case "set":
				values, err := tui.redisClient.SMembers(key).Result()
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: keyBindings.KeyID("key_list_value")})

			case "zset":
				values, err := tui.redisClient.ZRangeWithScores(key, 0, 1000).Result()
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				mainListView.ShowSecondaryText(true)
				for i, z := range values {
					val := fmt.Sprintf(" %3d | %v", i+1, z.Member)
					score := fmt.Sprintf("    Score: %v", z.Score)

					mainListView.AddItem(val, score, 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: keyBindings.KeyID("key_list_value")})

			case "hash":
				hashKeys, err := tui.redisClient.HKeys(key).Result()
				if err != nil {
					tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, k := range hashKeys {
					mainHashView.AddItem(fmt.Sprintf(" %3d | %s", i+1, k), "", 0, (func(k string) func() {
						return func() {
							val, err := tui.redisClient.HGet(key, k).Result()
							if err != nil {
								tui.outputChan <- OutputMessage{tcell.ColorRed, fmt.Sprintf("errors: %s", err)}
								return
							}

							mainStringView.SetText(fmt.Sprintf(" %s", val)).SetTitle(fmt.Sprintf(" Value: %s ", k))
						}
					})(k))
				}

				tui.mainPanel.AddItem(mainHashView, 0, 3, false).
					AddItem(mainStringView, 0, 7, false)

				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainHashView, Key: keyBindings.KeyID("key_hash")})
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: keyBindings.KeyID("key_string_value")})
			}
			tui.outputChan <- OutputMessage{tcell.ColorGreen, fmt.Sprintf("query %s OK, type=%s, ttl=%s", key, keyType, ttl.String())}
			tui.metaPanel.SetText(fmt.Sprintf("KeyID: %s\nType: %s, TTL: %s", key, keyType, ttl.String())).SetTextAlign(tview.AlignCenter)
		}
	}
}
