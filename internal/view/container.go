package view

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type containerView struct {
	view
	containers []types.Container
}

func newContainerView(containers []types.Container, app *App) *containerView {
	keys := append(basicKeyInputs, []keyInput{
		hotKeyMap["F"],
		hotKeyMap["T"],
		hotKeyMap["P"],
		hotKeyMap["E"],
		hotKeyMap["enter"],
		hotKeyMap["ctrlD"],
	}...)
	return &containerView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		containers: containers,
	}
}

func (app *App) showContainersPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	// no containers exists do nothing
	if app.task == nil || len(app.task.Containers) == 0 {
		app.back()
		return fmt.Errorf("no valid container")
	}
	view := newContainerView(app.task.Containers, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
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
	v.tableHandler()
	return v.bodyPages
}

// Build footer for container page
func (v *containerView) footerBuilder() *tview.Flex {
	v.footer.container.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Handlers for container table
func (v *containerView) tableHandler() {
	for row, container := range v.containers {
		c := container
		v.table.GetCell(row+1, 0).SetReference(Entity{container: &c, entityName: *c.ContainerArn})
	}

	v.table.SetSelectedFunc(func(row int, column int) {
		v.execShell()
	})
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
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, utils.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Status",
		"Health status ▾",
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
		}
		return data
	}

	return
}
