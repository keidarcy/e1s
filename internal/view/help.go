package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
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

func (app *App) showHelpPage(pageName string) error {
	view := newHelpView(app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.bodyPages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlZ {
			if app.Pages.HasPage(pageName) {
				logger.Info(pageName)
				app.Pages.SwitchToPage("clusters")
				go func() {
					app.Application.Draw()
				}()
				logger.Info(pageName)
			} else {
				logger.Info("WHY")
			}
		}
		if event.Rune() == 12 {
			app.Pages.SwitchToPage(pageName)
		}
		return event
	})
	return nil
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
