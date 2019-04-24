package tui

import (
	"fmt"
	"github.com/gdamore/tcell"
	"github.com/mylxsw/go-toolkit/collection"
	"github.com/mylxsw/redis-tui/api"
	"github.com/mylxsw/redis-tui/config"
	"github.com/mylxsw/redis-tui/core"
	"github.com/rivo/tview"
	"strings"
	"time"
)

type primitiveKey struct {
	Primitive tview.Primitive
	Key       string
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
	commandInputField  *tview.InputField
	commandResultPanel *tview.TextView
	commandMode        bool

	leftPanel  *tview.Flex
	rightPanel *tview.Flex

	layout *tview.Flex
	pages  *tview.Pages
	app    *tview.Application

	redisClient      api.RedisClient
	outputChan       chan core.OutputMessage
	uiViewUpdateChan chan func()

	itemSelectedHandler func(index int, key string) func()

	maxKeyLimit       int64
	maxCharacterLimit int64

	version   string
	gitCommit string

	focusPrimitives   []primitiveKey
	currentFocusIndex int

	config      config.Config
	keyBindings core.KeyBindings
}

// NewRedisTUI create a RedisTUI object
func NewRedisTUI(redisClient api.RedisClient, maxKeyLimit int64, version string, gitCommit string, outputChan chan core.OutputMessage, conf config.Config) *RedisTUI {
	tui := &RedisTUI{
		redisClient:       redisClient,
		maxKeyLimit:       maxKeyLimit,
		maxCharacterLimit: maxKeyLimit * 20,
		version:           version,
		gitCommit:         gitCommit,
		focusPrimitives:   make([]primitiveKey, 0),
		currentFocusIndex: 0,
		outputChan:        outputChan,
		config:            conf,
		keyBindings:       core.NewKeyBinding(),
		uiViewUpdateChan:  make(chan func()),
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

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.searchPanel, Key: tui.keyBindings.KeyID("search")})
	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.keyItemsPanel, Key: tui.keyBindings.KeyID("keys")})

	tui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if tui.config.Debug {
			tui.outputChan <- core.OutputMessage{Message: fmt.Sprintf("Key %s pressed", tcell.KeyNames[event.Key()])}
		}

		name := tui.keyBindings.SearchKey(event.Key())
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
					if p.Primitive == tui.commandInputField {
						tui.app.SetFocus(tui.commandInputField)
						tui.currentFocusIndex = i
						break
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
	go func() {
		for {
			select {
			case f := <-tui.uiViewUpdateChan:
				(func() {
					defer func() {
						if err := recover(); err != nil {
						}
					}()
					tui.app.QueueUpdateDraw(f)
				})()
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case out := <-tui.outputChan:
				tui.uiViewUpdateChan <- func() {
					// tui.outputPanel.SetTextColor(out.Color).SetText(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), out.Message))
					tui.outputPanel.AddItem(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), out.Message), "", 0, nil)
					tui.outputPanel.SetCurrentItem(-1)
				}
			case <-ticker.C:
				info, err := api.RedisServerInfo(tui.config, tui.redisClient)
				if err != nil {
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				}

				tui.uiViewUpdateChan <- func() {
					tui.helpServerInfoPanel.SetText(info)
				}
			}
		}
	}()

	go func() {
		info, err := api.RedisServerInfo(tui.config, tui.redisClient)
		if err != nil {
			tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
		}
		tui.app.QueueUpdateDraw(func() {
			tui.helpServerInfoPanel.SetText(info)
		})

		keys, _, err := tui.redisClient.Scan(0, "*", tui.maxKeyLimit).Result()
		if err != nil {
			tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
			return
		}
		tui.app.QueueUpdateDraw(func() {
			tui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))

			for i, k := range keys {
				tui.keyItemsPanel.AddItem(tui.keyItemsFormat(i, k), "", 0, tui.itemSelectedHandler(i, k))
			}

			tui.app.SetFocus(tui.keyItemsPanel)
		})
	}()

	tui.pages = tview.NewPages()
	tui.pages.AddPage("base", tui.layout, true, true)
	// welcomeScreen := tview.NewInputField().SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (i int, i2 int, i3 int, i4 int) {
	// 	// Draw a horizontal line across the middle of the box.
	// 	centerY := y + height/2
	// 	for cx := x + 1; cx < x+width-1; cx++ {
	// 		screen.SetContent(cx, centerY, tview.BoxDrawingsDoubleDownAndHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
	// 	}
	//
	// 	// Write some text along the horizontal line.
	// 	tview.Print(screen, " Hello, world! ", x+1, centerY, width-2, tview.AlignCenter, tcell.ColorYellow)
	//
	// 	// Space for other content.
	// 	return x + 1, centerY + 1, width - 2, height - (centerY + 1 - y)
	// })
	//
	// tui.pages.AddPage("welcome_screen", welcomeScreen, true, true)
	//
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	tui.app.QueueUpdateDraw(func() {
	// 		tui.pages.RemovePage("welcome_screen")
	// 	})
	// }()

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
	resultPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Results (%s) ", tui.keyBindings.Name("command_result")))

	formPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	formPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Commands (%s) ", tui.keyBindings.Name("command_focus")))
	commandTipView := tview.NewTextView().SetDynamicColors(true).SetRegions(true)

	commandInputField := tview.NewInputField().SetLabel("Command ")

	locked := make(chan interface{}, 1)
	locked <- struct{}{}
	commandInputField.SetDoneFunc(func(key tcell.Key) {
		select {
		case <-locked:

		default:
			pageID := "alert"
			tui.pages.AddPage(
				pageID,
				tview.NewModal().
					SetText("Other command is processing, please wait...").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						tui.pages.HidePage(pageID).RemovePage(pageID)
						tui.app.SetFocus(commandInputField)
					}),
				false,
				true,
			)

			return
		}

		cmdText := commandInputField.GetText()
		tui.outputChan <- core.OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("Command %s is processing...", cmdText)}

		go func(cmdText string) {
			defer func() {
				locked <- struct{}{}
			}()
			res, err := api.RedisExecute(tui.redisClient, cmdText)
			if err != nil {
				tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				return
			}

			// format redis output
			var output string

			switch res.(type) {
			case string:
				output = res.(string)
			case []interface{}:
				var resStrs = make([]string, len(res.([]interface{})))
				for i, v := range res.([]interface{}) {
					resStrs[i] = fmt.Sprintf("%v", v)
				}

				output = strings.Join(resStrs, "\n")
			default:
				output = fmt.Sprintf("%v", res)
			}

			// If the output content is too long, the interface will be suspended for a long time
			if len(output) > int(tui.maxCharacterLimit) {
				output = output[:tui.maxCharacterLimit] + fmt.Sprintf("\n\n ~ %d+ charactors omitted ~", len(output)-int(tui.maxCharacterLimit))
			}

			outputSlices := strings.Split(output, "\n")
			for i, v := range outputSlices {
				outputSlices[i] = fmt.Sprintf("%5d | %s", i+1, v)
			}

			output = strings.Join(outputSlices, "\n")

			tui.outputChan <- core.OutputMessage{Color: tcell.ColorGreen, Message: fmt.Sprintf("Command %s succeed", cmdText)}
			tui.uiViewUpdateChan <- func() {
				resultPanel.SetText(output)
			}
		}(cmdText)

		commandInputField.SetText("")
	}).SetChangedFunc(func(text string) {
		if text == "" {
			commandTipView.Clear()
			return
		}

		matchedCommands := api.RedisMatchedCommands(text)
		if len(matchedCommands) == 0 {
			commandTipView.Clear()
		} else if len(matchedCommands) == 1 {
			commandTipView.SetTextColor(tcell.ColorOrange).SetText(fmt.Sprintf(
				"\n%s %s\n    [green]%s (since %s).",
				matchedCommands[0].Command,
				matchedCommands[0].Args,
				matchedCommands[0].Desc,
				matchedCommands[0].Version,
			))
		} else {
			commandTipView.SetTextColor(tcell.ColorBlue).
				SetText(collection.MustNew(matchedCommands).Reduce(func(carry string, item api.RedisHelp) string {
					if carry == "" {
						return item.Command
					}

					return carry + ", " + item.Command
				}, "").(string))
		}
	})

	tui.commandInputField = commandInputField
	tui.commandResultPanel = resultPanel

	formPanel.AddItem(commandInputField, 1, 1, false).
		AddItem(commandTipView, 4, 4, false)

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.commandInputField, Key: tui.keyBindings.KeyID("command_focus")})
	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: tui.commandResultPanel, Key: tui.keyBindings.KeyID("command_result")})

	flex.AddItem(formPanel, 7, 0, false).
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
			tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
			return
		}

		tui.keyItemsPanel.Clear()

		tui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))
		for i, k := range keys {
			if i > int(tui.maxKeyLimit) {
				break
			}
			tui.keyItemsPanel.AddItem(tui.keyItemsFormat(i, k), "", 0, tui.itemSelectedHandler(i, k))
		}
	})
	searchArea.SetBorder(true).SetTitle(fmt.Sprintf(" Search (%s) ", tui.keyBindings.Name("search")))
	return searchArea
}

// createKeyItemsPanel create key items panel
func (tui *RedisTUI) createKeyItemsPanel() *tview.List {
	keyItemsList := tview.NewList().ShowSecondaryText(false)
	keyItemsList.SetBorder(true).SetTitle(fmt.Sprintf(" Keys (%s) ", tui.keyBindings.Name("keys")))
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
	outputArea.SetBorder(true).SetTitle(fmt.Sprintf(" Output (%s) ", tui.keyBindings.Name("output")))

	tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: outputArea, Key: tui.keyBindings.KeyID("output")})

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
		tui.keyBindings.Name("command"),
		tui.keyBindings.Name("switch_focus"),
		tui.keyBindings.Name("quit"),
	))

	helpPanel.AddItem(tui.helpMessagePanel, 1, 1, false)

	return helpPanel
}

// createKeySelectedHandler create a handler for item selected event
func (tui *RedisTUI) createKeySelectedHandler() func(index int, key string) func() {

	// 用于KV展示的视图
	mainStringView := tview.NewTextView()
	mainStringView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", tui.keyBindings.Name("key_string_value")))

	mainHashView := tview.NewList().ShowSecondaryText(false)
	mainHashView.SetBorder(true).SetTitle(fmt.Sprintf(" Hash KeyID (%s) ", tui.keyBindings.Name("key_hash")))

	mainListView := tview.NewList().ShowSecondaryText(false).SetSecondaryTextColor(tcell.ColorOrangeRed)
	mainListView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", tui.keyBindings.Name("key_list_value")))

	return func(index int, key string) func() {
		return func() {
			keyType, err := tui.redisClient.Type(key).Result()
			if err != nil {
				tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				return
			}

			ttl, err := tui.redisClient.TTL(key).Result()
			if err != nil {
				tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
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
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				tui.mainPanel.AddItem(mainStringView.SetText(fmt.Sprintf(" %s", result)), 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: tui.keyBindings.KeyID("key_string_value")})
			case "list":
				values, err := tui.redisClient.LRange(key, 0, 1000).Result()
				if err != nil {
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: tui.keyBindings.KeyID("key_list_value")})

			case "set":
				values, err := tui.redisClient.SMembers(key).Result()
				if err != nil {
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: tui.keyBindings.KeyID("key_list_value")})

			case "zset":
				values, err := tui.redisClient.ZRangeWithScores(key, 0, 1000).Result()
				if err != nil {
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				mainListView.ShowSecondaryText(true)
				for i, z := range values {
					val := fmt.Sprintf(" %3d | %v", i+1, z.Member)
					score := fmt.Sprintf("    Score: %v", z.Score)

					mainListView.AddItem(val, score, 0, nil)
				}

				tui.mainPanel.AddItem(mainListView, 0, 1, false)
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: tui.keyBindings.KeyID("key_list_value")})

			case "hash":
				hashKeys, err := tui.redisClient.HKeys(key).Result()
				if err != nil {
					tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, k := range hashKeys {
					mainHashView.AddItem(fmt.Sprintf(" %3d | %s", i+1, k), "", 0, (func(k string) func() {
						return func() {
							val, err := tui.redisClient.HGet(key, k).Result()
							if err != nil {
								tui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
								return
							}

							mainStringView.SetText(fmt.Sprintf(" %s", val)).SetTitle(fmt.Sprintf(" Value: %s ", k))
						}
					})(k))
				}

				tui.mainPanel.AddItem(mainHashView, 0, 3, false).
					AddItem(mainStringView, 0, 7, false)

				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainHashView, Key: tui.keyBindings.KeyID("key_hash")})
				tui.focusPrimitives = append(tui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: tui.keyBindings.KeyID("key_string_value")})
			}
			tui.outputChan <- core.OutputMessage{Color: tcell.ColorGreen, Message: fmt.Sprintf("query %s OK, type=%s, ttl=%s", key, keyType, ttl.String())}
			tui.metaPanel.SetText(fmt.Sprintf("KeyID: %s\nType: %s, TTL: %s", key, keyType, ttl.String())).SetTextAlign(tview.AlignCenter)
		}
	}
}
