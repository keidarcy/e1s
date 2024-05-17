package view

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/keidarcy/e1s/internal/api"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type PortForwardingSession struct {
	sessionId   *string
	port        string
	containerId string
}

// Get port forward form content
func (v *view) portForwardingForm() (*tview.Form, *string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, nil
	}
	// container name
	name := *selected.container.Name

	placeHolderPort := "8080"
	placeHolderLocalPort := "8080"

	td, err := v.app.Store.DescribeTaskDefinition(v.app.task.TaskDefinitionArn)
	if err != nil {
		return nil, nil
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
		readOnly = readOnlyLabel
	}

	title := " Port Forward [purple::b]" + name + readOnly

	f := ui.StyledForm(title)
	remoteForwardLabel := "Remote forward"
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
		return f, &title
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
			slog.Error(err.Error())
		} else {
			v.closeModal()

			v.app.sessions = append(v.app.sessions, &PortForwardingSession{
				sessionId:   sessionId,
				port:        localPort,
				containerId: fmt.Sprintf("%s.%s", clusterName, name),
			})

			v.app.Notice.Infof("port forwarding session started on %s", localPort)

			// Update port
			go func() {
				row, _ := v.table.GetSelection()
				if row == 0 {
					row++
				}
				cell := v.table.GetCell(row, 3)
				text := cell.Text
				if text == utils.EmptyText {
					text = localPort
				} else {
					text = fmt.Sprintf("%s,%s", text, localPort)
				}
				cell.SetText(text)
				v.app.Application.Draw()
			}()

			v.reloadResource(false)
		}
	})
	return f, &title
}

// Get task definition register content
func (v *view) terminatePortForwardingForm() (*tview.Form, *string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return nil, nil
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
		return nil, nil
	}

	readonly := ""
	if v.app.ReadOnly {
		readonly = readOnlyLabel
	}

	title := fmt.Sprintf(" There is no port forwarding session on [purple::b]%s[-:-:-] ", name)

	portText := strings.Join(ports, ",")
	if len(ports) != 0 {
		title = fmt.Sprintf(" Terminate port forwarding session on [purple::b]%s[-:-:-] ? ", portText) + readonly
	}
	f := ui.StyledForm(title)

	// handle form close
	f.AddButton("Cancel", func() {
		v.closeModal()
	})

	// readonly mode has no submit button
	if v.app.ReadOnly || len(sessionIds) == 0 {
		return f, &title
	}

	// handle form submit
	f.AddButton("Terminate", func() {
		// terminal targe container sessions
		err := v.app.Store.TerminateSessions(sessionIds)
		if err != nil {
			slog.Error("failed to terminated port forwarding sessions", "error", err)
		} else {
			v.app.Notice.Infof("success terminated sessions on port %s", portText)
		}

		// Update port
		go func() {
			row, _ := v.table.GetSelection()
			if row == 0 {
				row++
			}
			cell := v.table.GetCell(row, 3)
			cell.SetText(utils.EmptyText)
			v.app.Application.Draw()
		}()

		v.closeModal()

		// remove app.sessions item for ui
		for i := len(indexes) - 1; i >= 0; i-- {
			indexToRemove := indexes[i]
			v.app.sessions = append(v.app.sessions[:indexToRemove], v.app.sessions[indexToRemove+1:]...)
		}
		v.reloadResource(false)
	})
	return f, &title
}
