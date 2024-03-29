package ui

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
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
	v.infoPages.SwitchToPage(selected.entityName)
}

// Handle selected event for table when press Enter
func (v *View) handleSelected(row, column int) {
	if v.app.kind == ContainerPage {
		selected, err := v.getCurrentSelection()
		if err != nil {
			v.app.Notice.Warn("Failed to handleSelected")
			logger.Warnf("Failed to handleSelected, err: %v", err)
			return
		}
		containerName := *selected.container.Name
		v.ssh(containerName)
	}
	v.app.showPrimaryKindPage(v.app.kind.nextKind(), false, 0)
}

// Handle keyboard input
func (v *View) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// If it's single keystroke, event.Rune() is ascii code
	switch event.Rune() {
	case aKey:
		if v.app.kind == ServicePage {
			v.app.secondaryKind = AutoScalingPage
			v.showSecondaryKindPage(false)
			return event
		}
	case bKey:
		v.openInBrowser()
	case dKey:
		v.app.secondaryKind = DescriptionPage
		v.showSecondaryKindPage(false)
	case eKey:
		if v.app.kind == ServicePage {
			v.showEditServiceModal()
			return event
		}
		if v.app.kind == TaskPage {
			v.editTaskDefinition()
			return event
		}
	case lKey:
		if v.app.kind == ServicePage || v.app.kind == TaskPage {
			v.app.secondaryKind = LogPage
			v.showSecondaryKindPage(false)
			return event
		}
	case mKey:
		if v.app.kind == ServicePage {
			v.showMetricsModal()
			return event
		}
	case tKey:
		if v.app.kind == ServicePage || v.app.kind == TaskPage {
			v.app.secondaryKind = TaskDefinitionPage
			v.showSecondaryKindPage(false)
			return event
		}
	case vKey:
		if v.app.kind == ServicePage || v.app.kind == TaskPage {
			v.app.secondaryKind = TaskDefinitionRevisionsPage
			v.showSecondaryKindPage(false)
			return event
		}
	case wKey:
		if v.app.kind == ServicePage {
			v.app.secondaryKind = ServiceEventsPage
			v.showSecondaryKindPage(false)
			return event
		}
	case FKey:
		if v.app.kind == ContainerPage {
			v.showPortForwardingModal()
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
	if v.app.kind == ClusterPage {
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
	if v.app.kind == ClusterPage {
		v.app.cluster = selected.cluster
		v.app.entityName = *selected.cluster.ClusterArn
	} else if v.app.kind == ServicePage {
		v.app.service = selected.service
		v.app.entityName = *selected.service.ServiceArn
	} else if v.app.kind == TaskPage {
		v.app.task = selected.task
		v.app.entityName = *selected.task.TaskArn
	} else if v.app.kind == ContainerPage {
		v.app.container = selected.container
		v.app.entityName = *selected.container.ContainerArn
	} else {
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
	logger.Infof("Open url: %s\n", url)
	err = util.OpenURL(url)
	if err != nil {
		logger.Warnf("Failed to open url %s\n", url)
		v.app.Notice.Warnf("Failed to open url %s\n", url)
	}
}

func (v *View) editTaskDefinition() {
	// get td detail
	selected, err := v.getCurrentSelection()
	if err != nil {
		v.app.Notice.Warn("Failed to editTaskDefinition")
		logger.Warnf("Failed to editTaskDefinition, err: %v", err)
		return
	}
	taskDefinition := *selected.task.TaskDefinitionArn
	td, err := v.app.Store.DescribeTaskDefinition(&taskDefinition)
	if err != nil {
		logger.Warnf("Failed to describe task definition, err: %v", err)
		v.app.Notice.Warnf("Failed to describe task definition, err: %v", err)
		return
	}
	names := strings.Split(selected.entityName, "/")

	// create tmp file open and defer close it
	tmpfile, err := os.CreateTemp("", names[len(names)-1])
	if err != nil {
		logger.Warnf("Failed to create temporary file, err: %v", err)
		v.app.Notice.Warnf("Failed to create temporary file, err: %v", err)
		return
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	originalTD, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		logger.Warnf("Failed to read temporary file, err: %v", err)
		v.app.Notice.Warnf("Failed to read temporary file, err: %v", err)
		return
	}

	if _, err := tmpfile.Write(originalTD); err != nil {
		logger.Warnf("Failed to write to temporary file, err: %v", err)
		v.app.Notice.Warnf("Failed to write to temporary file, err: %v", err)
		return
	}

	// Open the vi editor to allow the user to modify the JSON data.
	bin := os.Getenv("EDITOR")
	if bin == "" {
		// if $EDITOR is empty use vi as default
		bin = "vi"
	}

	v.app.Suspend(func() {
		cmd := exec.Command(bin, tmpfile.Name())
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

		if err := cmd.Run(); err != nil {
			logger.Warnf("Failed to open editor, err: %v", err)
			v.app.Notice.Warnf("Failed to open editor, err: %v", err)
			return
		}

		editedTD, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			logger.Warnf("Failed to read temporary file, err: %v", err)
			v.app.Notice.Warnf("Failed to read temporary file, err: %v", err)
			return
		}

		// remove edited empty line
		if editedTD[len(editedTD)-1] == '\n' {
			originalTD = append(originalTD, '\n')
		}

		// if no change do nothing
		if bytes.Equal(originalTD, editedTD) {
			v.app.Notice.Info("Task definition has no change")
			return
		}

		var updatedTd ecs.RegisterTaskDefinitionInput
		if err := json.Unmarshal(editedTD, &updatedTd); err != nil {
			logger.Warnf("Failed to unmarshal JSON, err: %v", err)
			v.app.Notice.Warnf("Failed to unmarshal JSON, err: %v", err)
			return
		}

		register := func() {
			family, revision, err := v.app.Store.RegisterTaskDefinition(&updatedTd)

			if err != nil {
				logger.Warnf("Failed to open editor, err: %v", err)
				v.app.Notice.Warnf("Failed to open editor, err: %v", err)
				return
			}
			v.app.Notice.Infof("SUCCESS TaskDefinition Family: %s, Revision: %d", family, revision)
		}

		v.showTaskDefinitionConfirm(register)
	})
}
