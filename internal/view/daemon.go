package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type daemonView struct {
	view
	daemons []types.DaemonSummary
}

func newDaemonView(daemons []types.DaemonSummary, app *App) *daemonView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["t"],
	}...)
	return &daemonView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		daemons: daemons,
	}
}

func (app *App) showDaemonsPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	resources, err := app.Store.ListDaemons(app.cluster.ClusterArn)
	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newDaemonView(resources, app)
	})
	return err
}

func (v *daemonView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.daemon
}

func (v *daemonView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.daemons))
	for i, d := range v.daemons {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(d.DaemonArn),
			entityName: *d.DaemonArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

func (v *daemonView) headerPageItems(index int) (items []headerItem) {
	d := v.daemons[index]
	items = []headerItem{
		{name: "Daemon", value: utils.ArnToName(d.DaemonArn)},
		{name: "Status", value: string(d.Status)},
		{name: "Created at", value: utils.ShowTime(d.CreatedAt)},
		{name: "Updated at", value: utils.ShowTime(d.UpdatedAt)},
	}
	return
}

func (v *daemonView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, *v.app.cluster.ClusterName, len(v.daemons))
	headers = []string{
		"Daemon",
		"Status",
		"Created",
		"Updated",
	}
	rowsBuilder = func() (data [][]string) {
		for _, d := range v.daemons {
			status := string(d.Status)
			row := []string{
				utils.ArnToName(d.DaemonArn),
				utils.ShowGreenGrey(&status, "active"),
				utils.Age(d.CreatedAt),
				utils.Age(d.UpdatedAt),
			}
			data = append(data, row)

			entity := Entity{daemonSummary: &d, entityName: *d.DaemonArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
