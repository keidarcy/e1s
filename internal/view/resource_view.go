package view

import (
	"fmt"
	"log/slog"

	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

type headerPageParam struct {
	title      string
	entityName string
	items      []headerItem
}

// Interface to show each view
type resourceViewBuilder interface {
	getViewAndFooter() (*view, *tview.TextView)

	headerParamsBuilder() []headerPageParam
	headerPageItems(index int) []headerItem

	tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string)
}

// Common function to build resource page for each view
func buildResourcePage[T any](
	resources []T,
	app *App,
	err error,
	newResourceViewBuilder func() resourceViewBuilder,
) error {
	err = resourceViewPreHandler(resources, app, err)
	if err != nil {
		return err
	}
	b := newResourceViewBuilder()
	v, footer := b.getViewAndFooter()

	// table builder
	title, headers, rowsBuilder := b.tableParamsBuilder()
	v.buildTable(title, headers, rowsBuilder)

	// header pages builder
	headerPageParams := b.headerParamsBuilder()
	v.buildHeaderPages(headerPageParams)
	if len(headerPageParams) > 0 {
		v.headerPages.SwitchToPage(headerPageParams[0].entityName)
		v.changeSelectedValues()
	}

	// footer builder
	footer.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()

	page := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(v.headerPages, oneColumnCount+4, 1, false).
		AddItem(v.tablePages, 0, 2, true).
		AddItem(v.footer.footerFlex, 1, 1, false)

	v.mainFlex = page
	v.initFilterInput()

	// Restore sort/filter state from a previous view of this kind (e.g. after auto-refresh).
	if state, ok := app.viewStates[app.kind]; ok {
		if state.filterText != "" {
			v.filterInput.SetText(state.filterText)
		}
		if state.sortColumn >= 0 && state.sortColumn < len(v.headers) && len(v.originalRowData) > 0 {
			v.sortColumn = state.sortColumn
			v.sortOrder = state.sortOrder
			sortedIdx := v.getSortedOriginalIndexWithFilterText(v.sortColumn)
			sortedData := make([][]string, 0, len(sortedIdx))
			sortedRefs := make([]Entity, 0, len(sortedIdx))
			for _, idx := range sortedIdx {
				sortedData = append(sortedData, v.originalRowData[idx])
				sortedRefs = append(sortedRefs, v.originalRowReferences[idx])
			}
			v.table.Clear()
			v.buildTableContent(sortedData, sortedRefs)
			v.changeSelectedValues()
			v.updateFilterTitle()
		} else if state.filterText != "" {
			v.applyFilter()
		}
	}

	v.app.addAppPage(page)
	v.table.Select(v.app.rowIndex, 0)
	return nil
}

func resourceViewPreHandler[T any](resources []T, app *App, err error) error {
	if err != nil {
		slog.Warn("failed to show "+app.kind.String()+" pages in resourcePagePreHandler", "error", err)
		app.back()
		return err
	}
	if len(resources) == 0 {
		errMsg := "no " + app.kind.String() + " found"
		slog.Warn(errMsg + " in resourcePagePreHandler")
		if app.kind == ClusterKind {
			clusterPage := ClusterKind.getAppPageName(app.getPageHandle())
			// No cluster page exists for this profile/region yet (startup or newly
			// switched scope), so navigate to regions instead of switching to a
			// non-existent page via back().
			if !app.Pages.HasPage(clusterPage) {
				if regionErr := app.showRegionsPage(false); regionErr != nil {
					return fmt.Errorf("%s; failed to show regions: %w", errMsg, regionErr)
				}
				app.Notice.Warnf("No clusters found in region %s. Choose another region.", globalRegion)
				return ErrHandledNavigation
			}
		}
		err = fmt.Errorf("%s", errMsg)
		app.back()
		return err
	}
	return err
}
