package view

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type taskView struct {
	view
	tasks []types.Task
}

func newTaskView(tasks []types.Task, app *App) *taskView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["t"],
		hotKeyMap["l"],
		hotKeyMap["s"],
		hotKeyMap["S"],
	}...)
	return &taskView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
			LogKind:         logPageKeys,
		}),
		tasks: tasks,
	}
}

func (app *App) showTasksPages(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	var serviceName *string = nil
	if app.service != nil {
		serviceName = app.service.ServiceName
	}

	// true when show tasks from cluster
	if app.fromCluster {
		serviceName = nil
	}

	// true when show desiredStatus:stopped tasks
	if app.taskStatus == types.DesiredStatusStopped {
		serviceName = nil
	}

	tasks, err := app.Store.ListTasks(app.cluster.ClusterName, serviceName, app.taskStatus)

	if err != nil {
		slog.Warn("failed to show tasks pages", "error", err)
		app.back()
		return err
	}

	// no tasks exists do nothing
	if len(tasks) == 0 {
		err := fmt.Errorf("no valid %s task", strings.ToLower(string(app.taskStatus)))
		app.back()
		return err
	}

	view := newTaskView(tasks, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for task page
func (v *taskView) headerBuilder() *tview.Pages {
	for _, t := range v.tasks {
		title := utils.ArnToName(t.TaskArn)
		entityName := *t.TaskArn
		items := v.headerPagesParam(t)

		v.buildHeaderPages(items, title, entityName)
	}
	// prevent empty tasks
	if len(v.tasks) > 0 && v.tasks[0].TaskArn != nil {
		// show first when enter
		v.headerPages.SwitchToPage(*v.tasks[0].TaskArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Build table for task page
func (v *taskView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Build footer for task page
func (v *taskView) footerBuilder() *tview.Flex {
	v.footer.task.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Handlers for task table
func (v *taskView) tableHandler() {
	for row, task := range v.tasks {
		t := task
		v.table.GetCell(row+1, 0).SetReference(Entity{task: &t, entityName: *t.TaskArn})
	}
}

// Generate info pages params
func (v *taskView) headerPagesParam(t types.Task) (items []headerItem) {
	// containers
	containers := []string{}
	for _, c := range t.Containers {
		containers = append(containers, *c.Name)
	}
	// network
	subnetID := utils.EmptyText
	eniID := utils.EmptyText
	privateIP := utils.EmptyText
	for _, a := range t.Attachments {
		if *a.Type == "ElasticNetworkInterface" {
			for _, d := range a.Details {
				if *d.Name == "subnetId" {
					subnetID = *d.Value
				}
				if *d.Name == "networkInterfaceId" {
					eniID = *d.Value
				}

				if *d.Name == "privateIPv4Address" {
					privateIP = *d.Value
				}
			}
		}
	}

	items = []headerItem{
		{name: "Task ID", value: utils.ArnToName(t.TaskArn)},
		{name: "Task definition", value: utils.ArnToName(t.TaskDefinitionArn)},
		{name: "Containers", value: strings.Join(containers, ",")},
		{name: "Cluster", value: utils.ArnToName(t.ClusterArn)},
		{name: "Launch type", value: string(t.LaunchType)},
		{name: "Capacity provider", value: utils.ShowString(t.CapacityProviderName)},
		{name: "Subnet ID", value: subnetID},
		{name: "ENI ID", value: eniID},
		{name: "Private IP", value: privateIP},
		{name: "Execute command", value: strconv.FormatBool(t.EnableExecuteCommand)},
		{name: "Started by", value: utils.ShowString(t.StartedBy)},
		{name: "Started at", value: utils.ShowTime(t.StartedAt)},
		{name: "Pull started at", value: utils.ShowTime(t.PullStartedAt)},
		{name: "Pull stopped at", value: utils.ShowTime(t.PullStoppedAt)},
		{name: "StoppedReason", value: utils.ShowString(t.StoppedReason)},
		{name: "StoppedAt", value: utils.ShowTime(t.StoppedAt)},
		{name: "StoppingAt", value: utils.ShowTime(t.StoppingAt)},
		{name: "Platform family", value: utils.ShowString(t.PlatformFamily)},
		{name: "Platform version", value: utils.ShowString(t.PlatformVersion)},
		{name: "Tags count", value: strconv.Itoa(len(t.Tags))},
	}
	return
}

// Generate table params
func (v *taskView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	parent := *v.app.service.ServiceName
	if v.app.taskStatus == types.DesiredStatusStopped {
		parent = *v.app.cluster.ClusterName
	}
	title = fmt.Sprintf(color.TableTitleFmt, fmt.Sprintf("%s.%s", v.app.kind, strings.ToLower(string(v.app.taskStatus))), parent, len(v.tasks))
	headers = []string{
		"Task ID ▾",
		"Last status",
		"Health",
		"Service",
		"Task definition",
		"Containers",
		"CPU",
		"Memory",
		"Age",
	}
	dataBuilder = func() (data [][]string) {
		for _, t := range v.tasks {
			// healthy status
			health := string(t.HealthStatus)

			row := []string{}
			row = append(row, utils.ArnToName(t.TaskArn))
			row = append(row, utils.ShowGreenGrey(t.LastStatus, "running"))
			row = append(row, utils.ShowGreenGrey(&health, "healthy"))
			row = append(row, utils.ShowServiceByGroup(t.Group))
			row = append(row, utils.ArnToName(t.TaskDefinitionArn))
			row = append(row, strconv.Itoa(len(t.Containers)))
			row = append(row, utils.ShowString(t.Cpu))
			row = append(row, utils.ShowString(t.Memory))
			row = append(row, utils.Age(t.StartedAt))
			data = append(data, row)
		}
		return data
	}

	return
}

// task definition arn to family and revision
func getTaskDefinitionInfo(arn *string) (family, revision string) {
	if arn == nil {
		return utils.EmptyText, utils.EmptyText
	}
	td := strings.Split(utils.ArnToName(arn), ":")
	if len(td) < 2 {
		return utils.EmptyText, utils.EmptyText
	}
	family = td[0]
	revision = td[1]
	return
}
