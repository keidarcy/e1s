package ui

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	v.infoPages.SwitchToPage(selected.entityName)
}

// Handle selected event for table when press Enter
func (v *View) handleSelected(row, column int) {
	err := v.handleAppPageSwitch(v.app.entityName, false)
	if err != nil {
		logger.Printf("e1s - page change failed, error: %v\n", err)
		v.back()
	}
}

// Handle keyboard input
func (v *View) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() != tcell.KeyRune {
		switch event.Key() {
		case tcell.KeyLeft:
			// Handle left arrow key
			v.handleDone(0)
		case tcell.KeyRight:
			// Handle right arrow key
			v.handleSelected(0, 0)
		}
		return event
	}

	key := event.Rune()
	switch key {
	case bKey, bKey - upperLowerDiff:
		v.openInBrowser()
	case dKey, dKey - upperLowerDiff:
		v.switchToResourceJson()
	case tKey, tKey - upperLowerDiff:
		v.switchToTaskDefinitionJson()
	case rKey, rKey - upperLowerDiff:
		v.reloadResource()
	case vKey, vKey - upperLowerDiff:
		v.switchToTaskDefinitionRevisionsJson()
	case wKey, wKey - upperLowerDiff:
		v.switchToServiceEventsList()
	case mKey, mKey - upperLowerDiff:
		v.showMetricsModal()
	case aKey, aKey - upperLowerDiff:
		v.switchToAutoScalingJson()
		// v.showAutoScalingModal()
	case eKey, eKey - upperLowerDiff:
		v.showEditServiceModal()
		v.editTaskDefinition()
	case hKey, hKey - upperLowerDiff:
		v.handleDone(0)
	case lKey, lKey - upperLowerDiff:
		v.handleSelected(0, 0)
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
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
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
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
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
	err = util.OpenURL(url)
	if err != nil {
		logger.Printf("e1s - failed open url %s\n", url)
	}
}

func (v *View) editTaskDefinition() {
	const errMsg = "Error when editing task definition"
	if v.kind != TaskPage {
		return
	}

	// get td detail
	selected, err := v.getCurrentSelection()
	if err != nil {
		return
	}
	taskDefinition := *selected.task.TaskDefinitionArn
	td, err := v.app.Store.DescribeTaskDefinition(&taskDefinition)
	if err != nil {
		v.errorModal(errMsg, 2, 110, 10)
		return
	}
	names := strings.Split(selected.entityName, "/")

	// create tmp file open and defer close it
	tmpfile, err := os.CreateTemp("", names[len(names)-1])
	if err != nil {
		logger.Println("Error creating temporary file:", err)
		v.errorModal(errMsg, 2, 110, 10)
		return
	}
	defer os.Remove(tmpfile.Name())
	defer tmpfile.Close()

	originalTD, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		logger.Println("Error reading temporary file:", err)
		v.errorModal(errMsg, 2, 110, 10)
		return
	}

	if _, err := tmpfile.Write(originalTD); err != nil {
		logger.Println("Error writing to temporary file:", err)
		v.errorModal(errMsg, 2, 110, 10)
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
			logger.Println("Error opening editor:", err)
			v.errorModal(errMsg, 2, 110, 10)
			return
		}

		editedTD, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			logger.Println("Error reading temporary file:", err)
			v.errorModal(errMsg, 2, 110, 10)
			return
		}

		// remove edited empty line
		if editedTD[len(editedTD)-1] == '\n' {
			originalTD = append(originalTD, '\n')
		}

		// if no change do nothing
		if bytes.Equal(originalTD, editedTD) {
			v.flashModal(" Task definition has no change.", 2, 50, 3)
			return
		}

		var updatedTd ecs.RegisterTaskDefinitionInput
		if err := json.Unmarshal(editedTD, &updatedTd); err != nil {
			logger.Println("Error unmarshaling JSON:", err)
			v.errorModal(errMsg, 2, 110, 10)
			return
		}

		register := func() {
			family, revision, err := v.app.Store.RegisterTaskDefinition(&updatedTd)

			if err != nil {
				logger.Println("Error opening editor:", err)
				v.errorModal(errMsg, 2, 110, 10)
				return
			}
			v.successModal(fmt.Sprintf("SUCCESS 🚀\nTaskDefinition Family: %s\nRevision: %d\n", family, revision), 5, 110, 5)
		}

		v.showTaskDefinitionConfirm(register)

	})
}
