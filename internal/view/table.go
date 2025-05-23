package view

import (
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

const (
	L = tview.AlignLeft
	C = tview.AlignCenter
	R = tview.AlignRight
)

// Build common table
func (v *view) buildTable(title string, headers []string, dataBuilder func() [][]string) {
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
				cell.SetTextColor(color.Color(theme.Yellow))
				cell.SetSelectable(false)
			}
			v.table.SetCell(y, x, cell)
		}
	}

	v.searchLast = new(string)

	v.handleTableEvents()

	pageName := v.app.kind.getTablePageName(v.app.getPageHandle())
	v.bodyPages.AddPage(pageName, v.table, true, true)

}

// Handler common table events
func (v *view) handleTableEvents() {
	v.table.SetSelectionChangedFunc(v.handleSelectionChanged)

	v.table.SetSelectedFunc(v.handleSelected)

	v.table.SetInputCapture(v.handleInputCapture)

	v.table.SetDoneFunc(v.handleDone)

	// prevent table row selection out of range
	if v.app.rowIndex >= v.table.GetRowCount() {
		v.app.rowIndex = 1
	}
}

// Handle selected event for table when press up and down
// Detail page will switch
func (v *view) handleSelectionChanged(row, column int) {
	v.changeSelectedValues()
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to handleSelectionChanged")
		return
	}
	v.app.rowIndex = row
	v.headerPages.SwitchToPage(selected.entityName)
}

// Handle selected event for table when press Enter
func (v *view) handleSelected(row, column int) {
	if v.app.kind == TaskDefinitionKind || v.app.kind == InstanceKind {
		return
	}
	v.app.rowIndex = 0
	v.app.showPrimaryKindPage(v.app.kind.nextKind(), false)
}

// Handle keyboard input
func (v *view) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// If it's single keystroke, event.Rune() is ascii code
	switch event.Rune() {
	case 'a':
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = AutoScalingKind
			v.showSecondaryKindPage(false)
			return event
		}
	case 'b':
		v.openInBrowser()
	case 'd':
		v.app.secondaryKind = DescriptionKind
		v.showSecondaryKindPage(false)
	case 'l':
		if v.app.kind == ServiceKind || v.app.kind == TaskKind || v.app.kind == ContainerKind {
			v.app.secondaryKind = LogKind
			v.showSecondaryKindPage(false)
			return event
		}
	case 'm':
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.serviceMetricsForm, 15)
			return event
		}
	case 't':
		if v.app.kind == ServiceKind || v.app.kind == TaskKind {
			v.showKindPage(TaskDefinitionKind, false)
			return event
		}
	case 'p':
		if v.app.kind == ServiceKind {
			v.showKindPage(ServiceDeploymentKind, false)
			return event
		}
	case 'v':
		if v.app.kind == ServiceDeploymentKind {
			v.app.secondaryKind = ServiceRevisionKind
			v.showSecondaryKindPage(false)
			return event
		}
	case 'R':
		if v.app.kind == ServiceDeploymentKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.rollbackServiceDeploymentForm, 6)
			return event
		}
	case 'x':
		if v.app.kind == TaskKind {
			if v.app.taskStatus == types.DesiredStatusRunning {
				v.app.taskStatus = types.DesiredStatusStopped
			} else {
				v.app.taskStatus = types.DesiredStatusRunning
			}
			v.showKindPage(TaskKind, false)
			return event
		}
	case 's':
		if v.app.kind == ContainerKind {
			v.execShell()
		}
		if v.app.kind == InstanceKind || v.app.kind == TaskKind {
			v.instanceStartSession()
		}
		return event
	case 'S':
		if v.app.kind == TaskKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.stopTaskForm, 6)
			return event
		}
	case 'N':
		if v.app.kind == ClusterKind {
			v.app.fromCluster = true
			v.showKindPage(TaskKind, false)
			return event
		}
	case 'n':
		if v.app.kind == ClusterKind {
			v.app.fromCluster = true
			v.showKindPage(InstanceKind, false)
			return event
		}
	case 'w':
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ServiceEventsKind
			v.showSecondaryKindPage(false)
			return event
		}
	case 'F':
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.portForwardingForm, 15)
			return event
		}
	case 'U':
		if v.app.kind == ServiceKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.serviceUpdateForm, 15)
			return event
		}
		if v.app.kind == TaskDefinitionKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.serviceUpdateWithSpecificTaskDefinitionForm, 6)
			return event
		}
	case 'T':
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.terminatePortForwardingForm, 6)
			return event
		}
	case 'P':
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.cpForm, 15)
			return event
		}
	case 'E':
		if v.app.kind == ContainerKind {
			v.app.secondaryKind = ModalKind
			v.showFormModal(v.execCommandForm, 7)
			return event
		}
	case 'D':
		v.app.secondaryKind = ModalKind
		v.showFormModal(v.catFile, 10)
		return event
	case '/':
		v.app.secondaryKind = ModalKind
		v.showSearchFormModal(v.searchForm, 5)
		return event
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
func (v *view) handleDone(key tcell.Key) {
	if v.app.kind == ClusterKind {
		return
	}
	v.app.back()
}

// Handle common values changing for selected event for table when pressed Enter
func (v *view) changeSelectedValues() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to changeSelectedValues")
		return
	}
	switch v.app.kind {
	case ClusterKind:
		cluster := selected.cluster
		if cluster != nil {
			v.app.cluster = cluster
			v.app.entityName = *cluster.ClusterArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case ServiceKind:
		service := selected.service
		if service != nil {
			v.app.service = service
			v.app.entityName = *service.ServiceArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case TaskKind:
		task := selected.task
		if task != nil {

			v.app.task = task
			v.app.entityName = *task.TaskArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case ContainerKind:
		container := selected.container
		if container != nil {
			v.app.container = selected.container
			v.app.entityName = *container.ContainerArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case TaskDefinitionKind:
		taskDefinition := selected.taskDefinition
		if taskDefinition != nil {
			v.app.taskDefinition = selected.taskDefinition
			v.app.entityName = *taskDefinition.TaskDefinitionArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case ServiceDeploymentKind:
		serviceDeployment := selected.serviceDeployment
		if serviceDeployment != nil {
			v.app.serviceDeployment = selected.serviceDeployment
			v.app.entityName = *serviceDeployment.ServiceDeploymentArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	case InstanceKind:
		instance := selected.instance
		if instance != nil {
			v.app.instance = selected.instance
			v.app.entityName = *instance.ContainerInstanceArn
		} else {
			slog.Warn("unexpected in changeSelectedValues", "kind", v.app.kind)
			return
		}
	default:
		v.app.back()
	}
}

// Open selected resource in browser only support cluster and service
func (v *view) openInBrowser() {
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warnf("failed to openInBrowser")
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
	case TaskDefinitionKind:
		arn = *v.app.taskDefinition.TaskDefinitionArn
	case ServiceDeploymentKind:
		arn = *v.app.serviceDeployment.ServiceDeploymentArn
	}
	url := utils.ArnToUrl(arn, taskService)
	if len(url) == 0 {
		slog.Warn("open failed", "url", url, "kind", v.app.kind, "arn", arn)
		return
	}
	slog.Info("open", "url", url)
	err = utils.OpenURL(url)
	if err != nil {
		v.app.Notice.Warnf("failed to open url %s\n", url)
	}
}
