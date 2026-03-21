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
	return &regionView{
		view: *newView(app, tableInputs, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		regions: regions,
	}
}

func (app *App) showRegionsPage(reload bool) error {
	app.kind = RegionKind
	if switched := app.switchPage(reload); switched {
		return nil
	}

	regions, err := app.Store.ListRegions()

	err = buildResourcePage(regions, app, nil, func() resourceViewBuilder {
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

// Generate table params
func (v *regionView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, "all", len(v.regions))
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
