package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/keidarcy/e1s/api"
	"github.com/rivo/tview"
)

type PortForwardingSession struct {
	sessionId   *string
	port        string
	containerId string
}

// Show port forward modal and handle confirm event
func (v *View) showPortForwardingModal() {
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

			v.app.sessions = append(v.app.sessions, &PortForwardingSession{
				sessionId:   sessionId,
				port:        localPort,
				containerId: fmt.Sprintf("%s.%s", clusterName, name),
			})

			v.app.Notice.Infof("Port forwarding session started on %s", localPort)
			logger.Infof("Port forwarding session started on %s", localPort)

			v.reloadResource(false)
		}
	})
	return f, title
}

// Show task definition register confirm modal
func (v *View) showTerminatePortForwardingModal() {
	if v.app.kind != ContainerKind {
		return
	}
	content, title := v.terminatePortForwardingContent()
	if content == nil {
		return
	}
	v.app.Pages.AddPage(title, v.modal(content, 100, 6), true, true)
}

// Get task definition register content
func (v *View) terminatePortForwardingContent() (*tview.Form, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, ""
	}

	// container name
	name := *selected.container.Name
	clusterName := *v.app.cluster.ClusterName
	containerId := fmt.Sprintf("%s.%s", clusterName, name)
	ports := []string{}
	sessionIds := []*string{}
	indexes := []int{}
	for i, session := range v.app.sessions {
		if session.containerId == containerId {
			ports = append(ports, session.port)
			sessionIds = append(sessionIds, session.sessionId)
			indexes = append(indexes, i)
		}
	}

	if v.app.kind != ContainerKind {
		return nil, ""
	}

	readonly := ""
	if v.app.ReadOnly {
		readonly = readonlyLabel
	}

	title := fmt.Sprintf(" There is no port forwarding session on [purple::b]%s[-:-:-] ", name)

	portText := strings.Join(ports, ",")
	if len(ports) != 0 {
		title = fmt.Sprintf(" Terminate port forwarding session on [purple::b]%s[-:-:-] ? ", portText) + readonly
	}
	f := v.styledForm(title)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly || len(sessionIds) == 0 {
		return f, title
	}

	// handle form submit
	f.AddButton("Terminate", func() {
		// terminal targe container sessions
		err := v.app.Store.TerminateSessions(sessionIds)
		if err != nil {
			logger.Errorf("Failed to terminated port forwarding sessions err: %v", err)
		} else {
			v.app.Notice.Infof("Success terminated sessions on port %s", portText)
			logger.Debug("Terminated port forwarding session terminated")
		}

		v.closeModal()

		// remove app.sessions item for ui
		for i := len(indexes) - 1; i >= 0; i-- {
			indexToRemove := indexes[i]
			v.app.sessions = append(v.app.sessions[:indexToRemove], v.app.sessions[indexToRemove+1:]...)
		}
		v.reloadResource(false)
	})
	return f, title
}
