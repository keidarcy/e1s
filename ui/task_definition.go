package ui

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

type TaskDefinitionView struct {
	View
	taskDefinitions []types.TaskDefinition
}

func newTaskDefinitionView(taskDefinitions []types.TaskDefinition, app *App) *TaskDefinitionView {
	keys := append(basicKeyInputs, []KeyInput{
		{key: string(tKey), description: describeTaskDefinition},
		{key: string(vKey), description: describeTaskDefinitionRevisions},
		{key: string(eKey), description: editTaskDefinition},
		{key: string(lKey), description: showLogs},
	}...)
	return &TaskDefinitionView{
		View: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: descriptionPageKeys,
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

	items = []InfoItem{
		{name: "Revision", value: util.ArnToName(t.TaskDefinitionArn)},
		{name: "Task Role", value: util.ShowString(t.TaskRoleArn)},
		{name: "Execution Role", value: util.ShowString(t.ExecutionRoleArn)},
		{name: "Network Mode", value: string(t.NetworkMode)},
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
		"In Use",
		"Cpu",
		"Memory",
		"Registered At",
	}

	dataBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			inUse := "False"
			if *v.app.service.TaskDefinition == *t.TaskDefinitionArn {
				inUse = "True"
			}

			row := []string{}
			row = append(row, util.ArnToName(t.TaskDefinitionArn))
			row = append(row, util.ShowGreenGrey(&inUse, "true"))
			row = append(row, util.ShowString(t.Cpu))
			row = append(row, util.ShowString(t.Memory))
			row = append(row, util.ShowTime(t.RegisteredAt))
			data = append(data, row)
		}
		return data
	}

	return
}
