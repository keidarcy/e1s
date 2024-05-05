package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

const (
	// any form need at least one field
	placeholder = " (form placeholder) "
	// read only label for form
	readOnlyLabel = " [-:-:-] (read only) "
)

// Show modal with form
func (v *view) showFormModal(formContentFn func() (*tview.Form, *string), height int) {
	content, title := formContentFn()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(*title, ui.Modal(content, 100, height, v.closeModal), true, true)
}

// Show task definition register confirm modal
func (v *view) showTaskDefinitionConfirm(fn func()) {
	if v.app.kind != TaskDefinitionKind {
		return
	}
	content, title := v.taskDefinitionRegisterForm(fn)
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, ui.Modal(content, 100, 7, v.closeModal), true, true)
}

// Get task definition register content
func (v *view) taskDefinitionRegisterForm(fn func()) (*tview.Form, string) {
	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readOnlyLabel
	}

	title := " Register edited [purple::b]task definition?" + readOnly
	f := ui.StyledForm(title)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, title
	}

	// handle form submit
	f.AddButton("Register", func() {
		fn()
		v.closeModal()
	})
	return f, title
}

// Stop selected task
func (v *view) stopTaskForm() (*tview.Form, *string) {
	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readOnlyLabel
	}

	taskId := utils.ArnToName(v.app.task.TaskArn)
	clusterName := *v.app.cluster.ClusterName
	title := fmt.Sprintf(" Stop task [%s::b]%s[-:-:-] in [%s::b]%s[-:-:-] cluster %s? ", theme.Magenta, taskId, theme.Cyan, clusterName, readOnly)
	f := ui.StyledForm(title)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, &title
	}

	// handle form submit
	f.AddButton("Stop", func() {
		input := &ecs.StopTaskInput{
			Task:    aws.String(taskId),
			Cluster: aws.String(clusterName),
		}
		err := v.app.Store.StopTask(input)

		if err != nil {
			v.app.Notice.Error(err.Error())
			logger.Error(err.Error())
		} else {
			v.app.Notice.Infof("Task %s in %s cluster stopped", taskId, clusterName)
			logger.Infof("Task %s in %s cluster stopped", taskId, clusterName)
		}
		v.closeModal()
		v.showKindPage(TaskKind, true)
	})
	return f, &title
}

// Update service with selected task definition
func (v *view) serviceUpdateWithSpecificTaskDefinitionForm() (*tview.Form, *string) {
	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readOnlyLabel
	}

	if v.app.service == nil || v.app.taskDefinition == nil {
		logger.Warn("Unexpected nil to update service with task definition")
		return nil, nil
	}
	serviceName := *v.app.service.ServiceName
	td := utils.ArnToName(v.app.taskDefinition.TaskDefinitionArn)

	title := fmt.Sprintf(" Update [%s::b]%s[-:-:-] with task definition [%s::b]%s[-:-:-]%s? ", theme.Magenta, serviceName, theme.Cyan, td, readOnly)
	f := ui.StyledForm(title)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, &title
	}

	// handle form submit
	f.AddButton("Update", func() {
		input := &ecs.UpdateServiceInput{
			Service:        aws.String(serviceName),
			Cluster:        v.app.cluster.ClusterName,
			TaskDefinition: aws.String(td),
		}
		s, err := v.app.Store.UpdateService(input)

		if err != nil {
			v.app.Notice.Error(err.Error())
			logger.Error(err.Error())
		} else {
			v.app.Notice.Infof("Update service:\"%s\" with \"%d\" task definition:\"%s\" task(s)", *s.ServiceName, s.DesiredCount, utils.ArnToName(s.TaskDefinition))
			logger.Infof("Update service:\"%s\" with \"%d\" task definition:\"%s\" task(s)", *s.ServiceName, s.DesiredCount, utils.ArnToName(s.TaskDefinition))
		}
		v.closeModal()
		v.showKindPage(ServiceKind, true)
	})
	return f, &title
}

// Get service update form
func (v *view) serviceUpdateForm() (*tview.Form, *string) {
	const latest = "(LATEST)"

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, nil
	}
	name := *selected.service.ServiceName

	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readOnlyLabel
	}

	title := " Update [purple::b]" + name + readOnly
	currentFamily, currentRevision, _ := v.getTaskDefinitionDetail()

	// get data for form
	families, err := v.app.Store.ListTaskDefinitionFamilies()
	if err != nil {
		logger.Errorf("Failed list task definition families, err: %s", err.Error())
		v.app.Notice.Errorf("Failed list task definition families, err: %s", err.Error())
		v.closeModal()
	}

	f := ui.StyledForm(title)
	forceLabel := "Force new deployment"
	desiredLabel := "Desired tasks"
	familyLabel := "Task definition family"
	revisionLabel := "Task definition revision"

	f.AddCheckbox(forceLabel, false, nil)
	f.AddInputField(desiredLabel, strconv.Itoa(int(selected.service.DesiredCount)), 50, nil, nil)

	// If DeploymentController is CodeDeploy do not update task definition
	DeploymentControllerCodeDeploy := false
	if selected.service.DeploymentController.Type == types.DeploymentControllerTypeCodeDeploy {
		DeploymentControllerCodeDeploy = true
	}

	if !DeploymentControllerCodeDeploy {
		revisionDrop := tview.NewDropDown().
			SetLabel(revisionLabel).
			SetFieldWidth(50)

		currentFamilyIndex := 0
		for i, f := range families {
			if currentFamily == f {
				currentFamilyIndex = i
			}
		}

		familyDrop := tview.NewDropDown().
			SetLabel(familyLabel).
			SetOptions(families, func(text string, index int) {
				// when family option change, change revision drop down value
				taskDefinitions, err := v.app.Store.ListTaskDefinition(&text)
				if err != nil {
					logger.Errorf("Failed list task definition, err: %s", err.Error())
					v.app.Notice.Errorf("Failed list task definition, err: %s", err.Error())
					v.closeModal()
				}
				revisions := []string{}
				for i, td := range taskDefinitions {
					def := td
					_, revision := getTaskDefinitionInfo(&def)
					if i == 0 {
						revision += latest
					}
					revisions = append(revisions, revision)
				}
				revisionDrop.SetOptions(revisions, func(text string, index int) {})

				currentRevisionIndex := 0
				for i, r := range revisions {
					if currentRevision == r {
						currentRevisionIndex = i
					}
				}

				revisionDrop.SetCurrentOption(currentRevisionIndex)
			}).
			SetCurrentOption(currentFamilyIndex).
			SetFieldWidth(50)

		f.AddFormItem(familyDrop)
		f.AddFormItem(revisionDrop)
	}

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, &title
	}

	// handle form submit
	f.AddButton("Update", func() {
		var input *ecs.UpdateServiceInput
		var s *types.Service

		// get desired count
		desired := f.GetFormItemByLabel(desiredLabel).(*tview.InputField).GetText()
		desiredInt, err := strconv.Atoi(desired)
		if err != nil {
			return
		}
		// get force deploy bool
		force := f.GetFormItemByLabel(forceLabel).(*tview.Checkbox).IsChecked()

		if !DeploymentControllerCodeDeploy {
			// get task definition with revision
			_, selectedFamily := f.GetFormItemByLabel(familyLabel).(*tview.DropDown).GetCurrentOption()
			_, selectedRevision := f.GetFormItemByLabel(revisionLabel).(*tview.DropDown).GetCurrentOption()
			// if is latest cut suffix
			selectedRevision, _ = strings.CutSuffix(selectedRevision, latest)
			taskDefinitionWithRevision := selectedFamily + ":" + selectedRevision

			input = &ecs.UpdateServiceInput{
				Service:            aws.String(name),
				Cluster:            v.app.cluster.ClusterName,
				TaskDefinition:     aws.String(taskDefinitionWithRevision),
				DesiredCount:       aws.Int32(int32(desiredInt)),
				ForceNewDeployment: force,
			}
			s, err = v.app.Store.UpdateService(input)
		} else {
			input = &ecs.UpdateServiceInput{
				Service:            aws.String(name),
				Cluster:            v.app.cluster.ClusterName,
				DesiredCount:       aws.Int32(int32(desiredInt)),
				ForceNewDeployment: force,
			}
			s, err = v.app.Store.UpdateService(input)
		}

		if err != nil {
			v.closeModal()
			v.app.Notice.Error(err.Error())
			logger.Error(err.Error())
			v.reloadResource(false)
		} else {
			v.closeModal()

			// go v.app.QueueUpdateDraw(func() {
			// 	cell := v.table.GetCell(row, 3)
			// 	cell.SetText(strings.Replace(cell.Text,  "[green]completed[-:-:-]", "[grey]in_progress[-:-:-]", 1))
			// })

			// Update service last deployment
			go func() {
				row, _ := v.table.GetSelection()
				if row == 0 {
					row++
				}
				cell := v.table.GetCell(row, 4)
				cell.SetText(strings.Replace(cell.Text, "[green]Completed[-:-:-]", "[grey]In_progress[-:-:-]", 1))
				v.app.Application.Draw()
			}()

			v.app.Notice.Infof("Update service:\"%s\" with \"%d\" task definition:\"%s\" task(s)", *s.ServiceName, s.DesiredCount, utils.ArnToName(s.TaskDefinition))
			logger.Infof("Update service:\"%s\" with \"%d\" task definition:\"%s\" task(s)", *s.ServiceName, s.DesiredCount, utils.ArnToName(s.TaskDefinition))
			v.reloadResource(false)
		}
	})
	return f, &title
}

// Get service metrics charts
func (v *view) serviceMetricsForm() (*tview.Form, *string) {
	if v.app.kind != ServiceKind {
		return nil, nil
	}

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, nil
	}
	cluster := v.app.cluster.ClusterName
	service := *selected.service.ServiceName

	title := " Metrics [purple::b](" + service + ")" + readOnlyLabel

	f := ui.StyledForm(title)
	f.AddInputField("Service ", service+placeholder, len(service)+len(placeholder)+1, nil, nil)

	metrics, err := v.app.Store.GetMetrics(cluster, &service)

	// empty Metrics or empty
	if err != nil || (len(metrics.CPUUtilization) == 0 && len(metrics.MemoryUtilization) == 0) {
		return f, &title
	}
	if len(metrics.CPUUtilization) > 0 {
		cpuLabel := "CPUUtilization"
		f.AddTextView(cpuLabel, utils.BuildMeterText(*metrics.CPUUtilization[0].Average), 50, 1, true, false)
	}

	if len(metrics.MemoryUtilization) > 0 {
		memLabel := "MemoryUtilization"
		f.AddTextView(memLabel, utils.BuildMeterText(*metrics.MemoryUtilization[0].Average), 50, 1, true, false)
	}

	return f, &title
}
