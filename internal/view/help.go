package view

import (
	"fmt"

	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

type helpView struct {
	view
}

func newHelpView(app *App) *helpView {
	keys := append(basicKeyInputs, []keyInput{}...)
	return &helpView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
	}
}

var keys = []keyInput{
	{key: string("shift-u"), description: updateService},
	{key: string(wKey), description: describeServiceEvents},
	{key: string(tKey), description: showTaskDefinitions},
	{key: string(mKey), description: showMetrics},
	{key: string(aKey), description: describeAutoScaling},
	{key: string(lKey), description: showLogs},
}

func (app *App) showHelpPage() {
	view := newHelpView(app)
	t1 := tview.NewTable()
	resource := tview.NewTableCell("[aqua::b]Resource").SetAlign(L)
	t1.SetCell(0, 0, resource)
	for i, k := range keys {
		key := tview.NewTableCell(fmt.Sprintf("[purple]<%s>", k.key)).SetAlign(L).SetExpansion(7)
		description := tview.NewTableCell(fmt.Sprintf("[green]%s", k.description)).SetAlign(L)
		t1.SetCell(i+1, 0, key)
		t1.SetCell(i+1, 1, description)
	}
	flex := tview.NewFlex().
		AddItem(t1, 0, 1, false).
		// AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
		// 	AddItem(tview.NewTextView().SetText("HHH").SetBorder(true).SetTitle("Top"), 0, 1, false).
		// 	AddItem(tview.NewTextView().SetBorder(true).SetTitle("Middle (3 x height of Top)"), 0, 3, false).
		// 	AddItem(tview.NewTextView().SetBorder(true).SetTitle("Bottom (5 rows)"), 5, 1, false), 0, 2, false).
		AddItem(tview.NewTextView().SetBorder(true).SetTitle(" Navigation "), 0, 1, false)
	flex.SetBorder(true).SetTitle(" Help ")
	app.Pages.AddPage("HELP", ui.Modal(flex, 150, 30, view.closeModal), true, true)
}

// Build info pages for task page
func (v *helpView) headerBuilder() *tview.Pages {
	title := "HELP"
	entityName := title
	pageName := title
	items := v.headerPagesParam()
	v.buildHeaderPages(items, title, entityName)
	v.headerPages.SwitchToPage(pageName)
	return v.headerPages
}

// Build table for task page
func (v *helpView) bodyBuilder() *tview.Pages {
	return v.bodyPages
}

// Build footer for task page
func (v *helpView) footerBuilder() *tview.Flex {
	v.footer.help.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Generate info pages params
func (v *helpView) headerPagesParam() (items []headerItem) {
	items = []headerItem{
		{name: "Revision", value: "HELLO"},
		{name: "Task role", value: "WORLD"},
		{name: "Execution role", value: "E1S"},
	}
	return
}
