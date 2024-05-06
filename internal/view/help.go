package view

import (
	"fmt"
	"strconv"

	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/rivo/tview"
)

type helpView struct {
	view
}

func newHelpView(app *App) *helpView {
	keys := append(basicKeyInputs, []keyDescriptionPair{}...)
	return &helpView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
	}
}

func (app *App) showHelpPage() {
	view := newHelpView(app)
	resource := genColumn("Resource", []keyDescriptionPair{
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
	navigation := genColumn("Navigation", []keyDescriptionPair{
		hotKeyMap["j"],
		hotKeyMap["k"],
		hotKeyMap["G"],
		hotKeyMap["g"],
		hotKeyMap["ctrlF"],
		hotKeyMap["ctrlB"],
	})
	info := genColumn("Info", []keyDescriptionPair{
		{key: "debug", description: strconv.FormatBool(app.Option.Debug)},
		{key: "json", description: strconv.FormatBool(app.Option.JSON)},
		{key: "read-only", description: strconv.FormatBool(app.Option.ReadOnly)},
		{key: "log-file", description: app.Option.LogFile},
		{key: "config-file", description: app.Option.ConfigFile},
		{key: "shell", description: app.Option.Shell},
		{key: "refresh", description: strconv.Itoa(app.Option.Refresh)},
		{key: "theme", description: app.Option.Theme},
	})
	flex := tview.NewFlex().
		AddItem(resource, 0, 1, false).
		AddItem(navigation, 0, 1, false).
		AddItem(info, 0, 1, false)
	flex.SetBorder(true).SetTitle(" Help ")
	app.Pages.AddPage("help", ui.Modal(flex, 150, 25, view.closeModal), true, true)
}

func genColumn(title string, keys []keyDescriptionPair) tview.Primitive {
	t := tview.NewTable()
	t.SetBorderPadding(1, 0, 2, 0)
	adjust := 2
	t.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf(color.HelpTitleFmt, title)).SetAlign(L))
	for i, k := range keys {
		key := tview.NewTableCell(fmt.Sprintf(color.HelpKeyFmt, k.key)).SetAlign(L).SetExpansion(1)
		description := tview.NewTableCell(fmt.Sprintf(color.HelpDescriptionFmt, k.description)).SetAlign(L).SetExpansion(5)
		t.SetCell(i+adjust, 0, key)
		t.SetCell(i+adjust, 1, description)
	}
	return t
}
