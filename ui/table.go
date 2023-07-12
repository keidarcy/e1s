package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

// Build common table
func (v *View) buildTable(title string, headers []string, dataBuilder func() [][]string) {
	// init with first column width
	expansions := []int{2}
	alignment := []int{L}

	for i := 1; i < len(headers); i++ {
		expansions = append(expansions, 1)
		alignment = append(alignment, C)
	}

	v.table.
		SetFixed(5, 5).
		SetSelectable(true, false)

	v.table.
		SetBorder(true).
		SetTitle(title).
		SetBorderPadding(0, 0, 1, 1)

	data := [][]string{}
	data = append(data, headers)
	data = append(data, dataBuilder()...)

	for y, row := range data {
		for x, text := range row {
			cell := tview.NewTableCell(text).
				SetAlign(alignment[x]).
				SetExpansion(expansions[x]).
				SetMaxWidth(30)
			if y == 0 {
				cell.SetTextColor(tcell.ColorYellow)
				cell.SetSelectable(false)
			}
			v.table.SetCell(y, x, cell)
		}
	}

	v.handleTableEvents()

	pageName := v.kind.getTablePageName(v.getClusterArn())
	v.tablePages.AddPage(pageName, v.table, true, true)
}

// Handler common table events
func (v *View) handleTableEvents() {
	v.table.SetSelectionChangedFunc(v.handleSelectionChanged)

	v.table.SetSelectedFunc(v.handleSelected)

	v.table.SetInputCapture(v.handleInputCapture)

	v.table.SetDoneFunc(v.handleDone)
}

// Handle selected event for table when press up and down
// Detail page will switch
func (v *View) handleSelectionChanged(row, column int) {
	v.changeSelectedValues()
	selected := v.getCurrentSelection()
	v.infoPages.SwitchToPage(selected.entityName)
}

// Handle selected event for table when press Enter
func (v *View) handleSelected(row, column int) {
	err := v.handleAppPageSwitch(v.app.entityName, false)
	if err != nil {
		logger.Printf("page change failed, error: %v\n", err)
		v.back()
	}
}

// Handle keyboard input
func (v *View) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		key := event.Rune()
		if key == bKey || key == bKey-upperLowerDiff {
			v.openInBrowser()
		} else if key == dKey || key == dKey-upperLowerDiff {
			v.switchToResourceJson()
		} else if key == tKey || key == tKey-upperLowerDiff {
			v.switchToTaskDefinition()
		} else if key == rKey || key == rKey-upperLowerDiff {
			v.switchToTaskDefinitionRevisions()
		} else if key == wKey || key == wKey-upperLowerDiff {
			v.switchToServiceEvents()
		} else if key == mKey || key == mKey-upperLowerDiff {
			v.switchToMetrics()
		} else if key == aKey || key == aKey-upperLowerDiff {
			v.showAutoScaling()
			// } else if key == 'i' {
			// 	v.switchToAutoScaling()
		} else if key == eKey || key == eKey-upperLowerDiff {
			v.showUpdateServiceModal()
		} else if key == hKey || key == hKey-upperLowerDiff {
			v.handleDone(0)
		} else if key == lKey || key == lKey-upperLowerDiff {
			v.handleSelected(0, 0)
		}
	}
	return event
}

// Handle done event for table when press ESC
func (v *View) handleDone(key tcell.Key) {
	if v.kind == ClusterPage {
		return
	}
	v.back()
}

// Handle common values changing for selected event for table when pressed Enter
func (v *View) changeSelectedValues() {
	selected := v.getCurrentSelection()
	if v.kind == ClusterPage {
		v.app.cluster = selected.cluster
		v.app.entityName = *selected.cluster.ClusterArn
	} else if v.kind == ServicePage {
		v.app.service = selected.service
		v.app.entityName = *selected.service.ServiceArn
	} else if v.kind == TaskPage {
		v.app.task = selected.task
		v.app.entityName = *selected.task.TaskArn
	} else if v.kind == ContainerPage {
		v.app.container = selected.container
		v.app.entityName = *selected.container.ContainerArn
	} else {
		v.back()
	}
}

// Open selected resource in browser only support cluster and service
func (v *View) openInBrowser() {
	selected := v.getCurrentSelection()
	arn := ""
	taskService := ""
	switch v.kind {
	case ClusterPage:
		arn = *selected.cluster.ClusterArn
	case ServicePage:
		arn = *selected.service.ServiceArn
	case TaskPage:
		taskService = *v.app.service.ServiceName
		arn = *selected.task.TaskArn
	case ContainerPage:
		taskService = *v.app.service.ServiceName
		arn = *v.app.task.TaskArn
	}
	url := util.ArnToUrl(arn, taskService)
	if len(url) == 0 {
		return
	}
	logger.Printf("open url: %s\n", url)
	err := util.OpenURL(url)
	if err != nil {
		logger.Printf("failed open url %s\n", url)
	}
}
