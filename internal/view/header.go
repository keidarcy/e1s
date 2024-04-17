package view

import (
	"fmt"

	"github.com/rivo/tview"
)

const (
	// column height in info page
	oneColumnCount = 11
)

// Info item name and value
type headerItem struct {
	name  string
	value string
}

// Keyboard key and description
type keyInput struct {
	key         string
	description string
}

var basicKeyInputs = []keyInput{
	{key: string(bKey), description: openInBrowser},
	{key: string(dKey), description: describe},
	{key: ctrlR, description: reloadResource},
}

type secondaryPageKeyMap = map[kind][]keyInput

var describePageKeys = []keyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: string(eKey), description: openInEditor},
	{key: ctrlZ, description: backToPrevious},
}

var otherDescribePageKeys = []keyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: ctrlZ, description: backToPrevious},
}

var logPageKeys = []keyInput{
	{key: string(fKey), description: toggleFullScreen},
	{key: string(bKey), description: openInBrowser},
	{key: string(rKey), description: realtimeLog},
	{key: ctrlR, description: reloadResource},
	{key: ctrlZ, description: backToPrevious},
}

// Build info flex show on top of view, will change when selection change
func (v *view) buildHeaderFlex(title string, items []headerItem, keys []keyInput) *tview.Flex {
	infoFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

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
		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(keyFmt, k.key, k.description))
		keysColumn.AddItem(t, 1, 1, false)
	}
	columns = append(columns, keysColumn)

	if len(columns) == 2 {
		columns = append(columns, tview.NewFlex().SetDirection(tview.FlexRow))
	}

	for _, c := range columns {
		infoFlex.AddItem(c, 0, 1, false).SetTitle(fmt.Sprintf(headerTitleFmt, title))
		infoFlex.SetBorder(true)
	}

	return infoFlex
}
