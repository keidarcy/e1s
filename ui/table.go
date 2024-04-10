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

	pageName := v.app.kind.getTablePageName(v.app.getPageHandle())
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
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warn("Failed to handleSelectionChanged")
		logger.Warnf("Failed to handleSelectionChanged, err: %v", err)
		return
	}
	v.app.rowIndex = row
	v.infoPages.SwitchToPage(selected.entityName)
}

// Handle selected event for table when press Enter
func (v *View) handleSelected(row, column int) {
	if v.app.kind == TaskDefinitionKind {
		return
	}
	v.app.rowIndex = 0
	if v.app.kind == ContainerKind {
		selected, err := v.getCurrentSelection()
		if err != nil {
			v.app.Notice.Warn("Failed to handleSelected")
			logger.Warnf("Failed to handleSelected, err: %v", err)
			return
		}
		containerName := *selected.container.Name
		v.ssh(containerName)
	}
	v.app.showPrimaryKindPage(v.app.kind.nextKind(), false)
}

// Handle keyboard input
func (v *View) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// If it's single keystroke, event.Rune() is ascii code
	switch event.Rune() {
	case aKey:
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = AutoScalingKind
			v.showSecondaryKindPage(false)
			return event
		}
	case bKey:
		v.openInBrowser()
	case dKey:
		v.app.secondaryKind = DescriptionKind
		v.showSecondaryKindPage(false)
	case lKey:
		if v.app.kind == ServiceKind || v.app.kind == TaskKind {
			v.app.secondaryKind = LogKind
			v.showSecondaryKindPage(false)
			return event
		}
	case mKey:
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.serviceMetricsForm, 15)
			return event
		}
	case tKey:
		if v.app.kind == ServiceKind || v.app.kind == TaskKind {
			v.showKindPage(TaskDefinitionKind, false)
			return event
		}
	case wKey:
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ServiceEventsKind
			v.showSecondaryKindPage(false)
			return event
		}
	case FKey:
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.portForwardingForm, 15)
			return event
		}
	case UKey:
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.serviceUpdateForm, 15)
			return event
		}
	case TKey:
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showTerminatePortForwardingModal()
			return event
		}
	}

	// If it's composite keystroke, event.Key() is ctrl-char ascii code
	switch event.Key() {
	// Handle left arrow key
	case tcell.KeyLeft:
		v.handleDone(0)
	// Handle right arrow key
	case tcell.KeyRight:
		v.handleSelected(0, 0)
	// Handle <ctrl> + r
	case tcell.KeyCtrlR:
		v.reloadResource(true)
	case tcell.KeyCtrlZ:
		v.handleDone(0)
	}
	return event
}

// Handle done event for table when press ESC
func (v *View) handleDone(key tcell.Key) {
	if v.app.kind == ClusterKind {
		return
	}
	v.app.back()
}

// Handle common values changing for selected event for table when pressed Enter
func (v *View) changeSelectedValues() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warn("Failed to changeSelectedValues")
		logger.Warnf("Failed to changeSelectedValues, err: %v", err)
		return
	}
	switch v.app.kind {
	case ClusterKind:
		cluster := selected.cluster
		if cluster != nil {
			v.app.cluster = cluster
			v.app.entityName = *selected.cluster.ClusterArn
		} else {
			logger.Warnf("unexpected in changeSelectedValues kind: %s", v.app.kind)
			return
		}
	case ServiceKind:
		service := selected.service
		if service != nil {
			v.app.service = service
			v.app.entityName = *selected.service.ServiceArn
		} else {
			logger.Warnf("unexpected in changeSelectedValues kind: %s", v.app.kind)
			return
		}
	case TaskKind:
		task := selected.task
		if task != nil {

			v.app.task = task
			v.app.entityName = *selected.task.TaskArn
		} else {
			logger.Warnf("unexpected in changeSelectedValues kind: %s", v.app.kind)
			return
		}
	case ContainerKind:
		container := selected.container
		if container != nil {
			v.app.container = selected.container
			v.app.entityName = *selected.container.ContainerArn
		} else {
			logger.Warnf("unexpected in changeSelectedValues kind: %s", v.app.kind)
			return
		}
	case TaskDefinitionKind:
		taskDefinition := selected.taskDefinition
		if taskDefinition != nil {
			v.app.taskDefinition = selected.taskDefinition
			v.app.entityName = *selected.taskDefinition.TaskDefinitionArn
		} else {
			logger.Warnf("unexpected in changeSelectedValues kind: %s", v.app.kind)
			return
		}
	default:
		v.app.back()
	}
}

// Open selected resource in browser only support cluster and service
func (v *View) openInBrowser() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warn("Failed to openInBrowser")
		logger.Warnf("Failed to openInBrowser, err: %v", err)
		return
	}
	arn := ""
	taskService := ""
	switch v.app.kind {
	case ClusterKind:
		arn = *selected.cluster.ClusterArn
	case ServiceKind:
		arn = *selected.service.ServiceArn
	case TaskKind:
		taskService = *v.app.service.ServiceName
		arn = *selected.task.TaskArn
	case ContainerKind:
		taskService = *v.app.service.ServiceName
		arn = *v.app.task.TaskArn
	}
	url := util.ArnToUrl(arn, taskService)
	if len(url) == 0 {
		return
	}
	logger.Infof("Open url: %s\n", url)
	err = util.OpenURL(url)
	if err != nil {
		logger.Warnf("Failed to open url %s\n", url)
		v.app.Notice.Warnf("Failed to open url %s\n", url)
	}
}
