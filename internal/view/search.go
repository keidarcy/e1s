package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

func (v *view) searchForm() (*tview.Form, *string) {
	title := " Search in table"

	f := ui.StyledForm(title)
	searchLabel := "Input search (first column)"

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

	// handle form close
	f.AddButton("Cancel", func() {
		v.searchLast = new(string)
		v.closeModal()
	})

	// handle form submit
	f.AddButton("Search", func() {
		if submitForm(searchLabel, f, v) {
			v.closeModal()
		}
	})
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

	for i := 0; i < rowCount; i++ {
		currentRow := (selectedRow + i) % rowCount
		if currentRow < selectedRow {
			wrapSearch = true
		}
		cell := table.GetCell(currentRow, 0)
		if strings.Contains(cell.Text, searchInput) {
			found = true
			if selectedRow == currentRow {
				continue
			}
			table.Select(currentRow, 0)
			break
		}
	}
	if !found {
		v.app.Notice.Warnf("prefix %s not found", searchInput)
	} else if wrapSearch {
		v.app.Notice.Info("search hit BOTTOM, continuing at TOP")
	}
	return true
}
