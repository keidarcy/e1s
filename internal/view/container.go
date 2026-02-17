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

func (v *containerView) getView() *view {
	return &v.view
}

// Build info pages for container page
func (v *containerView) headerBuilder() *tview.Pages {
	for _, c := range v.containers {
		title := utils.ArnToName(c.ContainerArn)
		entityName := *c.ContainerArn
		items := v.headerPagesParam(c)

		v.buildHeaderPages(items, title, entityName)
	}
	// prevent empty containers
	if len(v.containers) > 0 && v.containers[0].ContainerArn != nil {
		// show first when enter
		v.headerPages.SwitchToPage(*v.containers[0].ContainerArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Build table for container page
func (v *containerView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.table.SetSelectedFunc(func(row int, column int) {
		v.execShell()
	})
	return v.bodyPages
}

// Build footer for container page
func (v *containerView) footerBuilder() *tview.Flex {
	v.footer.container.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Generate info pages params
func (v *containerView) headerPagesParam(c types.Container) (items []headerItem) {
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
func (v *containerView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, utils.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Status",
		"Health",
		"PF",
		"Registry",
		"Image name",
	}

	dataBuilder = func() (data [][]string) {
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
