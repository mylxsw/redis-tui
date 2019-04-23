package core

import (
	"github.com/gdamore/tcell"
	"strings"
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

func NewKeyBinding() KeyBindings {
	return keyBindings
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

type OutputMessage struct {
	Color   tcell.Color
	Message string
}