package ui

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type TaskDefinitionView struct {
	View
	taskDefinitions []types.TaskDefinition
}

func newTaskDefinitionView(taskDefinitions []types.TaskDefinition, app *App) *TaskDefinitionView {
	keys := append(basicKeyInputs, []KeyInput{}...)
	return &TaskDefinitionView{
		View: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		taskDefinitions: taskDefinitions,
	}
}

func (app *App) showTaskDefinitionPage(reload bool) error {
	if switched := app.SwitchPage(reload); switched {
		return nil
	}

	taskDefinitions, err := app.Store.ListFullTaskDefinition(app.service.TaskDefinition)

	if err != nil {
		logger.Warnf("Failed to show taskDefinition pages, error: %v", err)
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
func (v *TaskDefinitionView) infoBuilder() *tview.Pages {
	for _, t := range v.taskDefinitions {
		title := util.ArnToName(t.TaskDefinitionArn)
		entityName := *t.TaskDefinitionArn
		items := v.infoPagesParam(t)

		v.buildInfoPages(items, title, entityName)
	}
	// prevent empty tasks
	if len(v.taskDefinitions) > 0 && v.taskDefinitions[0].TaskDefinitionArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.taskDefinitions[0].TaskDefinitionArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for task page
func (v *TaskDefinitionView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for task page
func (v *TaskDefinitionView) footerBuilder() *tview.Flex {
	v.footer.taskDefinition.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for task table
func (v *TaskDefinitionView) tableHandler() {
	for row, task := range v.taskDefinitions {
		t := task
		v.table.GetCell(row+1, 0).SetReference(Entity{taskDefinition: &t, entityName: *t.TaskDefinitionArn})
	}
}

// Generate info pages params
func (v *TaskDefinitionView) infoPagesParam(t types.TaskDefinition) (items []InfoItem) {
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

	items = []InfoItem{
		{name: "Revision", value: util.ArnToName(t.TaskDefinitionArn)},
		{name: "Task role", value: util.ShowString(t.TaskRoleArn)},
		{name: "Execution role", value: util.ShowString(t.ExecutionRoleArn)},
		{name: "Network mode", value: string(t.NetworkMode)},
		{name: "Volumes", value: util.ShowArray(volumes)},
		{name: "Containers", value: util.ShowArray(containers)},
		{name: "Placement constraints", value: util.ShowArray(placements)},
		{name: "Status", value: string(t.Status)},
		{name: "Compatibilities", value: util.ShowArray(compatibilities)},
		{name: "Requires Compatibilities", value: util.ShowArray(requiresCompatibilities)},
		{name: "Cpu", value: util.ShowString(t.Cpu)},
		{name: "Memory", value: util.ShowString(t.Memory)},
		{name: "Registered At", value: util.ShowTime(t.RegisteredAt)},
		{name: "Registered By", value: util.ShowString(t.RegisteredBy)},
	}
	return
}

// Generate table params
func (v *TaskDefinitionView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, *v.app.service.ServiceName, len(v.taskDefinitions))
	headers = []string{
		"Revision ▾",
		"In use",
		"Cpu",
		"Memory",
		"Age",
	}

	dataBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			inUse := "False"
			if *v.app.service.TaskDefinition == *t.TaskDefinitionArn {
				inUse = "True"
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
			row = append(row, util.ArnToName(t.TaskDefinitionArn))
			row = append(row, util.ShowGreenGrey(&inUse, "true"))
			row = append(row, cpu)
			row = append(row, memory)
			row = append(row, util.Age(t.RegisteredAt))
			data = append(data, row)
		}
		return data
	}

	return
}
