package ui

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/api"
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
	if v.app.kind != ServicePage {
		return
	}
	content, title := v.serviceUpdateContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Show task definition register confirm modal
func (v *View) showTaskDefinitionConfirm(fn func()) {
	if v.app.kind != TaskPage {
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
	if v.app.kind != ServicePage {
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
	if v.app.kind != TaskPage {
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
		logger.Errorf("Failed list task definition families, err: %s", err.Error())
		v.app.Notice.Errorf("Failed list task definition families, err: %s", err.Error())
		v.closeModal()
	}

	f := v.styledForm(title)
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
		return f, title
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
			v.app.Notice.Infof("SUCCESS: DesiredCount: %d, TaskDefinition: %s", s.DesiredCount, *s.TaskDefinition)
			logger.Infof("SUCCESS: DesiredCount: %d, TaskDefinition: %s", s.DesiredCount, *s.TaskDefinition)
			v.reloadResource(false)
		}
	})
	return f, title
}

// Get service metrics charts
func (v *View) serviceMetricsContent() (*tview.Form, string) {
	if v.app.kind != ServicePage {
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

// Show port forward modal and handle confirm event
func (v *View) showPortForwardingModal() {
	if v.app.kind != ContainerPage {
		return
	}
	content, title := v.portForwardingForm()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 15), true, true)
}

// Get port forward form content
func (v *View) portForwardingForm() (*tview.Form, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}
	// container name
	name := *selected.container.Name

	placeHolderPort := "8080"
	placeHolderLocalPort := "8080"

	td, err := v.app.Store.DescribeTaskDefinition(v.app.task.TaskDefinitionArn)
	if err != nil {
		return nil, ""
	}

	for _, c := range td.ContainerDefinitions {
		if name == *c.Name {
			if len(c.PortMappings) > 0 {
				if p := c.PortMappings[0].ContainerPort; p != nil {
					placeHolderPort = strconv.Itoa(int((*p)))
					placeHolderLocalPort = strconv.Itoa(int((*p)))
				}
			}
			break
		}
	}

	readOnly := ""
	if v.app.ReadOnly {
		readOnly = readonlyLabel
	}

	title := " Port Forward [purple::b]" + name + readOnly

	f := v.styledForm(title)
	remoteForwardLabel := "Remote Forward"
	hostLabel := "Host"
	portLabel := "Port number"
	localPortLabel := "Local port number"

	f.AddCheckbox(remoteForwardLabel, false, nil)
	f.AddInputField(hostLabel, "", 50, nil, nil)
	f.AddInputField(portLabel, placeHolderPort, 50, nil, nil)
	f.AddInputField(localPortLabel, placeHolderLocalPort, 50, nil, nil)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly {
		return f, title
	}

	// handle form submit
	f.AddButton("Start", func() {
		taskArn := strings.Split(*v.app.task.TaskArn, "/")
		clusterName := taskArn[1]
		taskId := taskArn[2]
		runtimeId := *selected.container.RuntimeId

		remoteHost := f.GetFormItemByLabel(remoteForwardLabel).(*tview.Checkbox).IsChecked()
		host := f.GetFormItemByLabel(hostLabel).(*tview.InputField).GetText()
		port := f.GetFormItemByLabel(portLabel).(*tview.InputField).GetText()
		localPort := f.GetFormItemByLabel(localPortLabel).(*tview.InputField).GetText()

		sessionId, err := v.app.Store.StartSession(&api.SsmStartSessionInput{
			ClusterName: clusterName,
			Host:        host,
			TaskId:      taskId,
			RuntimeId:   runtimeId,
			RemoteHost:  remoteHost,
			Port:        port,
			LocalPort:   localPort,
		})

		if err != nil {
			v.closeModal()

			v.app.Notice.Error(err.Error())
			logger.Error(err.Error())
		} else {
			v.closeModal()

			if remoteHost {
				v.app.Notice.Infof("Remote host port forwarding, host: %s, port: %s, local port: %s", host, port, localPort)
				logger.Infof("Remote host port forwarding, host: %s, port: %s, local port: %s", host, port, localPort)
			}
			v.app.sessionIds = append(v.app.sessionIds, &sessionId)
			v.app.Notice.Infof("Port forwarding, port: %s, local port: %s", port, localPort)
			logger.Infof("Port forwarding, port: %s, local port: %s", port, localPort)

			// Update container table associated row PF text
			row, _ := v.table.GetSelection()
			if row == 0 {
				row++
			}
			go v.app.QueueUpdateDraw(func() {
				v.table.GetCell(row, 3).SetText("[green::i]F[-:-:-]")
			})
		}
	})
	return f, title
}
