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

type taskDefinitionView struct {
	view
	taskDefinitions []types.TaskDefinition
}

func newTaskDefinitionView(taskDefinitions []types.TaskDefinition, app *App) *taskDefinitionView {
	keys := append(basicKeyInputs, []keyDescriptionPair{
		hotKeyMap["U"],
	}...)
	return &taskDefinitionView{
		view: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		taskDefinitions: taskDefinitions,
	}
}

func (app *App) showTaskDefinitionPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	td := app.service.TaskDefinition
	if td == nil {
		td = app.task.TaskDefinitionArn
	}
	taskDefinitions, err := app.Store.ListFullTaskDefinition(td)

	if err != nil {
		slog.Warn("failed to show taskDefinition pages", "error", err)
		app.back()
		return err
	}

	// no taskDefinition exists do nothing
	if len(taskDefinitions) == 0 {
		app.back()
		return fmt.Errorf("no valid task definition")
	}

	view := newTaskDefinitionView(taskDefinitions, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for task page
func (v *taskDefinitionView) headerBuilder() *tview.Pages {
	for _, t := range v.taskDefinitions {
		title := utils.ArnToName(t.TaskDefinitionArn)
		entityName := *t.TaskDefinitionArn
		items := v.headerPagesParam(t)

		v.buildHeaderPages(items, title, entityName)
	}
	// prevent empty tasks
	if len(v.taskDefinitions) > 0 && v.taskDefinitions[0].TaskDefinitionArn != nil {
		// show first when enter
		v.headerPages.SwitchToPage(*v.taskDefinitions[0].TaskDefinitionArn)
		v.changeSelectedValues()
	}
	return v.headerPages
}

// Build table for task page
func (v *taskDefinitionView) bodyBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.bodyPages
}

// Build footer for task page
func (v *taskDefinitionView) footerBuilder() *tview.Flex {
	v.footer.taskDefinition.SetText(fmt.Sprintf(color.FooterSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footerFlex
}

// Handlers for task table
func (v *taskDefinitionView) tableHandler() {
	for row, task := range v.taskDefinitions {
		t := task
		v.table.GetCell(row+1, 0).SetReference(Entity{taskDefinition: &t, entityName: *t.TaskDefinitionArn})
	}
}

// Generate info pages params
func (v *taskDefinitionView) headerPagesParam(t types.TaskDefinition) (items []headerItem) {
	compatibilities := []string{}
	for _, c := range t.Compatibilities {
		compatibilities = append(compatibilities, string(c))
	}

	requiresCompatibilities := []string{}
	for _, r := range t.RequiresCompatibilities {
		requiresCompatibilities = append(requiresCompatibilities, string(r))
	}

	volumes := []string{}
	for _, v := range t.Volumes {
		volumes = append(volumes, *v.Name)
	}

	containers := []string{}
	for _, c := range t.ContainerDefinitions {
		containers = append(containers, *c.Name)
	}

	placements := []string{}
	for _, p := range t.PlacementConstraints {
		placements = append(placements, string(p.Type))
	}

	arch := "x86_64"
	if t.RuntimePlatform != nil {
		arch = strings.ToLower(string(t.RuntimePlatform.CpuArchitecture))
	}

	OS := "linux"
	if t.RuntimePlatform != nil {
		OS = strings.ToLower(string(t.RuntimePlatform.OperatingSystemFamily))
	}

	items = []headerItem{
		{name: "Revision", value: utils.ArnToName(t.TaskDefinitionArn)},
		{name: "Task role", value: utils.ShowString(t.TaskRoleArn)},
		{name: "Execution role", value: utils.ShowString(t.ExecutionRoleArn)},
		{name: "Network mode", value: string(t.NetworkMode)},
		{name: "Volumes", value: utils.ShowArray(volumes)},
		{name: "Containers", value: utils.ShowArray(containers)},
		{name: "Placement constraints", value: utils.ShowArray(placements)},
		{name: "Status", value: string(t.Status)},
		{name: "Compatibilities", value: utils.ShowArray(compatibilities)},
		{name: "Requires compatibilities", value: utils.ShowArray(requiresCompatibilities)},
		{name: "CPU", value: utils.ShowString(t.Cpu)},
		{name: "Memory", value: utils.ShowString(t.Memory)},
		{name: "CPU architecture", value: arch},
		{name: "OS", value: OS},
		{name: "Registered At", value: utils.ShowTime(t.RegisteredAt)},
		{name: "Registered By", value: utils.ShowString(t.RegisteredBy)},
	}
	return
}

// Generate table params
func (v *taskDefinitionView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	serviceName, td := "", ""
	if v.app.service.ServiceName != nil {
		serviceName = *v.app.service.ServiceName
	}
	if v.app.service.TaskDefinition != nil {
		td = *v.app.service.TaskDefinition
	}
	if v.app.task.TaskDefinitionArn != nil {
		td = *v.app.task.TaskDefinitionArn
	}
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, serviceName, len(v.taskDefinitions))
	headers = []string{
		"Revision ▾",
		"In use",
		"CPU",
		"Memory",
		"Age",
	}

	dataBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			inUse := "-"
			if td == *t.TaskDefinitionArn {
				inUse = "Yes"
			}

			var cpu string
			if t.Cpu == nil {
				sum := 0
				for _, c := range t.ContainerDefinitions {
					sum += int(c.Cpu)
				}
				cpu = strconv.Itoa(sum)
			} else {
				cpu = *t.Cpu
			}

			var memory string
			if t.Memory == nil {
				sum := 0
				for _, c := range t.ContainerDefinitions {
					sum += int(*c.Memory)
				}
				memory = strconv.Itoa(sum)
			} else {
				memory = *t.Memory
			}

			row := []string{}
			row = append(row, utils.ArnToName(t.TaskDefinitionArn))
			row = append(row, utils.ShowGreenGrey(&inUse, "yes"))
			row = append(row, cpu)
			row = append(row, memory)
			row = append(row, utils.Age(t.RegisteredAt))
			data = append(data, row)
		}
		return data
	}

	return
}
