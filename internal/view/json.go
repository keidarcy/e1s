package view

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/utils"
)

// Switch to current kind description JSON page
func (v *view) switchToDescriptionJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	v.showJsonPages(selected)
}

// Get td family
func (v *view) getTaskDefinitionDetail() (string, string, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return "", "", ""
	}
	taskDefinition := ""
	entityName := ""
	if v.app.kind == ServiceKind {
		taskDefinition = *selected.service.TaskDefinition
		entityName = *selected.service.ServiceArn
	} else if v.app.kind == TaskKind {
		taskDefinition = *selected.task.TaskDefinitionArn
		entityName = *selected.task.TaskArn
	} else {
		return "", "", ""
	}
	familyRevision := strings.Split(utils.ArnToName(&taskDefinition), ":")
	return familyRevision[0], familyRevision[1], entityName
}

// Switch to auto scaling get by applicationautoscaling
func (v *view) switchToAutoScalingJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	serviceArn := selected.service.ServiceArn

	if serviceArn == nil {
		return
	}

	serviceFullName := utils.ArnToFullName(serviceArn)
	autoScaling, err := v.app.Store.GetAutoscaling(&serviceFullName)

	if err != nil {
		return
	}
	entity := Entity{autoScaling: autoScaling, entityName: *serviceArn}
	v.showJsonPages(entity)
}

// Switch to service revision
func (v *view) switchToServiceRevisionJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	serviceRevisionArn := selected.serviceDeployment.TargetServiceRevision.Arn

	if serviceRevisionArn == nil {
		return
	}

	serviceRevision, err := v.app.Store.GetServiceRevision(serviceRevisionArn)

	if err != nil {
		return
	}
	entity := Entity{serviceRevision: serviceRevision, entityName: *selected.serviceDeployment.ServiceDeploymentArn}
	v.showJsonPages(entity)
}

// Show new page from JSON content in table area and handle done event to go back
func (v *view) showJsonPages(entity Entity) {
	colorizedJsonString, rawJsonString, err := v.getJsonString(entity)
	if err != nil {
		return
	}
	v.handleSecondaryPageSwitch(entity, colorizedJsonString, rawJsonString)
	v.handleHeaderPageSwitch(entity)
}

func (v *view) handleFullScreenContentInput(jsonBytes []byte) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'f':
			pageName := v.app.kind.getAppPageName(v.app.getPageHandle())
			v.app.Pages.SwitchToPage(pageName)
		case 'e':
			if v.app.secondaryKind == DescriptionKind || v.app.secondaryKind == AutoScalingKind {
				v.openInEditor(jsonBytes)
			}
		}

		switch event.Key() {
		case tcell.KeyCtrlR:
			v.reloadResource(true)
		}
		return event
	}
}

func (v *view) handleTableContentDone(key tcell.Key) {
	pageName := v.app.kind.getTablePageName(v.app.getPageHandle())
	v.app.secondaryKind = EmptyKind

	slog.Debug("v.tablePages navigation", "action", "SwitchToPage", "pageName", pageName, "app", v.app)
	v.bodyPages.SwitchToPage(pageName)

	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.back()
	}

	slog.Debug("v.headerPages navigation", "action", "SwitchToPage", "pageName", pageName, "app", v.app)
	v.headerPages.SwitchToPage(selected.entityName)
}

func (v *view) handleFullScreenContentDone() {
	pageName := v.app.kind.getAppPageName(v.app.getPageHandle())
	v.app.Pages.SwitchToPage(pageName)
}

func (v *view) openInEditor(beforeJson []byte) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		slog.Warn("failed to get current selection", "error", err)
		return
	}
	names := strings.Split(selected.entityName, "/")

	// create tmp file open and defer close it
	tmpfile, err := os.CreateTemp("", names[len(names)-1])
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	if err != nil {
		v.app.Notice.Warnf("failed to read temporary file, err: %v", err)
		return
	}

	if _, err := tmpfile.Write(beforeJson); err != nil {
		v.app.Notice.Warnf("failed to write to temporary file, err: %v", err)
		return
	}

	// Open the vi editor to allow the user to modify the JSON data.
	bin := os.Getenv("EDITOR")
	if bin == "" {
		// if $EDITOR is empty use vi as default
		bin = "vi"
	}

	slog.Info("open", "bin", bin, "name", tmpfile.Name())
	v.app.Suspend(func() {
		v.app.isSuspended = true
		cmd := exec.Command(bin, tmpfile.Name())
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

		if err := cmd.Run(); err != nil {
			v.app.Notice.Warnf("failed to open editor, err: %v", err)
			return
		}

		afterJson, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			v.app.Notice.Warnf("failed to read temporary file, err: %v", err)
			return
		}

		// remove edited empty line
		if afterJson[len(afterJson)-1] == '\n' {
			beforeJson = append(beforeJson, '\n')
		}

		// if no change do nothing
		if bytes.Equal(beforeJson, afterJson) {
			if v.app.kind == TaskDefinitionKind {
				v.app.Notice.Info("JSON content has no change")
				slog.Info("JSON content has no change")
			}
			return
		}

		// if not task definition do nothing
		if v.app.kind != TaskDefinitionKind {
			v.app.Notice.Warn("Not support to update current resource")
			return
		}

		// only task definition and edited json register task definition
		var updatedTd ecs.RegisterTaskDefinitionInput
		if err := json.Unmarshal(afterJson, &updatedTd); err != nil {
			v.app.Notice.Warnf("failed to unmarshal JSON, err: %v", err)
			return
		}

		register := func() {
			family, revision, err := v.app.Store.RegisterTaskDefinition(&updatedTd)

			if err != nil {
				v.app.Notice.Warnf("failed to register new task definition, err: %v", err)
				return
			}
			v.app.Notice.Infof("Success TaskDefinition Family: %s, Revision: %d", family, revision)
		}

		v.showTaskDefinitionConfirm(register)
		v.app.isSuspended = false
	})
}

func (v *view) getJsonString(entity Entity) (string, []byte, error) {
	var data any

	switch {
	case entity.cluster != nil && v.app.kind == ClusterKind:
		data = entity.cluster
	case entity.instance != nil && v.app.kind == InstanceKind:
		data = entity.instance
	// events need be upper then service
	case entity.events != nil && v.app.secondaryKind == ServiceEventsKind:
		data = entity.events
	case entity.service != nil && v.app.kind == ServiceKind:
		data = entity.service
	case entity.serviceDeployment != nil && v.app.kind == ServiceDeploymentKind:
		data = entity.serviceDeployment
	case entity.task != nil && v.app.kind == TaskKind:
		data = entity.task
	case entity.container != nil && v.app.kind == ContainerKind:
		data = entity.container
	case entity.taskDefinition != nil && v.app.kind == TaskDefinitionKind:
		data = entity.taskDefinition
	case entity.metrics != nil:
		data = entity.metrics
	case entity.autoScaling != nil:
		data = entity.autoScaling
	case entity.serviceRevision != nil:
		data = entity.serviceRevision
	default:
		slog.Error("failed to get json string", "data", data)
		data = struct {
			Message     string
			IssueReport string
		}{
			Message:     "unknown issue",
			IssueReport: "https://github.com/keidarcy/e1s/issues",
		}
	}

	// get formatted json bytes
	jsonBytes, err := json.MarshalIndent(data, "", "  ")

	if err != nil {
		v.app.Notice.Warnf("failed to json marshal indent, error: %v", err)
		return "", []byte{}, err
	}

	return colorizeJSON(jsonBytes), jsonBytes, nil
}

func colorizeJSON(raw []byte) string {
	// key value to colorize json
	keyValRX := regexp.MustCompile(`(\s*)"(.*?)"\s*:\s*(.*)`)
	// Split the JSON into lines
	lines := strings.Split(string(raw), "\n")

	buff := make([]string, 0, len(lines))
	for _, l := range lines {
		res := keyValRX.FindStringSubmatch(l)
		if len(res) == 4 {
			buff = append(buff, fmt.Sprintf(`%s"[%s::b]%s[-:-:-]": %s`, res[1], theme.Blue, res[2], res[3]))
		} else {
			buff = append(buff, l)
		}
	}

	return strings.Join(buff, "\n")
}
