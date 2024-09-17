package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

func (v *view) searchForm() (*tview.Form, *string) {
	title := " Search in table "

	f := ui.StyledForm(title)
	searchLabel := "Input search"

	inputField := tview.NewInputField().
		SetLabel(searchLabel).
		SetFieldWidth(50).
		SetText(*v.searchLast).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				if submitForm(searchLabel, f, v) {
					v.closeModal()
					v.app.QueueEvent(tcell.NewEventKey(tcell.KeyEsc, rune(tcell.KeyEsc), tcell.ModNone))
				}
			}
		})
	f.AddFormItem(inputField)

	return f, &title
}

func submitForm(searchLabel string, f *tview.Form, v *view) bool {
	searchInput := f.GetFormItemByLabel(searchLabel).(*tview.InputField).GetText()
	v.searchLast = &searchInput

	if searchInput == "" {
		return false
	}

	table := v.table
	found := false
	selectedRow, _ := table.GetSelection()
	rowCount := table.GetRowCount()
	wrapSearch := false

RowLoop:
	for i := 0; i < rowCount; i++ {
		currentRow := (selectedRow + i) % rowCount
		if currentRow < selectedRow {
			wrapSearch = true
		}

		for j := 0; j < table.GetColumnCount(); j++ {
			cell := table.GetCell(currentRow, j)
			if strings.Contains(cell.Text, searchInput) {
				found = true
				if selectedRow == currentRow {
					continue RowLoop
				}
				table.Select(currentRow, 0)
				break RowLoop
			}
		}
	}

	if !found {
		v.app.Notice.Warnf("%s not found", searchInput)
	} else if wrapSearch {
		v.app.Notice.Info("search hit BOTTOM, continuing at TOP")
	}
	return true
}
