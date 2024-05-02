package view

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	// column height in info page
	oneColumnCount = 11
	headerTitleFmt = " [blue]info([purple::b]%s[blue:-:-]) "
	headerKeyFmt   = " [purple::b]<%s> [green:-:-]%s "
	headerItemFmt  = " %s:[aqua::b] %s "
)

var hotKeyMap = map[string]keyDescriptionPair{
	"a":     {key: "a", description: "Describe service auto scaling"},
	"f":     {key: "f", description: "Toggle full screen"},
	"l":     {key: "l", description: "Show cloudwatch logs(Only support awslogs logDriver)"},
	"m":     {key: "m", description: "Show metrics(CPU/Memory)"},
	"r":     {key: "r", description: "Realtime log streaming(Only support one log group)"},
	"t":     {key: "t", description: "Show task definitions"},
	"n":     {key: "n", description: "Show all cluster tasks"},
	"s":     {key: "s", description: "Toggle running/stopped tasks"},
	"w":     {key: "w", description: "Describe service events"},
	"P":     {key: "shift-p", description: "File transfer"},
	"F":     {key: "shift-f", description: "Start port forwarding session"},
	"T":     {key: "shift-t", description: "Terminate port forwarding session"},
	"U":     {key: "shift-u", description: "Update service"},
	"E":     {key: "shift-e", description: "Exec command"},
	"ctrlD": {key: "ctrl-d", description: "Exit from container"},

	"enter": {key: "enter", description: "Select"},
	"esc":   {key: "esc", description: "Back"},
	"ctrlZ": {key: "ctrl-z", description: "Back"},
	"ctrlC": {key: "ctrl-c", description: "Exit"},
	"ctrlR": {key: "ctrl-r", description: "Refresh"},
	"?":     {key: "?", description: "Help"},
	"b":     {key: "b", description: "Open in browser"},
	"d":     {key: "d", description: "Describe"},
	"e":     {key: "e", description: "Open in default editor"},

	"j":     {key: "j", description: "Down"},
	"k":     {key: "k", description: "Up"},
	"G":     {key: "shift-g", description: "Go to bottom"},
	"g":     {key: "g", description: "Go to top"},
	"ctrlF": {key: "ctrl+f", description: "Page down"},
	"ctrlB": {key: "ctrl+b", description: "Page up"},
}

// Info item name and value
type headerItem struct {
	name  string
	value string
}

// Keyboard key and description
type keyDescriptionPair struct {
	key         string
	description string
}

var basicKeyInputs = []keyDescriptionPair{
	hotKeyMap["b"],
	hotKeyMap["d"],
	hotKeyMap["ctrlR"],
}

type secondaryPageKeyMap = map[kind][]keyDescriptionPair

var describePageKeys = []keyDescriptionPair{
	hotKeyMap["f"],
	hotKeyMap["b"],
	hotKeyMap["e"],
	hotKeyMap["ctrlZ"],
}

var otherDescribePageKeys = []keyDescriptionPair{
	hotKeyMap["f"],
	hotKeyMap["b"],
	hotKeyMap["ctrlZ"],
}

var logPageKeys = []keyDescriptionPair{
	hotKeyMap["f"],
	hotKeyMap["b"],
	hotKeyMap["r"],
	hotKeyMap["ctrlR"],
	hotKeyMap["ctrlZ"],
}

// Build header flex show on top of view, will change when selection change
func (v *view) buildHeaderFlex(title string, items []headerItem, keys []keyDescriptionPair) *tview.Flex {
	headerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	columnCount := len(items)/oneColumnCount + 1
	var columns []*tview.Flex
	for i := 0; i < columnCount; i++ {
		columns = append(columns, tview.NewFlex().SetDirection(tview.FlexRow))
	}
	count := 0
	columnCount = 0
	for _, item := range items {
		count++
		if count == oneColumnCount {
			count = 0
			columnCount++
		}

		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(headerItemFmt, item.name, item.value))
		columns[columnCount].AddItem(t, 1, 1, false)
	}
	keysColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, k := range keys {
		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(headerKeyFmt, k.key, k.description))
		keysColumn.AddItem(t, 1, 1, false)
	}
	columns = append(columns, keysColumn)

	if len(columns) == 2 {
		columns = append(columns, tview.NewFlex().SetDirection(tview.FlexRow))
	}

	for _, c := range columns {
		headerFlex.AddItem(c, 0, 1, false).SetTitle(fmt.Sprintf(headerTitleFmt, title))
		headerFlex.SetBorder(true)
	}

	return headerFlex
}
