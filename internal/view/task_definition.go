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
	resources, err := app.Store.ListFullTaskDefinition(td)
	err = buildResourcePage(resources, app, err, func() resourceViewBuilder {
		return newTaskDefinitionView(resources, app)
	})

	return err
}

func (v *taskDefinitionView) getViewAndFooter() (*view, *tview.TextView) {
	return &v.view, v.footer.taskDefinition
}

// Build info pages for task page
func (v *taskDefinitionView) headerParamsBuilder() []headerPageParam {
	params := make([]headerPageParam, 0, len(v.taskDefinitions))
	for i, t := range v.taskDefinitions {
		params = append(params, headerPageParam{
			title:      utils.ArnToName(t.TaskDefinitionArn),
			entityName: *t.TaskDefinitionArn,
			items:      v.headerPageItems(i),
		})
	}
	return params
}

// Generate info pages params
func (v *taskDefinitionView) headerPageItems(index int) (items []headerItem) {
	t := v.taskDefinitions[index]
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
func (v *taskDefinitionView) tableParamsBuilder() (title string, headers []string, rowsBuilder func() [][]string) {
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
		"Revision",
		"Status",
		"In use",
		"CPU",
		"Memory",
		"Containers",
		"Age",
	}

	rowsBuilder = func() (data [][]string) {
		for _, t := range v.taskDefinitions {
			inUse := "-"
			if td == *t.TaskDefinitionArn {
				inUse = "Yes"
			}

			cpu, memory := taskDefinitionResources(t.Cpu, t.Memory, len(t.ContainerDefinitions), func(i int) (int32, *int32) {
				c := t.ContainerDefinitions[i]
				return c.Cpu, c.Memory
			})

			row := []string{}
			row = append(row, utils.ArnToName(t.TaskDefinitionArn))
			row = append(row, string(t.Status))
			row = append(row, utils.ShowGreenGrey(&inUse, "yes"))
			row = append(row, cpu)
			row = append(row, memory)
			row = append(row, strconv.Itoa(len(t.ContainerDefinitions)))
			row = append(row, utils.Age(t.RegisteredAt))
			data = append(data, row)

			entity := Entity{taskDefinition: &t, entityName: *t.TaskDefinitionArn}
			v.originalRowReferences = append(v.originalRowReferences, entity)
		}
		return data
	}

	return
}

func taskDefinitionResources(cpu *string, memory *string, containerCount int, containerResources func(int) (int32, *int32)) (string, string) {
	cpuValue := utils.ShowString(cpu)
	memoryValue := utils.ShowString(memory)
	if cpu != nil && memory != nil {
		return cpuValue, memoryValue
	}

	cpuSum := 0
	memorySum := 0
	for i := 0; i < containerCount; i++ {
		containerCPU, containerMemory := containerResources(i)
		cpuSum += int(containerCPU)
		if containerMemory != nil {
			memorySum += int(*containerMemory)
		}
	}

	if cpu == nil {
		cpuValue = strconv.Itoa(cpuSum)
	}
	if memory == nil {
		memoryValue = strconv.Itoa(memorySum)
	}
	return cpuValue, memoryValue
}
