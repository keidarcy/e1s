package view

import (
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

// Add new type for instance view
type instanceView struct {
	view
	instances []types.ContainerInstance
}

// Constructor for instance view
func newInstanceView(instances []types.ContainerInstance, app *App) *instanceView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["v"],
		hotKeyMap["s"],
	}...)
	return &instanceView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		instances: instances,
	}
}

// Show instances page
func (app *App) showInstancesPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	instances, err := app.Store.ListContainerInstances(app.cluster.ClusterName)
	if err != nil {
		slog.Warn("failed to show instances page", "error", err)
		app.back()
		return err
	}

	if len(instances) == 0 {
		app.back()
		return fmt.Errorf("no instances found")
	}

	view := newInstanceView(instances, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for instance page
func (v *instanceView) headerBuilder() *tview.Pages {
	for _, instance := range v.instances {
		title := utils.ArnToName(instance.ContainerInstanceArn)
		entityName := *instance.ContainerInstanceArn
		items := v.headerPagesParam(instance)

		v.buildHeaderPages(items, title, entityName)
	}

	if len(v.instances) > 0 && v.instances[0].ContainerInstanceArn != nil {
		v.headerPages.SwitchToPage(*v.instances[0].ContainerInstanceArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Generate info pages params
func (v *instanceView) headerPagesParam(instance types.ContainerInstance) (items []headerItem) {
	items = []headerItem{
		{name: "Status", value: utils.ShowString(instance.Status)},
		{name: "Instance Type", value: utils.ShowString(instance.Ec2InstanceId)},
		{name: "Agent Connected", value: fmt.Sprintf("%v", instance.AgentConnected)},
		{name: "Running Tasks Count", value: fmt.Sprintf("%d", instance.RunningTasksCount)},
		{name: "Pending Tasks Count", value: fmt.Sprintf("%d", instance.PendingTasksCount)},
		{name: "Agent Version", value: utils.ShowString(instance.VersionInfo.AgentVersion)},
		{name: "Docker Version", value: utils.ShowString(instance.VersionInfo.DockerVersion)},
		{name: "Registered At", value: utils.ShowTime(instance.RegisteredAt)},
	}
	return
}

// Build footer for instance page
func (v *instanceView) footerBuilder() *tview.Flex {
	v.footer.instance.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Build table for instance page
func (v *instanceView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Handlers for instance table
func (v *instanceView) tableHandler() {
	for row, instance := range v.instances {
		i := instance
		v.table.GetCell(row+1, 0).SetReference(Entity{instance: &i, entityName: *i.ContainerInstanceArn})
	}
}

// Generate table params
func (v *instanceView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	clusterName := ""
	if v.app.cluster.ClusterName != nil {
		clusterName = *v.app.cluster.ClusterName
	}

	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, clusterName, len(v.instances))
	headers = []string{
		"Instance ID ▾",
		"Status",
		"Running Tasks",
		"Pending Tasks",
		"Agent Connected",
		"Agent Version",
		"Docker Version",
		"Registered At",
	}

	dataBuilder = func() (data [][]string) {
		for _, instance := range v.instances {
			row := []string{
				utils.ArnToName(instance.ContainerInstanceArn),
				utils.ShowGreenGrey(instance.Status, "ACTIVE"),
				fmt.Sprintf("%d", instance.RunningTasksCount),
				fmt.Sprintf("%d", instance.PendingTasksCount),
				fmt.Sprintf("%v", instance.AgentConnected),
				utils.ShowString(instance.VersionInfo.AgentVersion),
				utils.ShowString(instance.VersionInfo.DockerVersion),
				utils.ShowTime(instance.RegisteredAt),
			}
			data = append(data, row)
		}
		return data
	}

	return
}
