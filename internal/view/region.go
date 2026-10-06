package view

import (
	"fmt"

	"github.com/keidarcy/e1s/internal/api"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

type regionView struct {
	view
	regions []api.Region
}

func newRegionView(regions []api.Region, app *App) *regionView {
	v := &regionView{
		view: *newView(app, tableInputs, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		regions: regions,
	}
	app.regionsView = v
	return v
}

func (app *App) showRegionsPage(reload bool) error {
	app.kind = RegionKind
	if switched := app.switchPage(reload); switched {
		if app.regionsView != nil {
			app.regionsView.table.SetTitle(app.regionsView.tableTitle())
		}
		return nil
	}

	regions, err := app.Store.ListRegions()

	err = buildResourcePage(regions, app, err, func() resourceViewBuilder {
		return newRegionView(regions, app)
	})
	return err
}

func (v *regionView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.region
}

func (v *regionView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.regions))
	for i, r := range v.regions {
		params = append(params, headerPageParam{
			title:      r.Code,
			entityName: r.Code,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *regionView) headerPageItems(index int) (items []headerItem) {
	r := v.regions[index]
	items = []headerItem{
		{name: "Code", value: r.Code},
		{name: "Name", value: r.Name},
		{name: "Enabled", value: r.Enabled},
	}
	return
}

func (v *regionView) tableTitle() string {
	scope := "all"
	if v.app.regionWithoutClusters != "" {
		scope = " no ECS clusters in " + v.app.regionWithoutClusters + " "
	}
	return fmt.Sprintf(color.TableTitleFmt, RegionKind, scope, len(v.regions))
}

func (v *regionView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = v.tableTitle()
	headers = []string{
		"Code",
		"Name",
		"Enabled",
	}

	rowsBuilder = func() (data [][]string) {
		for _, r := range v.regions {
			row := []string{}
			row = append(row, r.Code)
			row = append(row, r.Name)
			row = append(row, r.Enabled)
			data = append(data, row)

			entity := Entity{region: &r, entityName: r.Code}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
