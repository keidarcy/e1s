package ui

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type ContainerView struct {
	View
	containers []types.Container
}

func newContainerView(containers []types.Container, app *App) *ContainerView {
	keys := append(basicKeyInputs, []KeyInput{
		{key: "shift-f", description: portForwarding},
		{key: "shift-t", description: terminatePortForwardingSession},
		{key: "enter", description: sshContainer},
	}...)
	return &ContainerView{
		View: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: descriptionPageKeys,
		}),
		containers: containers,
	}
}

func (app *App) showContainersPage(reload bool) error {
	if switched := app.SwitchPage(reload); switched {
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
func (v *ContainerView) infoBuilder() *tview.Pages {
	for _, c := range v.containers {
		title := util.ArnToName(c.ContainerArn)
		entityName := *c.ContainerArn
		items := v.infoPagesParam(c)

		v.buildInfoPages(items, title, entityName)
	}
	// prevent empty containers
	if len(v.containers) > 0 && v.containers[0].ContainerArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.containers[0].ContainerArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for container page
func (v *ContainerView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for container page
func (v *ContainerView) footerBuilder() *tview.Flex {
	v.footer.container.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for container table
func (v *ContainerView) tableHandler() {
	for row, container := range v.containers {
		c := container
		v.table.GetCell(row+1, 0).SetReference(Entity{container: &c, entityName: *c.ContainerArn})
	}

	v.table.SetSelectedFunc(func(row int, column int) {
		containerName := v.table.GetCell(row, column).Text
		v.ssh(containerName)
	})

	// v.table.SetInputCapture(v.handleInputCapture)
}

// Generate info pages params
func (v *ContainerView) infoPagesParam(c types.Container) (items []InfoItem) {
	// Managed agents
	mas := []string{}
	for _, m := range c.ManagedAgents {
		mas = append(mas, string(m.Name))
	}
	masString := strings.Join(mas, ",")
	if len(masString) == 0 {
		masString = util.EmptyText
	}

	items = []InfoItem{
		{name: "Name", value: util.ShowString(c.Name)},
		{name: "Task", value: util.ShowString(c.TaskArn)},
		{name: "Image url", value: util.ShowString(c.Image)},
		{name: "Image digest", value: util.ShowString(c.ImageDigest)},
		{name: "Runtime ID", value: util.ShowString(c.RuntimeId)},
		{name: "Last status", value: util.ShowString(c.LastStatus)},
		{name: "CPU", value: util.ShowString(c.Cpu)},
		{name: "Memory", value: util.ShowString(c.Memory)},
		{name: "Memory reservation", value: util.ShowString(c.MemoryReservation)},
		{name: "GPU IDs", value: util.ShowArray(c.GpuIds)},
		{name: "Exit code", value: util.ShowInt(c.ExitCode)},
		{name: "Reason", value: util.ShowString(c.Reason)},
		{name: "Managed agents", value: util.ShowString(&masString)},
	}
	return
}

// Generate table params
func (v *ContainerView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, util.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Status",
		"Health status ▾",
		"PF",
		"Container runtime id",
		"Image URI",
	}

	dataBuilder = func() (data [][]string) {
		for _, c := range v.containers {
			containerId := fmt.Sprintf("%s.%s", *v.app.cluster.ClusterName, *c.Name)
			portText := util.EmptyText
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

			row := []string{}
			row = append(row, util.ShowString(c.Name))
			row = append(row, util.ShowGreenGrey(c.LastStatus, "running"))
			row = append(row, util.ShowGreenGrey(&health, "healthy"))
			row = append(row, portText)
			row = append(row, util.ShowString(c.RuntimeId))
			row = append(row, util.ShowString(c.Image))
			data = append(data, row)
		}
		return data
	}

	return
}
