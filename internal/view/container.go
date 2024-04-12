package view

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

type ContainerView struct {
	View
	containers []types.Container
}

func newContainerView(containers []types.Container, app *App) *ContainerView {
	keys := append(basicKeyInputs, []keyInput{
		{key: "shift-f", description: portForwarding},
		{key: "shift-t", description: terminatePortForwardingSession},
		{key: "enter", description: sshContainer},
		{key: "ctrl-d", description: exitContainer},
	}...)
	return &ContainerView{
		View: *newView(app, keys, secondaryPageKeyMap{
			DescriptionKind: describePageKeys,
		}),
		containers: containers,
	}
}

func (app *App) showContainersPage(reload bool) error {
	if switched := app.SwitchPage(reload); switched {
		return nil
	}

	// no containers exists do nothing
	if app.task == nil || len(app.task.Containers) == 0 {
		app.back()
		return fmt.Errorf("no valid container")
	}
	view := newContainerView(app.task.Containers, app)
	page := buildAppPage(view)
	app.addAppPage(page)
	view.table.Select(app.rowIndex, 0)
	return nil
}

// Build info pages for container page
func (v *ContainerView) infoBuilder() *tview.Pages {
	for _, c := range v.containers {
		title := utils.ArnToName(c.ContainerArn)
		entityName := *c.ContainerArn
		items := v.infoPagesParam(c)

		v.buildInfoPages(items, title, entityName)
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
	v.footer.container.SetText(fmt.Sprintf(footerSelectedItemFmt, v.app.kind))
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

	// v.table.SetInputCapture(v.handleInputCapture)
}

// Generate info pages params
func (v *ContainerView) infoPagesParam(c types.Container) (items []infoItem) {
	// Managed agents
	mas := []string{}
	for _, m := range c.ManagedAgents {
		mas = append(mas, string(m.Name))
	}
	masString := strings.Join(mas, ",")
	if len(masString) == 0 {
		masString = utils.EmptyText
	}

	items = []infoItem{
		{name: "Name", value: utils.ShowString(c.Name)},
		{name: "Task", value: utils.ShowString(c.TaskArn)},
		{name: "Image url", value: utils.ShowString(c.Image)},
		{name: "Image digest", value: utils.ShowString(c.ImageDigest)},
		{name: "Runtime ID", value: utils.ShowString(c.RuntimeId)},
		{name: "Last status", value: utils.ShowString(c.LastStatus)},
		{name: "CPU", value: utils.ShowString(c.Cpu)},
		{name: "Memory", value: utils.ShowString(c.Memory)},
		{name: "Memory reservation", value: utils.ShowString(c.MemoryReservation)},
		{name: "GPU IDs", value: utils.ShowArray(c.GpuIds)},
		{name: "Exit code", value: utils.ShowInt(c.ExitCode)},
		{name: "Reason", value: utils.ShowString(c.Reason)},
		{name: "Managed agents", value: utils.ShowString(&masString)},
	}
	return
}

// Generate table params
func (v *ContainerView) tableParam() (title string, headers []string, dataBuilder func() [][]string) {
	title = fmt.Sprintf(nsTitleFmt, v.app.kind, utils.ArnToName(v.app.task.TaskArn), len(v.containers))
	headers = []string{
		"Name",
		"Status",
		"Health status ▾",
		"PF",
		"Container runtime id",
		"Registry",
		"Image name",
	}

	dataBuilder = func() (data [][]string) {
		for _, c := range v.containers {
			containerId := fmt.Sprintf("%s.%s", *v.app.cluster.ClusterName, *c.Name)
			portText := utils.EmptyText
			ports := []string{}
			for _, session := range v.app.sessions {
				if session.containerId == containerId {
					ports = append(ports, session.port)
				}
			}
			if len(ports) != 0 {
				portText = strings.Join(ports, ",")
			}
			health := string(c.HealthStatus)

			registry, imageName := utils.ImageInfo(c.Image)

			row := []string{}
			row = append(row, utils.ShowString(c.Name))
			row = append(row, utils.ShowGreenGrey(c.LastStatus, "running"))
			row = append(row, utils.ShowGreenGrey(&health, "healthy"))
			row = append(row, portText)
			row = append(row, utils.ShowString(c.RuntimeId))
			row = append(row, registry)
			row = append(row, imageName)
			data = append(data, row)
		}
		return data
	}

	return
}

// SSH into selected container
func (v *View) ssh(containerName string) {
	if v.app.kind != ContainerKind {
		v.app.Notice.Warn("Invalid operation")
		return
	}
	if v.app.ReadOnly {
		v.app.Notice.Warn("No ecs exec permission in read only e1s mode")
		return
	}

	// catch ctrl+C & SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	bin, err := exec.LookPath(awsCli)
	if err != nil {
		logger.Warnf("Failed to find %s path, please check %s", awsCli, "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
		v.app.Notice.Warnf("Failed to find %s path, please check %s", awsCli, "https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html")
		v.app.back()
	}

	_, err = exec.LookPath(smpCi)
	if err != nil {
		logger.Warnf("Failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
		v.app.Notice.Warnf("Failed to find %s path, please check %s", smpCi, "https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html")
		v.app.back()
	}

	args := []string{
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
		v.app.Option.Shell,
	}

	logger.Infof("Exec: `%s %s`", awsCli, strings.Join(args, " "))

	v.app.Suspend(func() {
		cmd := exec.Command(bin, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		// ignore the stderr from ssh server
		_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(sshBannerFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, utils.ArnToName(v.app.task.TaskArn), containerName)))
		err = cmd.Run()
		// return signal
		signal.Stop(interrupt)
		close(interrupt)
	})
}
