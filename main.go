package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type primitiveKey struct {
	Primitive tview.Primitive
	Key       tcell.Key
}

var focusPrimitives = make([]primitiveKey, 0)
var currentFocusIndex = 0

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
	Cluster  bool
}

var config = Config{}

var Version string
var GitCommit string

type OutputFunc func(color tcell.Color, message string)

func main() {

	flag.StringVar(&config.Host, "h", "127.0.0.1", "Server hostname")
	flag.IntVar(&config.Port, "p", 6379, "Server port")
	flag.StringVar(&config.Password, "a", "", "Password to use when connecting to the server")
	flag.IntVar(&config.DB, "n", 0, "Database number")
	flag.BoolVar(&config.Cluster, "c", false, "Enable cluster mode")

	flag.Parse()

	client := NewRedisClient(config)

	metaPanel := createMetaPanel()
	mainPanel := createMainPanel()
	outputPanel := createOutputPanel()

	outputFunc := createOutputFunc(outputPanel)
	keyItemsPanel := createKeyItemsPanel()

	itemSelectedHandler := createKeySelectedHandler(client, outputFunc, mainPanel, metaPanel)

	summaryPanel := createSummaryPanel()
	searchPanel := createSearchPanel(client, outputFunc, keyItemsPanel, summaryPanel, itemSelectedHandler)

	leftPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchPanel, 3, 0, false).
		AddItem(keyItemsPanel, 0, 1, false).
		AddItem(summaryPanel, 3, 1, false)

	rightPanel := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(metaPanel, 4, 1, false).
		AddItem(mainPanel, 0, 7, false).
		AddItem(outputPanel, 5, 1, false)

	app := tview.NewApplication()
	flex := tview.NewFlex().
		AddItem(leftPanel, 0, 3, false).
		AddItem(rightPanel, 0, 8, false)

	focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: searchPanel, Key: tcell.KeyF1})
	focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: keyItemsPanel, Key: tcell.KeyF2})

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			nextFocusIndex := currentFocusIndex + 1
			if nextFocusIndex > len(focusPrimitives)-1 {
				nextFocusIndex = 0
			}

			app.SetFocus(focusPrimitives[nextFocusIndex].Primitive)
			currentFocusIndex = nextFocusIndex
		case tcell.KeyEsc:
			app.Stop()
		default:
			for i, pv := range focusPrimitives {
				if pv.Key == event.Key() {
					app.SetFocus(pv.Primitive)
					currentFocusIndex = i
					break
				}
			}
		}

		return event
	})

	go app.QueueUpdateDraw(func() {
		keys, err := client.Keys("*").Result()
		if err != nil {
			outputPanel.SetTextColor(tcell.ColorRed).SetText(fmt.Sprintf("Errors: %s", err))
			return
		}

		summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))

		for _, k := range keys {
			keyItemsPanel.AddItem(fmt.Sprintf(" %s", k), "", 0, itemSelectedHandler(k))
		}

		app.SetFocus(keyItemsPanel)
	})

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func createSummaryPanel() *tview.TextView {
	panel := tview.NewTextView()
	panel.SetBorder(true).SetTitle("Info")
	return panel
}

// createSearchPanel create search panel
func createSearchPanel(client RedisClient, outputFunc OutputFunc, keyItemsPanel *tview.List, summaryPanel *tview.TextView, itemSelectedHandler func(key string) func()) *tview.InputField {
	searchArea := tview.NewInputField().SetLabel(" Key ").SetChangedFunc(func(text string) {
		keys, err := client.Keys(text).Result()
		if err != nil {
			outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
			return
		}

		keyItemsPanel.Clear()

		summaryPanel.SetText(fmt.Sprintf(" Total matched: %d", len(keys)))
		for _, k := range keys {
			keyItemsPanel.AddItem(fmt.Sprintf(" %s", k), "", 0, itemSelectedHandler(k))
		}
	})
	searchArea.SetBorder(true).SetTitle(" Search (F1) ")
	return searchArea
}

// createKeyItemsPanel create key items panel
func createKeyItemsPanel() *tview.List {
	keyItemsList := tview.NewList().ShowSecondaryText(false)
	keyItemsList.SetBorder(true).SetTitle(" Keys (F2) ")
	return keyItemsList
}

// primitivesFilter is a filter for primitives
func primitivesFilter(items []primitiveKey, filter func(item primitiveKey) bool) []primitiveKey {
	res := make([]primitiveKey, 0)
	for _, item := range items {
		if filter(item) {
			res = append(res, item)
		}
	}

	return res
}

// createMetaPanel create a panel for meta info
func createMetaPanel() *tview.TextView {
	metaInfoArea := tview.NewTextView().SetDynamicColors(true).SetRegions(true)
	metaInfoArea.SetBorder(true).SetTitle(fmt.Sprintf(" Version: %s (%s) ", Version, GitCommit[0:8]))

	return metaInfoArea
}

// createMainPanel create main panel
func createMainPanel() *tview.Flex {
	mainArea := tview.NewFlex()
	mainArea.SetBorder(true).SetTitle(" Value ")

	return mainArea
}

// createOutputPanel create a panel for output
func createOutputPanel() *tview.TextView {
	outputArea := tview.NewTextView()
	outputArea.SetBorder(true).SetTitle(" Output ")

	return outputArea
}

// createOutputFunc create a output func
func createOutputFunc(view *tview.TextView) OutputFunc {
	return func(color tcell.Color, message string) {
		view.SetTextColor(color).SetText(fmt.Sprintf("[%s] %s", time.Now().Format(time.RFC3339), message))
	}
}

// createKeySelectedHandler create a handler for item selected event
func createKeySelectedHandler(client RedisClient, outputFunc OutputFunc, mainPanel *tview.Flex, metaPanel *tview.TextView) func(key string) func() {

	// 用于KV展示的视图
	mainStringView := tview.NewTextView()
	mainStringView.SetBorder(true).SetTitle(" Value ")

	mainHashView := tview.NewList().ShowSecondaryText(false)
	mainHashView.SetBorder(true).SetTitle(" Hash Key (F3) ")

	mainListView := tview.NewList().ShowSecondaryText(false).SetSecondaryTextColor(tcell.ColorOrangeRed)
	mainListView.SetBorder(true).SetTitle(" Value (F3) ")

	return func(key string) func() {
		return func() {
			keyType, err := client.Type(key).Result()
			if err != nil {
				outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
				return
			}

			ttl, err := client.TTL(key).Result()
			if err != nil {
				outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
				return
			}

			// 移除主区域的边框，因为展示区域已经带有边框了
			mainPanel.SetBorder(false)

			// 重置展示视图
			mainHashView.Clear()
			mainStringView.Clear()
			mainListView.Clear().ShowSecondaryText(false)

			focusPrimitives = primitivesFilter(focusPrimitives, func(item primitiveKey) bool {
				return item.Primitive != mainHashView && item.Primitive != mainListView
			})

			mainPanel.RemoveItem(mainStringView)
			mainPanel.RemoveItem(mainHashView)
			mainPanel.RemoveItem(mainListView)

			// 根据不同的kv类型，展示不同的视图
			switch keyType {
			case "string":
				result, err := client.Get(key).Result()
				if err != nil {
					outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
					return
				}

				mainPanel.AddItem(mainStringView.SetText(fmt.Sprintf(" %s", result)), 0, 1, false)
			case "list":
				values, err := client.LRange(key, 0, 1000).Result()
				if err != nil {
					outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %d. %s", i+1, v), "", 0, nil)
				}

				mainPanel.AddItem(mainListView, 0, 1, false)
				focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: mainListView, Key: tcell.KeyF3})

			case "set":
				values, err := client.SMembers(key).Result()
				if err != nil {
					outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
					return
				}

				for i, v := range values {
					mainListView.AddItem(fmt.Sprintf(" %d. %s", i+1, v), "", 0, nil)
				}

				mainPanel.AddItem(mainListView, 0, 1, false)
				focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: mainListView, Key: tcell.KeyF3})

			case "zset":
				values, err := client.ZRangeWithScores(key, 0, 1000).Result()
				if err != nil {
					outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
					return
				}

				mainListView.ShowSecondaryText(true)
				for i, z := range values {
					val := fmt.Sprintf(" %d. %v", i, z.Member)
					score := fmt.Sprintf("    Score: %v", z.Score)

					mainListView.AddItem(val, score, 0, nil)
				}

				mainPanel.AddItem(mainListView, 0, 1, false)
				focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: mainListView, Key: tcell.KeyF3})

			case "hash":
				hashKeys, err := client.HKeys(key).Result()
				if err != nil {
					outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
					return
				}

				for _, k := range hashKeys {
					mainHashView.AddItem(fmt.Sprintf(" %s", k), "", 0, (func(k string) func() {
						return func() {
							val, err := client.HGet(key, k).Result()
							if err != nil {
								outputFunc(tcell.ColorRed, fmt.Sprintf("Errors: %s", err))
								return
							}

							mainStringView.SetText(fmt.Sprintf(" %s", val)).SetTitle(fmt.Sprintf(" Value: %s ", k))
						}
					})(k))
				}

				mainPanel.AddItem(mainHashView, 0, 3, false).
					AddItem(mainStringView, 0, 7, false)

				focusPrimitives = append(focusPrimitives, primitiveKey{Primitive: mainHashView, Key: tcell.KeyF3})
			}
			outputFunc(tcell.ColorGreen, fmt.Sprintf("Query %s OK, Type=%s, TTL=%s", key, keyType, ttl.String()))
			metaPanel.SetText(fmt.Sprintf("Key: %s\nType: %s, TTL: %s", key, keyType, ttl.String())).SetTextAlign(tview.AlignCenter)
		}
	}
}

