package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

// Switch to selected resource JSON page
func (v *View) switchToResourceJson() {
	selected := v.getCurrentSelection()
	v.showJsonPages(selected, "describe")
}

// Switch to selected task definition JSON page
func (v *View) switchToTaskDefinitionJson() {
	selected := v.getCurrentSelection()
	taskDefinition := ""
	if v.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
	} else if v.kind == TaskPage {
		taskDefinition = *selected.task.TaskDefinitionArn
	} else {
		return
	}

	td, err := v.app.Store.DescribeTaskDefinition(&taskDefinition)
	if err != nil {
		return
	}
	entity := Entity{taskDefinition: &td, entityName: *td.TaskDefinitionArn}
	v.showJsonPages(entity, "task definition")
}

// Switch to selected task definition revision list JSON page
func (v *View) switchToTaskDefinitionRevisionsJson() {
	family, _ := v.getTaskDefinitionDetail()

	revisions, err := v.app.Store.ListTaskDefinition(&family)
	if err != nil {
		return
	}
	entity := Entity{taskDefinitionRevisions: revisions, entityName: family}
	v.showJsonPages(entity, "revisions")
}

// Get td family
func (v *View) getTaskDefinitionDetail() (string, string) {
	selected := v.getCurrentSelection()
	taskDefinition := ""
	if v.kind == ServicePage {
		taskDefinition = *selected.service.TaskDefinition
	} else if v.kind == TaskPage {
		taskDefinition = *selected.task.TaskDefinitionArn
	} else {
		return "", ""
	}
	familyRevision := strings.Split(util.ArnToName(&taskDefinition), ":")
	return familyRevision[0], familyRevision[1]
}

// Switch to selected service events JSON page
func (v *View) switchToServiceEventsJson() {
	selected := v.getCurrentSelection()
	if v.kind != ServicePage {
		return
	}
	v.showJsonPages(selected, "events")
}

// Deprecated
// Switch to Metrics get by cloudwatch
func (v *View) switchToMetrics() {
	selected := v.getCurrentSelection()
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
	selected := v.getCurrentSelection()
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
	title := fmt.Sprintf(jsonTitleFmt, which, entity.entityName)
	jsonString := v.getJsonString(entity, which)
	jsonPageName := v.kind.getJsonPageName(entity.entityName + "." + which)

	tableJson := jsonTextView(jsonString, title)

	// press f toggle json
	tableJson.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fullScreenJson := jsonTextView(jsonString, title)

		// full screen json press ESC close full screen json and back to table
		fullScreenJson.SetDoneFunc(func(key tcell.Key) {
			v.handleFullScreenJsonDone(key)
			v.handleTableJsonDone(key)
		})

		// full screen json press f close full screen json
		fullScreenJson.SetInputCapture(v.handleFullScreenJsonInput)

		// tableJson press f open in full screen
		if event.Key() == tcell.KeyRune {
			key := event.Rune()
			if key == fKey || key == fKey-upperLowerDiff {
				v.app.Pages.AddPage(jsonPageName, fullScreenJson, true, true)
			}
		}
		return event
	})

	tableJson.SetDoneFunc(v.handleTableJsonDone)

	v.tablePages.AddPage(jsonPageName, tableJson, true, true)
}

func (v *View) handleFullScreenJsonInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		key := event.Rune()
		if key == fKey || key == fKey-upperLowerDiff {
			pageName := v.kind.getAppPageName(v.getClusterArn())
			v.app.Pages.SwitchToPage(pageName)
		}
	}
	return event
}

func (v *View) handleTableJsonDone(key tcell.Key) {
	pageName := v.kind.getTablePageName(v.getClusterArn())
	v.tablePages.SwitchToPage(pageName)
}

func (v *View) handleFullScreenJsonDone(key tcell.Key) {
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
		logger.Printf("json page marshal indent failed, error: %v\n", err)
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

func jsonTextView(jsonData string, title string) *tview.TextView {
	jsonText := tview.NewTextView().SetDynamicColors(true).SetText(jsonData)
	jsonText.SetBorder(true).SetTitle(title).SetBorderPadding(0, 0, 1, 1)
	return jsonText
}
