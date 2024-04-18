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

func (app *App) showHelpPage() {
	view := newHelpView(app)
	left := genColumn("Resource", []keyInput{
		hotKeyMap["enter"],
		hotKeyMap["esc"],
		hotKeyMap["ctrlZ"],
		hotKeyMap["ctrlC"],
		hotKeyMap["ctrlR"],
		hotKeyMap["?"],
		hotKeyMap["b"],
		hotKeyMap["d"],
		hotKeyMap["e"],
	})
	right := genColumn("Navigation", []keyInput{
		hotKeyMap["j"],
		hotKeyMap["k"],
		hotKeyMap["G"],
		hotKeyMap["g"],
		hotKeyMap["ctrlF"],
		hotKeyMap["ctrlB"],
	})
	flex := tview.NewFlex().
		AddItem(left, 0, 1, false).
		AddItem(right, 0, 1, false)
	flex.SetBorder(true).SetTitle(" Help ")
	app.Pages.AddPage("help", ui.Modal(flex, 150, 30, view.closeModal), true, true)
}

func genColumn(title string, keys []keyInput) tview.Primitive {
	t := tview.NewTable()
	adjust := 2
	t.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf("[aqua::b]%s", title)).SetAlign(L))
	for i, k := range keys {
		key := tview.NewTableCell(fmt.Sprintf("[purple]<%s>", k.key)).SetAlign(L).SetExpansion(1)
		description := tview.NewTableCell(fmt.Sprintf("[green]%s", k.description)).SetAlign(L).SetExpansion(5)
		t.SetCell(i+adjust, 0, key)
		t.SetCell(i+adjust, 1, description)
	}
	return t
}
