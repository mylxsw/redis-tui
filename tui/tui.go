package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/mylxsw/go-toolkit/collection"
	"github.com/mylxsw/redis-tui/api"
	"github.com/mylxsw/redis-tui/config"
	"github.com/mylxsw/redis-tui/core"
	"github.com/rivo/tview"
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

	maxKeyLimit       int
	maxCharacterLimit int

	version   string
	gitCommit string

	focusPrimitives   []primitiveKey
	currentFocusIndex int

	config      config.Config
	keyBindings core.KeyBindings

	searchKeyHistories  []string
	commandKeyHistories []string
}

// NewRedisTUI create a RedisTUI object
func NewRedisTUI(redisClient api.RedisClient, maxKeyLimit int, version string, gitCommit string, outputChan chan core.OutputMessage, conf config.Config) *RedisTUI {
	ui := &RedisTUI{
		redisClient:         redisClient,
		maxKeyLimit:         maxKeyLimit,
		maxCharacterLimit:   maxKeyLimit * 20,
		version:             version,
		gitCommit:           gitCommit,
		focusPrimitives:     make([]primitiveKey, 0),
		currentFocusIndex:   0,
		outputChan:          outputChan,
		config:              conf,
		keyBindings:         core.NewKeyBinding(),
		uiViewUpdateChan:    make(chan func()),
		searchKeyHistories:  make([]string, 0),
		commandKeyHistories: make([]string, 0),
	}

	ui.welcomeScreen = tview.NewTextView().SetTitle("Hello, world!")

	ui.metaPanel = ui.createMetaPanel()
	ui.mainPanel = ui.createMainPanel()
	ui.outputPanel = ui.createOutputPanel()
	ui.summaryPanel = ui.createSummaryPanel()
	ui.keyItemsPanel = ui.createKeyItemsPanel()
	ui.itemSelectedHandler = ui.createKeySelectedHandler()
	ui.searchPanel = ui.createSearchPanel()
	ui.helpPanel = ui.createHelpPanel()

	ui.commandPanel = ui.createCommandPanel()

	ui.leftPanel = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.searchPanel, 3, 0, false).
		AddItem(ui.keyItemsPanel, 0, 1, false).
		AddItem(ui.summaryPanel, 3, 1, false)

	ui.rightPanel = tview.NewFlex().SetDirection(tview.FlexRow)
	ui.redrawRightPanel(ui.mainPanel)

	ui.app = tview.NewApplication()
	ui.layout = tview.NewFlex().
		AddItem(ui.leftPanel, 0, 3, false).
		AddItem(ui.rightPanel, 0, 8, false)

	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.searchPanel, Key: ui.keyBindings.KeyID("search")})
	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.keyItemsPanel, Key: ui.keyBindings.KeyID("keys")})

	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		if ui.config.Debug {
			keyName := event.Name()
			if event.Key() == tcell.KeyRune {
				keyName = string(event.Rune())
			}
			ui.outputChan <- core.OutputMessage{Message: fmt.Sprintf("Key %s pressed", keyName)}
		}

		name := ui.keyBindings.SearchKey(event.Key())
		switch name {
		case "switch_focus":
			nextFocusIndex := ui.currentFocusIndex + 1
			if nextFocusIndex > len(ui.focusPrimitives)-1 {
				nextFocusIndex = 0
			}

			ui.app.SetFocus(ui.focusPrimitives[nextFocusIndex].Primitive)
			ui.currentFocusIndex = nextFocusIndex

			return nil
		case "quit":
			ui.app.Stop()
		case "command":
			if ui.commandMode {
				ui.commandMode = false
				ui.redrawRightPanel(ui.mainPanel)
				ui.app.SetFocus(ui.searchPanel)
				ui.currentFocusIndex = 0
			} else {
				ui.commandMode = true
				ui.redrawRightPanel(ui.commandPanel)

				for i, p := range ui.focusPrimitives {
					if p.Primitive == ui.commandInputField {
						ui.app.SetFocus(ui.commandInputField)
						ui.currentFocusIndex = i
						break
					}
				}
			}
		default:
			for i, pv := range ui.focusPrimitives {
				if pv.Key == name {
					ui.app.SetFocus(pv.Primitive)
					ui.currentFocusIndex = i
					break
				}
			}
		}

		return event
	})

	return ui
}

// Start create the ui and start the program
func (ui *RedisTUI) Start() error {
	go func() {
		for {
			select {
			case f := <-ui.uiViewUpdateChan:
				(func() {
					defer func() {
						if err := recover(); err != nil {
						}
					}()
					ui.app.QueueUpdateDraw(f)
				})()
			}
		}
	}()
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case out := <-ui.outputChan:
				ui.uiViewUpdateChan <- func() {
					// clear outputPanel to avoid to many message caused ui hangup
					if ui.outputPanel.GetItemCount() > 20 {
						ui.outputPanel.Clear()
					}

					// ui.outputPanel.SetTextColor(out.Color).SetText(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), out.Message))
					ui.outputPanel.AddItem(fmt.Sprintf("%d | [%s] %s", ui.outputPanel.GetItemCount(), time.Now().Format(time.RFC3339), out.Message), "", 0, nil)
					ui.outputPanel.SetCurrentItem(-1)
				}
			case <-ticker.C:
				// update server info
				info, err := api.RedisServerInfo(ui.config, ui.redisClient)
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				}

				ui.uiViewUpdateChan <- func() {
					ui.helpServerInfoPanel.SetText(info)
				}
			}
		}
	}()

	go func() {
		info, err := api.RedisServerInfo(ui.config, ui.redisClient)
		if err != nil {
			ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
		}
		ui.app.QueueUpdateDraw(func() {
			ui.helpServerInfoPanel.SetText(info)
		})

		keys, err := api.RedisAllKeys(ui.redisClient, false)
		if err != nil {
			ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
			return
		}
		ui.app.QueueUpdateDraw(func() {
			ui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))

			for i, k := range limit(keys, ui.maxKeyLimit) {
				ui.keyItemsPanel.AddItem(ui.keyItemsFormat(i, k), "", 0, ui.itemSelectedHandler(i, k))
			}

			ui.app.SetFocus(ui.keyItemsPanel)
		})
	}()

	ui.pages = tview.NewPages()
	ui.pages.AddPage("base", ui.layout, true, true)
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
	// ui.pages.AddPage("welcome_screen", welcomeScreen, true, true)
	//
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	ui.app.QueueUpdateDraw(func() {
	// 		ui.pages.RemovePage("welcome_screen")
	// 	})
	// }()

	return ui.app.SetRoot(ui.pages, true).Run()
}

func (ui *RedisTUI) redrawRightPanel(center tview.Primitive) {
	ui.rightPanel.RemoveItem(ui.metaPanel).
		RemoveItem(ui.outputPanel).
		RemoveItem(ui.mainPanel).
		RemoveItem(ui.commandPanel).
		RemoveItem(ui.helpPanel)

	ui.rightPanel.AddItem(ui.helpPanel, 5, 1, false).
		AddItem(ui.metaPanel, 4, 1, false).
		AddItem(center, 0, 7, false).
		AddItem(ui.outputPanel, 8, 1, false)
}

func (ui *RedisTUI) createSummaryPanel() *tview.TextView {
	panel := tview.NewTextView()
	panel.SetBorder(true).SetTitle(" Info ")
	return panel
}

func (ui *RedisTUI) keyItemsFormat(index int, key string) string {
	return fmt.Sprintf("%3d | %s", index+1, key)
}

func (ui *RedisTUI) createCommandPanel() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	resultPanel := tview.NewTextView()
	resultPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Results (%s) ", ui.keyBindings.Name("command_result")))

	formPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	formPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Commands (%s) ", ui.keyBindings.Name("command_focus")))
	commandTipView := tview.NewTextView().SetDynamicColors(true).SetRegions(true)

	commandInputField := tview.NewInputField().SetLabel("Command ")

	var currentIndex = -1
	commandInputField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if len(ui.commandKeyHistories) == 0 {
				break
			}

			currentIndex = currentIndex - 1
			if currentIndex < 0 {
				currentIndex = len(ui.commandKeyHistories) - 1
			}

			commandInputField.SetText(ui.commandKeyHistories[currentIndex])
		case tcell.KeyDown:
			if len(ui.commandKeyHistories) == 0 {
				break
			}

			currentIndex = currentIndex + 1
			if currentIndex > len(ui.commandKeyHistories)-1 {
				currentIndex = 0
			}

			commandInputField.SetText(ui.commandKeyHistories[currentIndex])
		}

		return event
	})

	commandInputField.SetAutocompleteFunc(func(currentText string) (entries []string) {
		currentText = strings.ToLower(strings.TrimSpace(currentText))
		if currentText == "" || len(strings.Split(currentText, " ")) > 2 {
			return
		}

		cmds := api.RedisMatchedCommands(currentText)
		for _, c := range cmds {
			entries = append(entries, c.Command)
		}

		return entries
	})

	locked := make(chan interface{}, 1)
	locked <- struct{}{}
	commandInputField.SetDoneFunc(func(key tcell.Key) {
		select {
		case <-locked:

		default:
			pageID := "alert"
			ui.pages.AddPage(
				pageID,
				tview.NewModal().
					SetText("Other command is processing, please wait...").
					AddButtons([]string{"OK"}).
					SetDoneFunc(func(buttonIndex int, buttonLabel string) {
						ui.pages.HidePage(pageID).RemovePage(pageID)
						ui.app.SetFocus(commandInputField)
					}),
				false,
				true,
			)

			return
		}

		cmdText := commandInputField.GetText()
		ui.outputChan <- core.OutputMessage{Color: tcell.ColorOrange, Message: fmt.Sprintf("Command %s is processing...", cmdText)}

		go func(cmdText string) {
			defer func() {
				locked <- struct{}{}
			}()
			res, err := api.RedisExecute(ui.redisClient, cmdText)
			if err != nil {
				ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
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
			if len(output) > int(ui.maxCharacterLimit) {
				output = output[:ui.maxCharacterLimit] + fmt.Sprintf("\n\n ~ %d+ charactors omitted ~", len(output)-int(ui.maxCharacterLimit))
			}

			outputSlices := strings.Split(output, "\n")
			for i, v := range outputSlices {
				outputSlices[i] = fmt.Sprintf("%5d | %s", i+1, v)
			}

			output = strings.Join(outputSlices, "\n")

			ui.outputChan <- core.OutputMessage{Color: tcell.ColorGreen, Message: fmt.Sprintf("Command %s succeed", cmdText)}
			ui.uiViewUpdateChan <- func() {
				resultPanel.SetText(output)
			}
		}(cmdText)

		commandInputField.SetText("")
		currentIndex = 0
		if cmdText != "" {
			if len(ui.commandKeyHistories) > 0 {
				lastHis := ui.commandKeyHistories[len(ui.commandKeyHistories)-1]
				if lastHis != cmdText {
					ui.commandKeyHistories = append(ui.commandKeyHistories, cmdText)
				}
			} else {
				ui.commandKeyHistories = append(ui.commandKeyHistories, cmdText)
			}
		}
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

	ui.commandInputField = commandInputField
	ui.commandResultPanel = resultPanel

	formPanel.AddItem(commandInputField, 1, 1, false).
		AddItem(commandTipView, 4, 4, false)

	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.commandInputField, Key: ui.keyBindings.KeyID("command_focus")})
	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: ui.commandResultPanel, Key: ui.keyBindings.KeyID("command_result")})

	flex.AddItem(formPanel, 7, 0, false).
		AddItem(resultPanel, 0, 1, false)

	return flex
}

// createSearchPanel create search panel
func (ui *RedisTUI) createSearchPanel() *tview.InputField {
	searchArea := tview.NewInputField().SetLabel(" Key ")
	var currentIndex = -1
	searchArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if len(ui.searchKeyHistories) == 0 {
				break
			}

			currentIndex = currentIndex - 1
			if currentIndex < 0 {
				currentIndex = len(ui.searchKeyHistories) - 1
			}

			searchArea.SetText(ui.searchKeyHistories[currentIndex])
		case tcell.KeyDown:
			if len(ui.searchKeyHistories) == 0 {
				break
			}

			currentIndex = currentIndex + 1
			if currentIndex > len(ui.searchKeyHistories)-1 {
				currentIndex = 0
			}

			searchArea.SetText(ui.searchKeyHistories[currentIndex])
		}

		return event
	})
	searchArea.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		var text = searchArea.GetText()

		ui.keyItemsPanel.Clear()
		searchArea.SetText("")

		currentIndex = 0
		if text != "" {
			if len(ui.searchKeyHistories) > 0 {
				lastHis := ui.searchKeyHistories[len(ui.searchKeyHistories)-1]
				if lastHis != text {
					ui.searchKeyHistories = append(ui.searchKeyHistories, text)
				}
			} else {
				ui.searchKeyHistories = append(ui.searchKeyHistories, text)
			}
		}

		var keys []string
		var err error

		// if text == "" || text == "*" {
		// 	keys, _, err = ui.redisClient.Scan(0, text, ui.maxKeyLimit).Result()
		// } else {
		// 	keys, err = ui.redisClient.Keys(text).Result()
		// }
		keys, err = api.RedisKeys(ui.redisClient, text)

		if err != nil {
			ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
			return
		}

		ui.summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))
		for i, k := range keys {
			if i > int(ui.maxKeyLimit) {
				break
			}
			ui.keyItemsPanel.AddItem(ui.keyItemsFormat(i, k), "", 0, ui.itemSelectedHandler(i, k))
		}
	})
	searchArea.SetAutocompleteFunc(func(currentText string) (entries []string) {
		currentText = strings.TrimSpace(currentText)
		if len(currentText) == 0 {
			return
		}

		keys, err := api.KeysWithLimit(ui.redisClient, currentText+"*", 10)
		if err != nil {
			ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
			return
		}

		matchedCount := 0
		for _, k := range keys {
			if matchedCount >= 10 {
				break
			}

			if strings.HasPrefix(k, currentText) {
				entries = append(entries, k)
				matchedCount++
			}
		}

		return
	})
	searchArea.SetBorder(true).SetTitle(fmt.Sprintf(" Search (%s) ", ui.keyBindings.Name("search")))
	return searchArea
}

// createKeyItemsPanel create key items panel
func (ui *RedisTUI) createKeyItemsPanel() *tview.List {
	keyItemsList := tview.NewList().ShowSecondaryText(false)
	keyItemsList.SetBorder(true).SetTitle(fmt.Sprintf(" Keys (%s) ", ui.keyBindings.Name("keys")))
	return keyItemsList
}

// primitivesFilter is a filter for primitives
func (ui *RedisTUI) primitivesFilter(items []primitiveKey, filter func(item primitiveKey) bool) []primitiveKey {
	res := make([]primitiveKey, 0)
	for _, item := range items {
		if filter(item) {
			res = append(res, item)
		}
	}

	return res
}

// createMetaPanel create a panel for meta info
func (ui *RedisTUI) createMetaPanel() *tview.TextView {
	metaInfoArea := tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	metaInfoArea.SetBorder(true).SetTitle(" Meta ")

	return metaInfoArea
}

// createMainPanel create main panel
func (ui *RedisTUI) createMainPanel() *tview.Flex {
	mainArea := tview.NewFlex()
	mainArea.SetBorder(true).SetTitle(" Value ")

	mainArea.AddItem(ui.welcomeScreen, 0, 1, false)

	return mainArea
}

// createOutputPanel create a panel for outputFunc
func (ui *RedisTUI) createOutputPanel() *tview.List {
	outputArea := tview.NewList().ShowSecondaryText(false)
	outputArea.SetBorder(true).SetTitle(fmt.Sprintf(" Output (%s) ", ui.keyBindings.Name("output")))

	ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: outputArea, Key: ui.keyBindings.KeyID("output")})

	return outputArea
}

// createHelpPanel create a panel for help message display
func (ui *RedisTUI) createHelpPanel() *tview.Flex {
	helpPanel := tview.NewFlex().SetDirection(tview.FlexRow)
	helpPanel.SetBorder(true).SetTitle(fmt.Sprintf(" Version: %s (%s) ", ui.version, ui.gitCommit))

	ui.helpServerInfoPanel = tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	helpPanel.AddItem(ui.helpServerInfoPanel, 2, 1, false)

	ui.helpMessagePanel = tview.NewTextView()
	ui.helpMessagePanel.SetTextColor(tcell.ColorOrange).SetText(fmt.Sprintf(
		" ❈ %s - open command panel, %s - switch focus, %s - quit",
		ui.keyBindings.Name("command"),
		ui.keyBindings.Name("switch_focus"),
		ui.keyBindings.Name("quit"),
	))

	helpPanel.AddItem(ui.helpMessagePanel, 1, 1, false)

	return helpPanel
}

// createKeySelectedHandler create a handler for item selected event
func (ui *RedisTUI) createKeySelectedHandler() func(index int, key string) func() {

	// 用于KV展示的视图
	mainStringView := tview.NewTextView()
	mainStringView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", ui.keyBindings.Name("key_string_value")))

	mainHashView := tview.NewList().ShowSecondaryText(false)
	mainHashView.SetBorder(true).SetTitle(fmt.Sprintf(" Hash KeyID (%s) ", ui.keyBindings.Name("key_hash")))

	mainListView := tview.NewList().ShowSecondaryText(false).SetSecondaryTextColor(tcell.ColorOrangeRed)
	mainListView.SetBorder(true).SetTitle(fmt.Sprintf(" Value (%s) ", ui.keyBindings.Name("key_list_value")))

	return func(index int, key string) func() {
		return func() {
			keyType, err := ui.redisClient.Type(key).Result()
			if err != nil {
				ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				return
			}

			ttl, err := ui.redisClient.TTL(key).Result()
			if err != nil {
				ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
				return
			}

			// 移除主区域的边框，因为展示区域已经带有边框了
			ui.mainPanel.RemoveItem(ui.welcomeScreen).SetBorder(false)

			// 重置展示视图
			mainHashView.Clear()
			mainStringView.Clear()
			mainListView.Clear().ShowSecondaryText(false)

			ui.focusPrimitives = ui.primitivesFilter(ui.focusPrimitives, func(item primitiveKey) bool {
				return item.Primitive != mainHashView && item.Primitive != mainListView && item.Primitive != mainStringView
			})

			ui.mainPanel.RemoveItem(mainStringView)
			ui.mainPanel.RemoveItem(mainHashView)
			ui.mainPanel.RemoveItem(mainListView)

			// 根据不同的kv类型，展示不同的视图
			switch keyType {
			case "string":
				result, err := ui.redisClient.Get(key).Result()
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				ui.mainPanel.AddItem(mainStringView.SetText(fmt.Sprintf(" %s", result)), 0, 1, false)
				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: ui.keyBindings.KeyID("key_string_value")})
			case "list":
				values, err := ui.redisClient.LRange(key, 0, 1000).Result()
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				ui.mainPanel.AddItem(mainListView, 0, 1, false)
				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: ui.keyBindings.KeyID("key_list_value")})

			case "set":
				values, err := ui.redisClient.SMembers(key).Result()
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %3d | %s", i+1, v), "", 0, nil)
				}

				ui.mainPanel.AddItem(mainListView, 0, 1, false)
				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: ui.keyBindings.KeyID("key_list_value")})

			case "zset":
				values, err := ui.redisClient.ZRangeWithScores(key, 0, 1000).Result()
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				mainListView.ShowSecondaryText(true)
				for i, z := range values {
					val := fmt.Sprintf(" %3d | %v", i+1, z.Member)
					score := fmt.Sprintf("    Score: %v", z.Score)

					mainListView.AddItem(val, score, 0, nil)
				}

				ui.mainPanel.AddItem(mainListView, 0, 1, false)
				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainListView, Key: ui.keyBindings.KeyID("key_list_value")})

			case "hash":
				hashKeys, err := ui.redisClient.HKeys(key).Result()
				if err != nil {
					ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
					return
				}

				for i, k := range hashKeys {
					mainHashView.AddItem(fmt.Sprintf(" %3d | %s", i+1, k), "", 0, (func(k string) func() {
						return func() {
							val, err := ui.redisClient.HGet(key, k).Result()
							if err != nil {
								ui.outputChan <- core.OutputMessage{Color: tcell.ColorRed, Message: fmt.Sprintf("errors: %s", err)}
								return
							}

							mainStringView.SetText(fmt.Sprintf(" %s", val)).SetTitle(fmt.Sprintf(" Value: %s ", k))
						}
					})(k))
				}

				ui.mainPanel.AddItem(mainHashView, 0, 3, false).
					AddItem(mainStringView, 0, 7, false)

				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainHashView, Key: ui.keyBindings.KeyID("key_hash")})
				ui.focusPrimitives = append(ui.focusPrimitives, primitiveKey{Primitive: mainStringView, Key: ui.keyBindings.KeyID("key_string_value")})
			}
			ui.outputChan <- core.OutputMessage{Color: tcell.ColorGreen, Message: fmt.Sprintf("query %s OK, type=%s, ttl=%s", key, keyType, ttl.String())}
			ui.metaPanel.SetText(fmt.Sprintf("KeyID: %s\nType: %s, TTL: %s", key, keyType, ttl.String())).SetTextAlign(tview.AlignCenter)
		}
	}
}

func limit(input []string, maxReturn int) []string {
	if len(input) <= maxReturn {
		return input
	}

	return input[:maxReturn]
}
