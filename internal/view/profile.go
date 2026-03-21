package view

import (
	"fmt"

	"github.com/keidarcy/e1s/internal/api"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

type profileView struct {
	view
	profiles []api.Profile
}

func newProfileView(profiles []api.Profile, app *App) *profileView {
	return &profileView{
		view: *newView(app, tableInputs, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		profiles: profiles,
	}
}

func (app *App) showProfilesPage(reload bool) error {
	app.kind = ProfileKind
	if switched := app.switchPage(reload); switched {
		return nil
	}

	profiles, err := app.Store.ListProfiles()

	err = buildResourcePage(profiles, app, err, func() resourceViewBuilder {
		return newProfileView(profiles, app)
	})
	return err
}

func (v *profileView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.profile
}

func (v *profileView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.profiles))
	for i, p := range v.profiles {
		params = append(params, headerPageParam{
			title:      p.Name,
			entityName: p.Name,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *profileView) headerPageItems(index int) (items []headerItem) {
	p := v.profiles[index]
	region := p.DefaultRegion
	if region == "" {
		region = "—"
	}
	items = []headerItem{
		{name: "Name", value: p.Name},
		{name: "Source", value: p.Source},
		{name: "Default region", value: region},
		{name: "Auth style", value: p.AuthStyle},
	}
	return
}

// Generate table params
func (v *profileView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, "all", len(v.profiles))
	headers = []string{
		"Name",
		"Source",
		"Default region",
		"Auth style",
	}

	rowsBuilder = func() (data [][]string) {
		for _, p := range v.profiles {
			region := p.DefaultRegion
			if region == "" {
				region = "—"
			}
			row := []string{p.Name, p.Source, region, p.AuthStyle}
			data = append(data, row)
			entity := Entity{profile: p.Name, entityName: p.Name}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
