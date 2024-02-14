package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
)

// Switch to selected resource JSON page
func (v *View) switchToResourceJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	v.showJsonPages(selected, "describe")
}

// Switch to selected task definition JSON page
func (v *View) switchToTaskDefinitionJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	taskDefinition := ""
	entityName := ""
	if v.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
		entityName = *selected.service.ServiceArn
	} else if v.kind == TaskPage {
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
	v.showJsonPages(entity, "task definition")
}

// Switch to selected task definition revision list JSON page
func (v *View) switchToTaskDefinitionRevisionsJson() {
	if v.kind == ClusterPage {
		return
	}
	family, _, entityName := v.getTaskDefinitionDetail()

	revisions, err := v.app.Store.ListTaskDefinition(&family)
	if err != nil {
		return
	}
	entity := Entity{taskDefinitionRevisions: revisions, entityName: entityName}
	v.showJsonPages(entity, "revisions")
}

// Get td family
func (v *View) getTaskDefinitionDetail() (string, string, string) {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return "", "", ""
	}
	taskDefinition := ""
	entityName := ""
	if v.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
		entityName = *selected.service.ServiceArn
	} else if v.kind == TaskPage {
		taskDefinition = *selected.task.TaskDefinitionArn
		entityName = *selected.task.TaskArn
	} else {
		return "", "", ""
	}
	familyRevision := strings.Split(util.ArnToName(&taskDefinition), ":")
	return familyRevision[0], familyRevision[1], entityName
}

// Deprecated
// Switch to Metrics get by cloudwatch
func (v *View) switchToMetrics() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	if v.kind != ServicePage {
		return
	}
	cluster := v.app.cluster.ClusterName
	service := selected.service.ServiceName

	metrics, err := v.app.Store.GetMetrics(cluster, service)

	if err != nil {
		return
	}
	entity := Entity{metrics: metrics, entityName: fmt.Sprintf("%s/%s", *cluster, *service)}
	v.showJsonPages(entity, "metrics")
}

// Deprecated
// not called anywhere
// Switch to auto scaling get by applicationautoscaling
func (v *View) switchToAutoScalingJson() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	if v.kind != ServicePage {
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
	v.showJsonPages(entity, "auto_scaling")
}

// Show new page from JSON content in table area and handle done event to go back
func (v *View) showJsonPages(entity Entity, which string) {
	contentString := v.getJsonString(entity, which)
	v.handleContentPageSwitch(entity, which, contentString)
	v.handleInfoPageSwitch(entity, JsonPage)
}

func (v *View) handleFullScreenContentInput(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case fKey, fKey - upperLowerDiff:
		pageName := v.kind.getAppPageName(v.getClusterArn())
		v.app.Pages.SwitchToPage(pageName)
	}

	switch event.Key() {
	case tcell.KeyCtrlR:
		v.reloadResource()
	}
	return event
}

func (v *View) handleTableContentDone(key tcell.Key) {
	v.secondaryKind = v.kind
	pageName := v.kind.getTablePageName(v.getClusterArn())
	v.tablePages.SwitchToPage(pageName)

	selected, err := v.getCurrentSelection()
	if err != nil {
		v.back()
	}
	v.infoPages.SwitchToPage(selected.entityName)
}

func (v *View) handleFullScreenContentDone(key tcell.Key) {
	pageName := v.kind.getAppPageName(v.getClusterArn())
	v.app.Pages.SwitchToPage(pageName)
}

func (v *View) getJsonString(entity Entity, which string) string {
	var data any

	switch {
	case entity.cluster != nil:
		data = entity.cluster
	// events need be upper then service
	case entity.events != nil && which == "events":
		data = entity.events
	case entity.service != nil:
		data = entity.service
	case entity.task != nil:
		data = entity.task
	case entity.container != nil:
		data = entity.container
	case entity.taskDefinition != nil:
		data = entity.taskDefinition
	case entity.taskDefinitionRevisions != nil:
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
