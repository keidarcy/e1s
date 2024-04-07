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
			DescriptionKind:          descriptionPageKeys,
			LogKind:                  logPageKeys,
			TaskDefinitionDetailKind: descriptionPageKeys,
			TaskDefinitionKind:       descriptionPageKeys,
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
	items = []InfoItem{
		{name: "Task ID", value: util.ArnToName(t.TaskDefinitionArn)},
	}
	return
}

// Generate table params
func (v *TaskDefinitionView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, *v.app.service.ServiceName, len(v.taskDefinitions))
	headers = []string{
		"Task ID ▾",
	}
	dataBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			row := []string{}
			row = append(row, util.ArnToName(t.TaskDefinitionArn))
			data = append(data, row)
		}
		return data
	}

	return
}
