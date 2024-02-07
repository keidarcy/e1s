package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

const (
	// any form need at least one field
	placeholder = " (form placeholder) "
	// readonly label for form
	readonlyLabel = " [-:-:-] (readonly) "
)

// Show update service modal and handle submit event
func (v *View) showEditServiceModal() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceUpdateContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Deprecated
// Autoscaling is too complex to show in a form
// Show service auto scaling modal
func (v *View) showAutoScalingModal() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceAutoScalingContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 35), true, true)
}

// Show task definition register confirm modal
func (v *View) showTaskDefinitionConfirm(fn func()) {
	if v.kind != TaskPage {
		return
	}
	content, title := v.taskDefinitionRegisterContent(fn)
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 10), true, true)
}

// Show service metrics modal(Memory/CPU)
func (v *View) showMetricsModal() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceMetricsContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Get task definition register content
func (v *View) taskDefinitionRegisterContent(fn func()) (*tview.Form, string) {
	if v.kind != TaskPage {
		return nil, ""
	}

	readonly := ""
	if v.app.ReadOnly {
		readonly = readonlyLabel
	}

	title := " Register edited [purple::b]task definition?" + readonly
	f := v.styledForm(title)

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
	})
	return f, title
}

// Deprecated
// Autoscaling is too complex to show in a form
// Get service auto scaling form
// Show all three type autoscaling https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-auto-scaling.html
// 1. Target tracking scaling policies
// 2. Step scaling policies
// 3. Scheduled Scaling
func (v *View) serviceAutoScalingContent() (*tview.Form, string) {
	if v.kind != ServicePage {
		return nil, ""
	}

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}
	name := *selected.service.ServiceName

	title := " Auto scaling [purple::b](" + name + ")" + readonlyLabel
	f := v.styledForm(title)
	f.AddInputField("Service ", name+placeholder, len(name)+len(placeholder)+1, nil, nil)

	serviceArn := selected.service.ServiceArn

	if serviceArn == nil {
		f.AddTextView("No valid auto scaling configuration", "", 1, 1, false, false)
		return f, title
	}

	serviceFullName := util.ArnToFullName(serviceArn)
	autoscaling, err := v.app.Store.GetAutoscaling(&serviceFullName)
	// empty auto scaling or empty
	if err != nil || (len(autoscaling.Targets) == 0 && len(autoscaling.Policies) == 0 && len(autoscaling.Activities) == 0) {
		f.AddTextView("No valid auto scaling configuration", "", 10, 1, false, false)
		return f, title
	}

	autoscalingTypeLabel := "Autoscaling Type"
	minCountLabel := "Minimum number of tasks"
	maxCountLabel := "Maximum number of tasks"

	if len(autoscaling.Actions) > 0 {
		// autoscaling type: Scheduled Scaling
		f.AddTextView(autoscalingTypeLabel, "Scheduled Scaling", 50, 1, true, false)
		for _, action := range autoscaling.Actions {
			scheduledActionNameLabel := "Action name"
			scheduleLabel := "Schedule"
			timezoneLabel := "Timezone"
			// lineLabel := "------"
			f.AddTextView(scheduledActionNameLabel, string(*action.ScheduledActionName), 50, 1, true, false)
			f.AddTextView(scheduleLabel, string(*action.Schedule), 50, 1, true, false)
			f.AddTextView(timezoneLabel, string(*action.Timezone), 50, 1, true, false)
			f.AddTextView(minCountLabel, strconv.Itoa(int(*action.ScalableTargetAction.MinCapacity)), 50, 1, true, false)
			f.AddTextView(maxCountLabel, strconv.Itoa(int(*action.ScalableTargetAction.MaxCapacity)), 50, 1, true, false)
			// f.AddTextView(lineLabel, "", 50, 1, true, false)
		}
	}

	if len(autoscaling.Policies) == 1 {
		f.AddTextView(autoscalingTypeLabel, "Target tracking scaling policies", 50, 1, true, false)

		// autoscaling type: Target tracking scaling policies
		if len(autoscaling.Targets) == 1 {
			f.AddTextView(minCountLabel, strconv.Itoa(int(*autoscaling.Targets[0].MinCapacity)), 50, 1, true, false)
			f.AddTextView(maxCountLabel, strconv.Itoa(int(*autoscaling.Targets[0].MaxCapacity)), 50, 1, true, false)
		}

		policyNameLabel := "Policy name"
		metricNameLabel := "ECS service metric"
		targetValueLabel := "Target value"
		scaleOutPeriodLabel := "Scale-out cooldown period"
		scaleInPeriodLabel := "Scale-in cooldown period"
		noScaleInLabel := "Turn off scale-in"
		f.AddTextView(policyNameLabel, *autoscaling.Policies[0].PolicyName, 20, 1, true, false)
		f.AddTextView(metricNameLabel, string(autoscaling.Policies[0].TargetTrackingScalingPolicyConfiguration.PredefinedMetricSpecification.PredefinedMetricType), 50, 1, true, false)
		f.AddTextView(targetValueLabel, strconv.Itoa(int(*autoscaling.Policies[0].TargetTrackingScalingPolicyConfiguration.TargetValue)), 50, 1, true, false)
		f.AddTextView(scaleOutPeriodLabel, strconv.Itoa(int(*autoscaling.Policies[0].TargetTrackingScalingPolicyConfiguration.ScaleOutCooldown)), 50, 1, true, false)
		f.AddTextView(scaleInPeriodLabel, strconv.Itoa(int(*autoscaling.Policies[0].TargetTrackingScalingPolicyConfiguration.ScaleInCooldown)), 50, 1, true, false)
		f.AddTextView(noScaleInLabel, strconv.FormatBool(*autoscaling.Policies[0].TargetTrackingScalingPolicyConfiguration.DisableScaleIn), 50, 1, true, false)
	}

	if len(autoscaling.Policies) > 1 {
		f.AddTextView(autoscalingTypeLabel, "Step scaling policies", 50, 1, true, false)

		if len(autoscaling.Targets) == 1 {
			f.AddTextView(minCountLabel, strconv.Itoa(int(*autoscaling.Targets[0].MinCapacity)), 50, 1, true, false)
			f.AddTextView(maxCountLabel, strconv.Itoa(int(*autoscaling.Targets[0].MaxCapacity)), 50, 1, true, false)
		}

		for _, policy := range autoscaling.Policies {
			policyNameLabel := "Policy name"
			f.AddTextView(policyNameLabel, *policy.PolicyName, 20, 1, true, false)
			for _, alarms := range policy.Alarms {
				alarmLabel := "Alarm name"
				f.AddTextView(alarmLabel, *alarms.AlarmName, 20, 1, true, false)
			}
		}

	}

	return f, title
}

// Get service update form
func (v *View) serviceUpdateContent() (*tview.Form, string) {
	const latest = "(LATEST)"

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}
	name := *selected.service.ServiceName

	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readonlyLabel
	}

	title := " Update [purple::b]" + name + readOnly
	currentFamily, currentRevision, _ := v.getTaskDefinitionDetail()

	// get data for form
	families, err := v.app.Store.ListTaskDefinitionFamilies()
	if err != nil {
		v.errorModal("aws api error!", 2, 20, 10)
		v.closeModal()
	}

	f := v.styledForm(title)
	forceLabel := "Force new deployment"
	desiredLabel := "Desired tasks"
	familyLabel := "Task definition family"
	revisionLabel := "Task definition revision"

	f.AddCheckbox(forceLabel, false, nil)
	f.AddInputField(desiredLabel, strconv.Itoa(int(selected.service.DesiredCount)), 50, nil, nil)

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
				v.errorModal("aws api error!", 2, 20, 10)
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

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, title
	}

	// handle form submit
	f.AddButton("Update", func() {
		// get desired count
		desired := f.GetFormItemByLabel(desiredLabel).(*tview.InputField).GetText()
		desiredInt, err := strconv.Atoi(desired)
		if err != nil {
			return
		}

		// get task definition with revision
		_, selectedFamily := f.GetFormItemByLabel(familyLabel).(*tview.DropDown).GetCurrentOption()
		_, selectedRevision := f.GetFormItemByLabel(revisionLabel).(*tview.DropDown).GetCurrentOption()
		// if is latest cut suffix
		selectedRevision, _ = strings.CutSuffix(selectedRevision, latest)
		taskDefinitionWithRevision := selectedFamily + ":" + selectedRevision

		// get force deploy bool
		force := f.GetFormItemByLabel(forceLabel).(*tview.Checkbox).IsChecked()
		input := &ecs.UpdateServiceInput{
			Service:            aws.String(name),
			Cluster:            v.app.cluster.ClusterName,
			TaskDefinition:     aws.String(taskDefinitionWithRevision),
			DesiredCount:       aws.Int32(int32(desiredInt)),
			ForceNewDeployment: force,
		}
		s, err := v.app.Store.UpdateService(input)

		if err != nil {
			v.closeModal()
			go v.errorModal(err.Error(), 5, 100, 10)
			v.reloadResource()
		} else {
			v.closeModal()
			go v.successModal(fmt.Sprintf("SUCCESS 🚀\nDesiredCount: %d\nTaskDefinition: %s\n", s.DesiredCount, *s.TaskDefinition), 5, 110, 5)
			v.reloadResource()
		}
	})
	return f, title
}

// Get service metrics charts
func (v *View) serviceMetricsContent() (*tview.Form, string) {
	if v.kind != ServicePage {
		return nil, ""
	}

	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}
	cluster := v.app.cluster.ClusterName
	service := *selected.service.ServiceName

	title := " Metrics [purple::b](" + service + ")" + readonlyLabel

	f := v.styledForm(title)
	f.AddInputField("Service ", service+placeholder, len(service)+len(placeholder)+1, nil, nil)

	metrics, err := v.app.Store.GetMetrics(cluster, &service)

	// empty Metrics or empty
	if err != nil || (len(metrics.CPUUtilization) == 0 && len(metrics.MemoryUtilization) == 0) {
		return f, title
	}
	if len(metrics.CPUUtilization) > 0 {
		cpuLabel := "CPUUtilization"
		f.AddTextView(cpuLabel, util.BuildMeterText(*metrics.CPUUtilization[0].Average), 50, 1, true, false)
	}

	if len(metrics.MemoryUtilization) > 0 {
		memLabel := "MemoryUtilization"
		f.AddTextView(memLabel, util.BuildMeterText(*metrics.MemoryUtilization[0].Average), 50, 1, true, false)
	}

	return f, title
}
