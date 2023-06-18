package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

// Build info flex show on top of view, will change when selection change
func (v *View) buildInfoFlex(title string, items []InfoItem, keys []KeyInput) *tview.Flex {
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

		t := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(infoItemFmt, item.name, item.value))
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
		infoFlex.AddItem(c, 0, 1, false).SetTitle(fmt.Sprintf(infoTitleFmt, title))
		infoFlex.SetBorder(true)
	}

	return infoFlex
}
