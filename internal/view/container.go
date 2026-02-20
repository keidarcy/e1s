package view

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type containerView struct {
	view
	containers []types.Container
}

func newContainerView(containers []types.Container, app *App) *containerView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["L"],
		hotKeyMap["F"],
		hotKeyMap["T"],
		hotKeyMap["P"],
		hotKeyMap["D"],
		hotKeyMap["E"],
		hotKeyMap["s"],
		hotKeyMap["ctrlD"],
	}...)
	return &containerView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
			LogKind:         logPageKeys,
		}),
		containers: containers,
	}
}

func (app *App) showContainersPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	if app.task == nil {
		app.back()
		app.Notice.Warnf("no valid task")
		return fmt.Errorf("no valid task")
	}

	resources := app.task.Containers
	err := buildResourcePage(resources, app, nil, func() resourceViewBuilder {
		return newContainerView(resources, app)
	})
	return err
}

func (v *containerView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.container
}

func (v *containerView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.containers))
	for i, c := range v.containers {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(c.ContainerArn),
			entityName: *c.ContainerArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *containerView) headerPageItems(index int) (items []headerItem) {
	c := v.containers[index]
	// Managed agents
	mas := []string{}
	for _, m := range c.ManagedAgents {
		mas = append(mas, string(m.Name))
	}
	masString := strings.Join(mas, ",")
	if len(masString) == 0 {
		masString = utils.EmptyText
	}

	items = []headerItem{
		{name: "Name", value: utils.ShowString(c.Name)},
		{name: "Task", value: utils.ShowString(c.TaskArn)},
		{name: "Image url", value: utils.ShowString(c.Image)},
		{name: "Image digest", value: utils.ShowString(c.ImageDigest)},
		{name: "Runtime ID", value: utils.ShowString(c.RuntimeId)},
		{name: "Last status", value: utils.ShowString(c.LastStatus)},
		{name: "CPU", value: utils.ShowString(c.Cpu)},
		{name: "Memory", value: utils.ShowString(c.Memory)},
		{name: "Memory reservation", value: utils.ShowString(c.MemoryReservation)},
		{name: "GPU IDs", value: utils.ShowArray(c.GpuIds)},
		{name: "Exit code", value: utils.ShowInt(c.ExitCode)},
		{name: "Reason", value: utils.ShowString(c.Reason)},
		{name: "Managed agents", value: utils.ShowString(&masString)},
	}
	return
}

// Generate table params
func (v *containerView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, utils.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Status",
		"Health",
		"PF",
		"Registry",
		"Image name",
	}

	rowsBuilder = func() (data [][]string) {
		for _, c := range v.containers {
			containerId := fmt.Sprintf("%s.%s", *v.app.cluster.ClusterName, *c.Name)
			portText := utils.EmptyText
			ports := []string{}
			for _, session := range v.app.sessions {
				if session.containerId == containerId {
					ports = append(ports, session.port)
				}
			}
			if len(ports) != 0 {
				portText = strings.Join(ports, ",")
			}
			health := string(c.HealthStatus)

			registry, imageName := utils.ImageInfo(c.Image)

			row := []string{}
			row = append(row, utils.ShowString(c.Name))
			row = append(row, utils.ShowGreenGrey(c.LastStatus, "running"))
			row = append(row, utils.ShowGreenGrey(&health, "healthy"))
			row = append(row, portText)
			row = append(row, registry)
			row = append(row, imageName)
			data = append(data, row)

			entity := Entity{container: &c, entityName: *c.ContainerArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}

	return
}
