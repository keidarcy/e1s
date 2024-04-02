package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type TaskView struct {
	View
	tasks []types.Task
}

func newTaskView(tasks []types.Task, app *App) *TaskView {
	keys := append(basicKeyInputs, []KeyInput{
		{key: string(tKey), description: describeTaskDefinition},
		{key: string(vKey), description: describeTaskDefinitionRevisions},
		{key: string(eKey), description: editTaskDefinition},
		{key: string(lKey), description: showLogs},
	}...)
	return &TaskView{
		View: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind:             descriptionPageKeys,
			LogKind:                     logPageKeys,
			TaskDefinitionKind:          descriptionPageKeys,
			TaskDefinitionRevisionsKind: descriptionPageKeys,
		}),
		tasks: tasks,
	}
}

func (app *App) showTasksPages(reload bool, rowIndex int) error {
	if switched := app.SwitchPage(reload); switched {
		return nil
	}

	tasks, err := app.Store.ListTasks(app.cluster.ClusterName, app.service.ServiceName)

	if err != nil {
		logger.Warnf("Failed to show tasks pages, error: %v", err)
		app.back()
		return err
	}

	// no tasks exists do nothing
	if len(tasks) == 0 {
		app.back()
		return fmt.Errorf("no valid task")
	}

	view := newTaskView(tasks, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(rowIndex, 0)
	return nil
}

// Build info pages for task page
func (v *TaskView) infoBuilder() *tview.Pages {
	for _, t := range v.tasks {
		title := util.ArnToName(t.TaskArn)
		entityName := *t.TaskArn
		items := v.infoPagesParam(t)

		v.buildInfoPages(items, title, entityName)
	}
	// prevent empty tasks
	if len(v.tasks) > 0 && v.tasks[0].TaskArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.tasks[0].TaskArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for task page
func (v *TaskView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for task page
func (v *TaskView) footerBuilder() *tview.Flex {
	v.footer.task.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for task table
func (v *TaskView) tableHandler() {
	for row, task := range v.tasks {
		t := task
		v.table.GetCell(row+1, 0).SetReference(Entity{task: &t, entityName: *t.TaskArn})
	}
}

// Generate info pages params
func (v *TaskView) infoPagesParam(t types.Task) (items []InfoItem) {
	// containers
	containers := []string{}
	for _, c := range t.Containers {
		containers = append(containers, *c.Name)
	}
	// network
	subnetID := util.EmptyText
	eniID := util.EmptyText
	privateIP := util.EmptyText
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
	items = []InfoItem{
		{name: "Task ID", value: util.ArnToName(t.TaskArn)},
		{name: "Task definition", value: util.ArnToName(t.TaskDefinitionArn)},
		{name: "Containers", value: strings.Join(containers, ",")},
		{name: "Cluster", value: util.ArnToName(t.ClusterArn)},
		{name: "Launch type", value: string(t.LaunchType)},
		{name: "Capacity provider", value: util.ShowString(t.CapacityProviderName)},
		{name: "Group", value: util.ShowString(t.Group)},
		{name: "Subnet ID", value: subnetID},
		{name: "ENI ID", value: eniID},
		{name: "Private IP", value: privateIP},
		{name: "Execute command", value: strconv.FormatBool(t.EnableExecuteCommand)},
		{name: "Started by", value: util.ShowString(t.StartedBy)},
		{name: "Started at", value: util.ShowTime(t.StartedAt)},
		{name: "Pull started at", value: util.ShowTime(t.PullStartedAt)},
		{name: "Pull stopped at", value: util.ShowTime(t.PullStoppedAt)},
		{name: "StoppedReason", value: util.ShowString(t.StoppedReason)},
		{name: "StoppedAt", value: util.ShowTime(t.StoppedAt)},
		{name: "StoppingAt", value: util.ShowTime(t.StoppingAt)},
		{name: "Platform family", value: util.ShowString(t.PlatformFamily)},
		{name: "Platform version", value: util.ShowString(t.PlatformVersion)},
	}
	return
}

// Generate table params
func (v *TaskView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, *v.app.service.ServiceName, len(v.tasks))
	headers = []string{
		"Task ID ▾",
		"Last status",
		"Desired status",
		"Task definition",
		"Started at",
		"Containers",
		"Health status",
		"Launch type",
		"CPU",
		"Memory",
	}
	dataBuilder = func() (data [][]string) {
		for _, t := range v.tasks {
			// healthy status
			health := string(t.HealthStatus)

			row := []string{}
			row = append(row, util.ArnToName(t.TaskArn))
			row = append(row, util.ShowGreenGrey(t.LastStatus, "running"))
			row = append(row, util.ShowGreenGrey(t.DesiredStatus, "running"))
			row = append(row, util.ArnToName(t.TaskDefinitionArn))
			row = append(row, util.ShowTime(t.StartedAt))
			row = append(row, strconv.Itoa(len(t.Containers)))
			row = append(row, util.ShowGreenGrey(&health, "healthy"))
			row = append(row, string(t.LaunchType))
			row = append(row, util.ShowString(t.Cpu))
			row = append(row, util.ShowString(t.Memory))
			data = append(data, row)
		}
		return data
	}

	return
}

// task definition arn to family and revision
func getTaskDefinitionInfo(arn *string) (family, revision string) {
	if arn == nil {
		return util.EmptyText, util.EmptyText
	}
	td := strings.Split(util.ArnToName(arn), ":")
	if len(td) < 2 {
		return util.EmptyText, util.EmptyText
	}
	family = td[0]
	revision = td[1]
	return
}
