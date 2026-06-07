package view

import (
	"testing"

	"github.com/rivo/tview"
)

func TestCopyablePageName(t *testing.T) {
	app, _ := newApp(Option{Splash: true})
	app.kind = ClusterKind
	v := newView(app, nil, nil)
	v.table.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false))
	v.table.SetCell(1, 0, tview.NewTableCell(" cluster-a "))

	v.table.Select(1, 0)

	want := app.kind.getTablePageName(app.getPageHandle()) + ".cluster-a"
	if got := v.copyablePageName(); got != want {
		t.Errorf("Got: %s, Want: %s", got, want)
	}
}

func TestCopyablePageNameUsesFirstDataRowWhenHeaderSelected(t *testing.T) {
	app, _ := newApp(Option{Splash: true})
	app.kind = ClusterKind
	v := newView(app, nil, nil)
	v.table.SetCell(0, 0, tview.NewTableCell("Name").SetSelectable(false))
	v.table.SetCell(1, 0, tview.NewTableCell("cluster-a"))

	v.table.Select(0, 0)

	want := app.kind.getTablePageName(app.getPageHandle()) + ".cluster-a"
	if got := v.copyablePageName(); got != want {
		t.Errorf("Got: %s, Want: %s", got, want)
	}
}
