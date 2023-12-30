package ui

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

const (
	shell        = "/bin/sh"
	awsCli       = "aws"
	sshBannerFmt = "\033[1;31m<<ECS-EXEC-SSH>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mTask\033[0m: \"%s\" \n\033[1;32mContainer\033[0m: \"%s\"\n#######################################\n"
)

type ContainerView struct {
	View
	containers []types.Container
}

func newContainerView(containers []types.Container, app *App) *ContainerView {
	keys := append(basicKeyInputs, []KeyInput{
		{key: "Enter", description: sshContainer},
	}...)
	return &ContainerView{
		View:       *newView(app, ContainerPage, keys),
		containers: containers,
	}
}

func (app *App) showContainersPage() error {
	// no containers exists do nothing
	if len(app.task.Containers) == 0 {
		return nil
	}
	view := newContainerView(app.task.Containers, app)
	page := buildAppPage(view)
	view.addAppPage(page)
	return nil
}

// Build info pages for container page
func (v *ContainerView) infoBuilder() *tview.Pages {
	for _, c := range v.containers {
		items := v.infoPagesParam(c)
		infoFlex := v.buildInfoFlex(util.ArnToName(c.ContainerArn), items, v.keys)
		v.infoPages.AddPage(*c.ContainerArn, infoFlex, true, true)
	}
	// prevent empty containers
	if len(v.containers) > 0 && v.containers[0].ContainerArn != nil {
		// show first when enter
		v.infoPages.SwitchToPage(*v.containers[0].ContainerArn)
		v.changeSelectedValues()
	}
	return v.infoPages
}

// Build table for container page
func (v *ContainerView) tableBuilder() *tview.Pages {
	title, headers, dataBuilder := v.tableParam()
	v.buildTable(title, headers, dataBuilder)
	v.tableHandler()
	return v.tablePages
}

// Build footer for container page
func (v *ContainerView) footerBuilder() *tview.Flex {
	v.footer.container.SetText(fmt.Sprintf(footerSelectedItemFmt, v.kind))
	v.addFooterItems()
	return v.footer.footer
}

// Handlers for container table
func (v *ContainerView) tableHandler() {
	for row, container := range v.containers {
		c := container
		v.table.GetCell(row+1, 0).SetReference(Entity{container: &c, entityName: *c.ContainerArn})
	}

	v.table.SetSelectedFunc(func(row int, column int) {
		containerName := v.table.GetCell(row, column).Text
		v.ssh(containerName)
	})

	v.table.SetInputCapture(v.handleInputCapture)
}

// Container page specific input handler
func (v *ContainerView) handleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	// simulate selected action(ssh)
	sshHandler := func() {
		selected, err := v.getCurrentSelection()
		if err != nil {
			return
		}
		containerName := *selected.container.Name
		v.ssh(containerName)
	}

	// handle right arrow key
	if event.Key() == tcell.KeyRight {
		sshHandler()
		return event
	}

	// handle l key
	key := event.Rune()
	switch key {
	case rKey, rKey - upperLowerDiff:
		v.reloadResource()
	case lKey, lKey - upperLowerDiff:
		sshHandler()
	case hKey, hKey - upperLowerDiff:
		v.handleDone(0)
	case bKey, bKey - upperLowerDiff:
		v.openInBrowser()
	case dKey, dKey - upperLowerDiff:
		v.switchToResourceJson()
	}
	return event
}

// Generate info pages params
func (v *ContainerView) infoPagesParam(c types.Container) (items []InfoItem) {
	// Managed agents
	mas := []string{}
	for _, m := range c.ManagedAgents {
		mas = append(mas, string(m.Name))
	}
	masString := strings.Join(mas, ",")
	if len(masString) == 0 {
		masString = util.EmptyText
	}

	items = []InfoItem{
		{name: "Name", value: util.ShowString(c.Name)},
		{name: "Task", value: util.ShowString(c.TaskArn)},
		{name: "Image url", value: util.ShowString(c.Image)},
		{name: "Image digest", value: util.ShowString(c.ImageDigest)},
		{name: "Runtime ID", value: util.ShowString(c.RuntimeId)},
		{name: "Last status", value: util.ShowString(c.LastStatus)},
		{name: "CPU", value: util.ShowString(c.Cpu)},
		{name: "Memory", value: util.ShowString(c.Memory)},
		{name: "Memory reservation", value: util.ShowString(c.MemoryReservation)},
		{name: "GPU IDs", value: util.ShowArray(c.GpuIds)},
		{name: "Exit code", value: util.ShowInt(c.ExitCode)},
		{name: "Reason", value: util.ShowString(c.Reason)},
		{name: "Managed agents", value: util.ShowString(&masString)},
	}
	return
}

// Generate table params
func (v *ContainerView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.kind, util.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Health status ▾",
		"Status",
		"Container runtime id",
		"Image URI",
	}
	dataBuilder = func() (data [][]string) {
		for _, c := range v.containers {
			health := string(c.HealthStatus)

			row := []string{}
			row = append(row, util.ShowString(c.Name))
			row = append(row, util.ShowGreenGrey(&health, "healthy"))
			row = append(row, util.ShowGreenGrey(c.LastStatus, "running"))
			row = append(row, util.ShowString(c.RuntimeId))
			row = append(row, util.ShowString(c.Image))
			data = append(data, row)
		}
		return data
	}

	return
}

// SSH into selected container
func (v *ContainerView) ssh(containerName string) {
	// catch ctrl+C & SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	if v.app.readonly {
		return
	}
	bin, err := exec.LookPath(awsCli)
	if err != nil {
		logger.Printf("e1s - aws cli binary not found, error: %v\n", err)
		v.back()
	}
	arg := []string{
		"ecs",
		"execute-command",
		"--cluster",
		*v.app.cluster.ClusterName,
		"--task",
		*v.app.task.TaskArn,
		"--container",
		containerName,
		"--interactive",
		"--command",
		shell,
	}

	logger.Printf("%s %s\n", awsCli, strings.Join(arg, " "))

	v.app.Suspend(func() {
		cmd := exec.Command(bin, arg...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		// ignore the stderr from ssh server
		_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(sshBannerFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, util.ArnToName(v.app.task.TaskArn), containerName)))
		err = cmd.Run()
		// return signal
		signal.Stop(interrupt)
		close(interrupt)
	})
}
