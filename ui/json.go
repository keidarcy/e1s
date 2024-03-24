package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
	"github.com/sirupsen/logrus"
)

// Switch to current kind description JSON page
func (v *View) switchToDescriptionJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	v.showJsonPages(selected)
}

// Switch to selected task definition JSON page
func (v *View) switchToTaskDefinitionJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	taskDefinition := ""
	entityName := ""
	if v.app.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
		entityName = *selected.service.ServiceArn
	} else if v.app.kind == TaskPage {
		taskDefinition = *selected.task.TaskDefinitionArn
		entityName = *selected.task.TaskArn
	} else {
		return
	}

	td, err := v.app.Store.DescribeTaskDefinition(&taskDefinition)
	if err != nil {
		return
	}
	entity := Entity{taskDefinition: &td, entityName: entityName}
	v.showJsonPages(entity)
}

// Switch to selected task definition revision list JSON page
func (v *View) switchToTaskDefinitionRevisionsJson() {
	if v.app.kind == ClusterPage {
		return
	}
	family, _, entityName := v.getTaskDefinitionDetail()

	revisions, err := v.app.Store.ListTaskDefinition(&family)
	if err != nil {
		return
	}
	entity := Entity{taskDefinitionRevisions: revisions, entityName: entityName}
	v.showJsonPages(entity)
}

// Get td family
func (v *View) getTaskDefinitionDetail() (string, string, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return "", "", ""
	}
	taskDefinition := ""
	entityName := ""
	if v.app.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
		entityName = *selected.service.ServiceArn
	} else if v.app.kind == TaskPage {
		taskDefinition = *selected.task.TaskDefinitionArn
		entityName = *selected.task.TaskArn
	} else {
		return "", "", ""
	}
	familyRevision := strings.Split(util.ArnToName(&taskDefinition), ":")
	return familyRevision[0], familyRevision[1], entityName
}

// Deprecated
// not called anywhere
// Switch to auto scaling get by applicationautoscaling
func (v *View) switchToAutoScalingJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	if v.app.kind != ServicePage {
		return
	}
	serviceArn := selected.service.ServiceArn

	if serviceArn == nil {
		return
	}

	serviceFullName := util.ArnToFullName(serviceArn)
	autoScaling, err := v.app.Store.GetAutoscaling(&serviceFullName)

	if err != nil {
		return
	}
	entity := Entity{autoScaling: autoScaling, entityName: *serviceArn}
	v.showJsonPages(entity)
}

// Show new page from JSON content in table area and handle done event to go back
func (v *View) showJsonPages(entity Entity) {
	contentString := v.getJsonString(entity)
	v.handleContentPageSwitch(entity, contentString)
	v.handleInfoPageSwitch(entity)
}

func (v *View) handleFullScreenContentInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case fKey:
		pageName := v.app.kind.getAppPageName(v.app.getPageHandle())
		v.app.Pages.SwitchToPage(pageName)
	}

	switch event.Key() {
	case tcell.KeyCtrlR:
		v.reloadResource(true)
	}
	return event
}

func (v *View) handleTableContentDone(key tcell.Key) {
	pageName := v.app.kind.getTablePageName(v.app.getPageHandle())
	v.app.secondaryKind = EmptyKind

	logger.WithFields(logrus.Fields{
		"Action":        "SwitchToPage",
		"PageName":      pageName,
		"Kind":          v.app.kind.String(),
		"SecondaryKind": v.app.secondaryKind.String(),
		"Cluster":       *v.app.cluster.ClusterName,
		"Service":       *v.app.service.ServiceName,
	}).Debug("SwitchToPage v.tablePages")

	v.tablePages.SwitchToPage(pageName)

	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.back()
	}

	logger.WithFields(logrus.Fields{
		"Action":        "SwitchToPage",
		"PageName":      selected.entityName,
		"Kind":          v.app.kind.String(),
		"SecondaryKind": v.app.secondaryKind.String(),
		"Cluster":       *v.app.cluster.ClusterName,
		"Service":       *v.app.service.ServiceName,
	}).Debug("SwitchToPage v.infoPages")

	v.infoPages.SwitchToPage(selected.entityName)
}

func (v *View) handleFullScreenContentDone(key tcell.Key) {
	pageName := v.app.kind.getAppPageName(v.app.getPageHandle())
	v.app.Pages.SwitchToPage(pageName)
}

func (v *View) getJsonString(entity Entity) string {
	var data any

	switch {
	case entity.cluster != nil && v.app.kind == ClusterPage:
		data = entity.cluster
	// events need be upper then service
	case entity.events != nil && v.app.secondaryKind == ServiceEventsPage:
		data = entity.events
	case entity.service != nil && v.app.kind == ServicePage:
		data = entity.service
	case entity.task != nil && v.app.kind == TaskPage:
		data = entity.task
	case entity.container != nil && v.app.kind == ContainerPage:
		data = entity.container
	case entity.taskDefinition != nil && v.app.secondaryKind == TaskDefinitionPage:
		data = entity.taskDefinition
	case entity.taskDefinitionRevisions != nil && v.app.secondaryKind == TaskDefinitionRevisionsPage:
		data = entity.taskDefinitionRevisions
	case entity.metrics != nil:
		data = entity.metrics
	case entity.autoScaling != nil:
		data = entity.autoScaling
	default:
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
		logger.Warnf("Failed to json marshal indent, error: %v", err)
		v.app.Notice.Warnf("Failed to json marshal indent, error: %v", err)
		return "json page marshal indent failed"
	}

	return colorizeJSON(jsonBytes)
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
			buff = append(buff, fmt.Sprintf(colorJSONFmt, res[1], res[2], res[3]))
		} else {
			buff = append(buff, l)
		}
	}

	return strings.Join(buff, "\n")
}
