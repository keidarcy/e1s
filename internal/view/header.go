package view

import (
	"fmt"

	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

const (
	// column height in info page
	oneColumnCount = 13
	headerTitleFmt = " [blue]info([purple::b]%s[blue:-:-]) "
	headerItemFmt  = " %s:[aqua::b] %s "
	headerKeyFmt   = " [purple::b]<%s> [green:-:-]%s "
)

var hotKeyMap = map[string]keyDescriptionPair{
	"/":     {key: "/", description: "Search in table"},
	"a":     {key: "a", description: "Show service auto scaling"},
	"f":     {key: "f", description: "Toggle full screen"},
	"l":     {key: "l", description: "Show cloudwatch logs(Only support awslogs logDriver)"},
	"m":     {key: "m", description: "Show metrics(CPU/Memory)"},
	"r":     {key: "r", description: "Realtime log streaming(Only support one log group)"},
	"R":     {key: "shift-r", description: "Rollback service deployment"},
	"t":     {key: "t", description: "Show task definitions"},
	"p":     {key: "p", description: "Show service deployments"},
	"n":     {key: "n", description: "Show related EC2 instances"},
	"N":     {key: "shift-n", description: "Show all cluster tasks"},
	"s":     {key: "s", description: "Shell access"},
	"x":     {key: "x", description: "Toggle running/stopped tasks"},
	"w":     {key: "w", description: "Show service events"},
	"v":     {key: "v", description: "Show service revision"},
	"S":     {key: "shift-s", description: "Stop task"},
	"P":     {key: "shift-p", description: "Transfer file though a S3 bucket"},
	"D":     {key: "shift-d", description: "Download text file content(beta)"},
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

	"j":       {key: "j, down arrow", description: "Down"},
	"k":       {key: "k, up arrow", description: "Up"},
	"G":       {key: "shift-g, end", description: "Go to bottom"},
	"g":       {key: "g, home", description: "Go to top"},
	"ctrlF":   {key: "ctrl+f, page down", description: "Page down"},
	"ctrlB":   {key: "ctrl+b, page up", description: "Page up"},
	"tab":     {key: "tab", description: "Field next"},
	"backtab": {key: "backtab", description: "Field previous"},
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
	hotKeyMap["/"],
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
	hotKeyMap["e"],
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

		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.HeaderItemFmt, item.name, item.value))
		columns[columnCount].AddItem(t, 1, 1, false)
	}
	keysColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	for _, k := range keys {
		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(color.HeaderKeyFmt, k.key, k.description))
		keysColumn.AddItem(t, 1, 1, false)
	}
	columns = append(columns, keysColumn)

	if len(columns) == 2 {
		columns = append(columns, tview.NewFlex().SetDirection(tview.FlexRow))
	}

	for _, c := range columns {
		headerFlex.AddItem(c, 0, 1, false).SetTitle(fmt.Sprintf(color.HeaderTitleFmt, title))
		headerFlex.SetBorder(true)
	}

	return headerFlex
}
