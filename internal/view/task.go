package view

import (
	"fmt"
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
		hotKeyMap["L"],
		hotKeyMap["x"],
		hotKeyMap["S"],
		hotKeyMap["s"],
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

	resources, noRunningShowStopped, err := app.Store.ListTasks(app.cluster.ClusterName, serviceName, app.taskStatus)

	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		if noRunningShowStopped && len(resources) > 0 {
			app.Notice.Warn("0 running task show stopped")
		}
		return newTaskView(resources, app)
	})
	return err
}

func (v *taskView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.task
}

// Build info pages for task page
func (v *taskView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.tasks))
	for i, t := range v.tasks {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(t.TaskArn),
			entityName: *t.TaskArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *taskView) headerPageItems(index int) (items []headerItem) {
	t := v.tasks[index]
	// containers
	containers := []string{}
	for _, c := range t.Containers {
		containers = append(containers, *c.Name)
	}
	// network
	subnetID := utils.EmptyText
	arch := utils.EmptyText
	privateIP := utils.EmptyText
	for _, a := range t.Attachments {
		if *a.Type == "ElasticNetworkInterface" {
			for _, d := range a.Details {
				if *d.Name == "subnetId" {
					subnetID = *d.Value
				}

				if *d.Name == "privateIPv4Address" {
					privateIP = *d.Value
				}
			}
		}
	}

	if len(t.Attributes) > 0 {
		for _, attribute := range t.Attributes {
			if *attribute.Name == "ecs.cpu-architecture" {
				arch = *attribute.Value
			}
		}
	}

	// Ephemeral storage
	ephemeralStorage := utils.EmptyText
	if t.EphemeralStorage != nil {
		ephemeralStorage = strconv.Itoa(int(t.EphemeralStorage.SizeInGiB)) + "Gb"
	}

	items = []headerItem{
		{name: "Task ID", value: utils.ArnToName(t.TaskArn)},
		{name: "Task definition", value: utils.ArnToName(t.TaskDefinitionArn)},
		{name: "Containers", value: strings.Join(containers, ",")},
		{name: "Cluster", value: utils.ArnToName(t.ClusterArn)},
		{name: "Launch type", value: string(t.LaunchType)},
		{name: "Capacity provider", value: utils.ShowString(t.CapacityProviderName)},
		{name: "Subnet ID", value: subnetID},
		{name: "Cpu architecture", value: arch},
		{name: "Private IP", value: privateIP},
		{name: "Ephemeral storage", value: ephemeralStorage},
		{name: "Execute command", value: strconv.FormatBool(t.EnableExecuteCommand)},
		{name: "Started by", value: utils.ShowString(t.StartedBy)},
		{name: "Started at", value: utils.ShowTime(t.StartedAt)},
		{name: "Pull started at", value: utils.ShowTime(t.PullStartedAt)},
		{name: "Pull stopped at", value: utils.ShowTime(t.PullStoppedAt)},
		{name: "Stopped reason", value: utils.ShowString(t.StoppedReason)},
		{name: "Stopped at", value: utils.ShowTime(t.StoppedAt)},
		{name: "Stopping at", value: utils.ShowTime(t.StoppingAt)},
		{name: "Platform family", value: utils.ShowString(t.PlatformFamily)},
		{name: "Platform version", value: utils.ShowString(t.PlatformVersion)},
		{name: "Tags count", value: strconv.Itoa(len(t.Tags))},
	}
	return
}

// Generate table params
func (v *taskView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	parent := *v.app.service.ServiceName
	if v.app.taskStatus == types.DesiredStatusStopped {
		parent = *v.app.cluster.ClusterName
	}
	title = fmt.Sprintf(color.TableTitleFmt, fmt.Sprintf("%s.%s", v.app.kind, strings.ToLower(string(v.app.taskStatus))), parent, len(v.tasks))
	headers = []string{
		"Task ID",
		"Last status",
		"Health",
		"Service",
		"Task definition",
		"Containers",
		"CPU",
		"Memory",
		"Age",
	}
	rowsBuilder = func() (data [][]string) {
		for _, t := range v.tasks {
			// healthy status
			health := string(t.HealthStatus)

			row := []string{}
			row = append(row, utils.ArnToName(t.TaskArn))
			row = append(row, utils.ShowGreenGrey(t.LastStatus, "running"))
			row = append(row, utils.ShowGreenGrey(&health, "healthy"))
			row = append(row, utils.GetServiceByTaskGroup(t.Group))
			row = append(row, utils.ArnToName(t.TaskDefinitionArn))
			row = append(row, strconv.Itoa(len(t.Containers)))
			row = append(row, utils.ShowString(t.Cpu))
			row = append(row, utils.ShowString(t.Memory))
			row = append(row, utils.Age(t.StartedAt))
			data = append(data, row)

			entity := Entity{task: &t, entityName: *t.TaskArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
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
