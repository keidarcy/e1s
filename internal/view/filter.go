package view

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

// Initialize filter input component
func (v *view) initFilterInput() {
	v.filterInput = tview.NewInputField().
		SetLabel("[gray]🔍 :[-] ").
		SetLabelColor(color.Color(theme.Cyan))
	v.filterInput.SetBackgroundColor(color.Color(theme.BgColor))
	v.filterInput.SetFieldBackgroundColor(color.Color(theme.BgColor))
	v.filterInput.SetFieldTextColor(color.Color(theme.FgColor))
	v.filterInput.SetBorderPadding(0, 0, 1, 0)
	v.filterInput.SetBorder(false)

	v.filterInput.SetFocusFunc(func() {
		v.filterInput.SetBorder(true).SetBorderColor(color.Color(theme.Blue))
	})
	v.filterInput.SetBlurFunc(func() {
		v.filterInput.SetBorder(false)
	})

	// Auto-apply filter after 1 second of no typing
	v.filterInput.SetChangedFunc(func(text string) {
		if v.filterApplyTimer != nil {
			v.filterApplyTimer.Stop()
		}
		v.filterApplyTimer = time.AfterFunc(1*time.Second, func() {
			v.app.QueueUpdateDraw(func() {
				if v.filterActive {
					v.applyFilter()
				}
			})
		})

		v.updateFilterTitle()
	})

	v.filterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			v.hideFilterInput()
			v.applyFilter()
		case tcell.KeyEsc:
			v.filterInput.SetText("")
			v.hideFilterInput()
			v.applyFilter()
		}
	})
}

func (v *view) updateFilterTitle() {
	if v.filterInput == nil {
		return
	}
	filterText := v.filterInput.GetText()
	count := v.table.GetRowCount() - 1 // -1 for headers
	currentTitle := fmt.Sprintf(color.TableTitleFmt, v.app.kind, "all", count)
	if len(filterText) > 0 {
		filterLabel := fmt.Sprintf("[black:blue]</%s>[-:-]", filterText)
		currentTitle = currentTitle + " " + filterLabel
	}
	v.table.SetTitle(currentTitle)
}

func (v *view) applyFilter() {
	if v.filterInput == nil {
		return
	}
	filterText := v.filterInput.GetText()
	filteredData := [][]string{}
	filteredReferences := []Entity{}
	for i, row := range v.originalRowData {
		if v.shouldShow(row, filterText) {
			filteredData = append(filteredData, row)
			filteredReferences = append(filteredReferences, v.originalRowReferences[i])
		}
	}
	v.table.Clear()
	v.buildTableContent(filteredData, filteredReferences)
	v.updateFilterTitle()
	if len(filteredData) > 0 {
		v.table.Select(1, 0)
	}
	slog.Info("apply filter", "filterText", filterText, "filteredRowCount", len(filteredData))
}

func (v *view) shouldShow(row []string, filterText string) bool {
	parts := strings.Split(filterText, ":")
	// if no colon, only match first column
	if len(parts) != 2 {
		return strings.Contains(strings.ToLower(row[0]), strings.ToLower(filterText))
	}
	headerKey := strings.Trim(parts[0], " ")
	value := strings.Trim(parts[1], " ")
	for i, header := range v.headers {
		if strings.ToLower(header) == strings.ToLower(headerKey) {
			return strings.Contains(strings.ToLower(row[i]), strings.ToLower(value))
		}
	}
	return false
}

func (v *view) showFilterInput() error {
	if v.filterActive {
		return nil
	}
	v.filterActive = true
	if v.mainFlex != nil {
		v.mainFlex.RemoveItem(v.bodyPages)
		v.mainFlex.RemoveItem(v.footer.footerFlex)
		v.mainFlex.AddItem(v.filterInput, 3, 0, false)
		v.mainFlex.AddItem(v.bodyPages, 0, 2, true)
		v.mainFlex.AddItem(v.footer.footerFlex, 1, 1, false)
	}
	v.app.SetFocus(v.filterInput)
	return nil
}

func (v *view) hideFilterInput() {
	if !v.filterActive {
		return
	}
	v.filterActive = false
	if v.filterApplyTimer != nil {
		v.filterApplyTimer.Stop()
		v.filterApplyTimer = nil
	}
	if v.mainFlex != nil {
		v.mainFlex.RemoveItem(v.filterInput)
	}
	v.app.SetFocus(v.table)
}
