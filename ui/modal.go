package ui

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

// Show update service modal and handle submit event
func (v *View) showUpdateServiceModal() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceUpdateContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Show service auto scaling modal
func (v *View) showAutoScaling() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceAutoScalingContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 25), true, true)
}

// Show service metrics modal(Memory/CPU)
func (v *View) showMetrics() {
	if v.kind != ServicePage {
		return
	}
	content, title := v.serviceMetricsContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Get service auto scaling form
func (v *View) serviceAutoScalingContent() (*tview.Form, string) {
	if v.kind != ServicePage {
		return nil, ""
	}

	selected := v.getCurrentSelection()
	name := *selected.service.ServiceName

	readonly := "[-:-:-](readonly) "
	title := " Auto scaling [purple::b](" + name + ")" + readonly
	f := v.styledForm(title)
	f.AddInputField("Service ", name, len(name)+1, nil, nil)

	serviceArn := selected.service.ServiceArn

	if serviceArn == nil {
		f.AddTextView("No valid auto scaling configuration", "", 1, 1, false, false)
		return f, title
	}

	serviceFullName := util.ArnToFullName(serviceArn)
	autoScaling, err := v.app.Store.GetAutoScaling(&serviceFullName)
	// empty auto scaling or empty
	if err != nil || (len(autoScaling.Targets) == 0 && len(autoScaling.Policies) == 0 && len(autoScaling.Activities) == 0) {
		f.AddTextView("No valid auto scaling configuration", "", 10, 1, false, false)
		return f, title
	}

	if len(autoScaling.Targets) == 1 {
		minCountLabel := "Minimum number of tasks"
		maxCountLabel := "Maximum number of tasks"
		f.AddTextView(minCountLabel, strconv.Itoa(int(*autoScaling.Targets[0].MinCapacity)), 50, 1, true, false)
		f.AddTextView(maxCountLabel, strconv.Itoa(int(*autoScaling.Targets[0].MaxCapacity)), 50, 1, true, false)
	}

	if len(autoScaling.Policies) == 1 {
		policyNameLabel := "Policy name"
		metricNameLabel := "ECS service metric"
		targetValueLabel := "Target value"
		scaleOutPeriodLabel := "Scale-out cooldown period"
		scaleInPeriodLabel := "Scale-in cooldown period"
		noScaleInLabel := "Turn off scale-in"
		f.AddTextView(policyNameLabel, *autoScaling.Policies[0].PolicyName, 20, 1, true, false)
		f.AddTextView(metricNameLabel, string(autoScaling.Policies[0].TargetTrackingScalingPolicyConfiguration.PredefinedMetricSpecification.PredefinedMetricType), 50, 1, true, false)
		f.AddTextView(targetValueLabel, strconv.Itoa(int(*autoScaling.Policies[0].TargetTrackingScalingPolicyConfiguration.TargetValue)), 50, 1, true, false)
		f.AddTextView(scaleOutPeriodLabel, strconv.Itoa(int(*autoScaling.Policies[0].TargetTrackingScalingPolicyConfiguration.ScaleOutCooldown)), 50, 1, true, false)
		f.AddTextView(scaleInPeriodLabel, strconv.Itoa(int(*autoScaling.Policies[0].TargetTrackingScalingPolicyConfiguration.ScaleInCooldown)), 50, 1, true, false)
		f.AddTextView(noScaleInLabel, strconv.FormatBool(*autoScaling.Policies[0].TargetTrackingScalingPolicyConfiguration.DisableScaleIn), 50, 1, true, false)
	}

	return f, title
}

// Get service update form
func (v *View) serviceUpdateContent() (*tview.Form, string) {
	selected := v.getCurrentSelection()
	name := *selected.service.ServiceName

	readonly := ""
	if v.app.readonly {
		readonly = "[-:-:-](readonly) "
	}

	title := " Update [purple::b]" + name + " " + readonly
	family := v.getTaskDefinitionFamily()

	// get data for form
	taskDefinitions, err := v.app.Store.ListTaskDefinition(&family)
	if err != nil {
		v.closeModal()
	}
	revisions := []string{}
	for _, td := range taskDefinitions {
		def := td
		family, revision := getTaskDefinitionInfo(&def)
		revisions = append(revisions, family+":"+revision)
	}
	f := v.styledForm(title)

	forceLabel := "Force new deployment"
	desiredLabel := "Desired tasks"
	taskDefLabel := "Task definition revision"
	f.AddCheckbox(forceLabel, false, nil)
	f.AddInputField(desiredLabel, strconv.Itoa(int(selected.service.DesiredCount)), 10, nil, nil)
	f.AddDropDown(taskDefLabel, revisions, 30, nil)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.readonly {
		return f, title
	}

	// handle form submit
	f.AddButton("Update", func() {
		desired := f.GetFormItemByLabel(desiredLabel).(*tview.InputField).GetText()
		desiredInt, err := strconv.Atoi(desired)
		if err != nil {
			return
		}

		_, revision := f.GetFormItemByLabel(taskDefLabel).(*tview.DropDown).GetCurrentOption()
		force := f.GetFormItemByLabel(forceLabel).(*tview.Checkbox).IsChecked()
		input := &ecs.UpdateServiceInput{
			Service:            aws.String(name),
			Cluster:            v.app.cluster.ClusterName,
			TaskDefinition:     aws.String(revision),
			DesiredCount:       aws.Int32(int32(desiredInt)),
			ForceNewDeployment: force,
		}
		s, err := v.app.Store.UpdateService(input)

		if err != nil {
			v.closeModal()
			go v.errorModal(err.Error())
		} else {
			v.closeModal()
			go v.successModal(fmt.Sprintf("SUCCESS 🚀\nDesiredCount: %d\nTaskDefinition: %s\n", s.DesiredCount, *s.TaskDefinition))
		}
	})
	return f, title
}

// Get service metrics charts
func (v *View) serviceMetricsContent() (*tview.Form, string) {
	if v.kind != ServicePage {
		return nil, ""
	}

	selected := v.getCurrentSelection()
	cluster := v.app.cluster.ClusterName
	service := *selected.service.ServiceName

	title := " Metrics [purple::b](" + service + ")"

	f := v.styledForm(title)
	f.AddInputField("Service ", service, len(service)+1, nil, nil)

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
