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

type daemonTaskDefinitionView struct {
	view
	taskDefinitions []types.DaemonTaskDefinition
}

func newDaemonTaskDefinitionView(taskDefinitions []types.DaemonTaskDefinition, app *App) *daemonTaskDefinitionView {
	return &daemonTaskDefinitionView{
		view: *newView(app, basicKeyInputs, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		taskDefinitions: taskDefinitions,
	}
}

func (app *App) showDaemonTaskDefinitionPage(reload bool) error {
	if switched := app.switchPage(reload); switched {
		return nil
	}

	var family *string

	// Coming from a daemon task (via task view with group "daemon:...")
	if app.task != nil && app.task.TaskDefinitionArn != nil && app.task.Group != nil && strings.HasPrefix(*app.task.Group, "daemon:") {
		name := utils.ArnToName(app.task.TaskDefinitionArn)
		parts := strings.Split(name, ":")
		family = &parts[0]
	} else if app.daemonSummary != nil && app.daemonSummary.DaemonArn != nil {
		// Coming from the daemon view — resolve via DescribeDaemon + DescribeDaemonRevisions
		detail, err := app.Store.DescribeDaemon(app.daemonSummary.DaemonArn)
		if err == nil && detail != nil && len(detail.CurrentRevisions) > 0 {
			revOutput, revErr := app.Store.DescribeDaemonRevisions([]string{*detail.CurrentRevisions[0].Arn})
			if revErr == nil && len(revOutput) > 0 && revOutput[0].DaemonTaskDefinitionArn != nil {
				name := utils.ArnToName(revOutput[0].DaemonTaskDefinitionArn)
				parts := strings.Split(name, ":")
				family = &parts[0]
			}
		}
	}

	if family == nil {
		return fmt.Errorf("no daemon task definition found")
	}

	resources, err := app.Store.ListDaemonTaskDefinitions(family)
	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newDaemonTaskDefinitionView(resources, app)
	})
	return err
}

func (v *daemonTaskDefinitionView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.daemonTaskDefinition
}

func (v *daemonTaskDefinitionView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.taskDefinitions))
	for i, t := range v.taskDefinitions {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(t.DaemonTaskDefinitionArn),
			entityName: *t.DaemonTaskDefinitionArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

func (v *daemonTaskDefinitionView) headerPageItems(index int) (items []headerItem) {
	t := v.taskDefinitions[index]

	containers := []string{}
	for _, c := range t.ContainerDefinitions {
		if c.Name != nil {
			containers = append(containers, *c.Name)
		}
	}

	items = []headerItem{
		{name: "Family", value: utils.ShowString(t.Family)},
		{name: "Revision", value: strconv.Itoa(int(t.Revision))},
		{name: "Status", value: string(t.Status)},
		{name: "CPU", value: utils.ShowString(t.Cpu)},
		{name: "Memory", value: utils.ShowString(t.Memory)},
		{name: "Containers", value: strings.Join(containers, ",")},
		{name: "Task role", value: utils.ShowString(t.TaskRoleArn)},
		{name: "Execution role", value: utils.ShowString(t.ExecutionRoleArn)},
		{name: "PID mode", value: string(t.PidMode)},
		{name: "IPC mode", value: string(t.IpcMode)},
		{name: "Registered at", value: utils.ShowTime(t.RegisteredAt)},
		{name: "Registered by", value: utils.ShowString(t.RegisteredBy)},
	}
	return
}

func (v *daemonTaskDefinitionView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
	parent := ""
	if v.app.daemonSummary != nil && v.app.daemonSummary.DaemonArn != nil {
		parent = utils.ArnToName(v.app.daemonSummary.DaemonArn)
	} else if v.app.task != nil && v.app.task.TaskDefinitionArn != nil {
		parent = utils.ArnToName(v.app.task.TaskDefinitionArn)
	}
	title = fmt.Sprintf(color.TableTitleFmt, v.app.kind, parent, len(v.taskDefinitions))
	headers = []string{
		"Revision",
		"Status",
		"CPU",
		"Memory",
		"Containers",
		"Age",
	}

	rowsBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			status := string(t.Status)
			cpu := utils.ShowString(t.Cpu)
			memory := utils.ShowString(t.Memory)

			row := []string{
				fmt.Sprintf("%s:%d", utils.ShowString(t.Family), t.Revision),
				utils.ShowGreenGrey(&status, "active"),
				cpu,
				memory,
				strconv.Itoa(len(t.ContainerDefinitions)),
				utils.Age(t.RegisteredAt),
			}
			data = append(data, row)

			entity := Entity{daemonTaskDefinition: &t, entityName: *t.DaemonTaskDefinitionArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}
	return
}
