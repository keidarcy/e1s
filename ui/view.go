package ui

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

const (
	titleFmt              = "[aqua::b]%s[aqua::-]([purple::b]%d[aqua::-]) "
	nsTitleFmt            = " [aqua::-]<[purple::b]%s[aqua::-]>" + titleFmt
	contentTitleFmt       = " [blue]%s([purple::b]%s[blue:-:-]) "
	infoTitleFmt          = " [blue]info([purple::b]%s[blue:-:-]) "
	footerSelectedItemFmt = " [black:aqua:b] <%s> [-:-:-] "
	footerItemFmt         = " [black:grey:] <%s> [-:-:-] "
	keyFmt                = " [purple::b]<%s> [green:-:-]%s "
	infoItemFmt           = " %s:[aqua::b] %s "
	clusterTasksFmt       = "[blue]%d pending[-] | [green]%d running"
	serviceTasksFmt       = "%d/%d tasks running"
	footerKeyFmt          = "[::b][↓,j/↑,k][::-] Down/Up [::b][Enter/Esc][::-] Enter/Back [::b][ctrl-c[][::-] Quit"
	colorJSONFmt          = `%s"[steelblue::b]%s[-:-:-]": %s`

	describe                        = "Describe"
	describeTaskDefinition          = "Describe task definition"
	describeTaskDefinitionRevisions = "Describe task definition revisions"
	describeServiceEvents           = "Describe service events"
	showAutoScaling                 = "Describe autoscaling targets, policies, actions, activities"
	showMetrics                     = "Show metrics"
	showLogs                        = "Show logs"

	editService        = "Edit Service"
	editTaskDefinition = "Edit Task Definition"
	reloadResource     = "Reload Resources"
	openInBrowser      = "Open in browser"
	sshContainer       = "SSH container"
	toggleFullScreen   = "Content Toggle full screen"

	shell        = "/bin/sh"
	awsCli       = "aws"
	sshBannerFmt = "\033[1;31m<<ECS-EXEC-SSH>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mTask\033[0m: \"%s\" \n\033[1;32mContainer\033[0m: \"%s\"\n#######################################\n"
)

const (
	aKey  = 'a'
	bKey  = 'b'
	dKey  = 'd'
	eKey  = 'e'
	fKey  = 'f'
	lKey  = 'l'
	mKey  = 'm'
	tKey  = 't'
	vKey  = 'v'
	wKey  = 'w'
	ctrlR = "ctrl-r"

	upperLowerDiff = rune(32)
)

var basicKeyInputs = []KeyInput{
	{key: string(bKey), description: openInBrowser},
	{key: string(dKey), description: describe},
	{key: ctrlR, description: reloadResource},
	{key: string(fKey), description: toggleFullScreen},
}

// Keyboard key and description
type KeyInput struct {
	key         string
	description string
}

// Info item name and value
type InfoItem struct {
	name  string
	value string
}

// Base struct of different views
type View struct {
	app           *App
	table         *tview.Table
	infoPages     *tview.Pages
	tablePages    *tview.Pages
	kind          Kind
	secondaryKind Kind
	keys          []KeyInput
	footer        *Footer
}

func newView(app *App, kind Kind, keys []KeyInput) *View {
	return &View{
		app:        app,
		infoPages:  tview.NewPages(),
		tablePages: tview.NewPages(),
		table:      tview.NewTable(),
		keys:       keys,
		kind:       kind,
		footer:     newFooter(),
	}
}

// Interface to show each view
type DataView interface {
	infoBuilder() *tview.Pages
	tableBuilder() *tview.Pages
	footerBuilder() *tview.Flex
}

// View footer struct
type Footer struct {
	footer    *tview.Flex
	cluster   *tview.TextView
	service   *tview.TextView
	task      *tview.TextView
	container *tview.TextView
}

func newFooter() *Footer {
	return &Footer{
		footer:    tview.NewFlex().SetDirection(tview.FlexColumn),
		cluster:   tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(footerItemFmt, ClusterPage)),
		service:   tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(footerItemFmt, ServicePage)),
		task:      tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(footerItemFmt, TaskPage)),
		container: tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf(footerItemFmt, ContainerPage)),
	}
}

const (
	// column height in info page
	oneColumnCount = 11
)

// Common function to build page for each view
func buildAppPage(v DataView) *tview.Flex {
	// build table reference first
	tablePages := v.tableBuilder()
	infoPages := v.infoBuilder()
	footer := v.footerBuilder()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(infoPages, oneColumnCount+2, 1, false).
		AddItem(tablePages, 0, 2, true).
		AddItem(footer, 1, 1, false)
	return flex
}

// Get current table selection and return as entity
func (v *View) getCurrentSelection() (Entity, error) {
	row, _ := v.table.GetSelection()
	if row == 0 {
		row++
	}
	cell := v.table.GetCell(row, 0)
	// entity := cell.GetReference().(Entity)
	switch entity := cell.GetReference().(type) {
	case Entity:
		return entity, nil
	default:
		logger.Printf("e1s - unexpected error: %v (%T)", entity, entity)
		return Entity{}, fmt.Errorf("unexpected error: %v (%T)", entity, entity)
	}
}

// Add new page to app.Pages
func (v *View) addAppPage(page *tview.Flex) {
	pageName := v.kind.getAppPageName(v.getClusterArn())
	v.app.Pages.AddPage(pageName, page, true, true)
}

func (v *View) getClusterArn() string {
	name := ""
	if v.kind != ClusterPage {
		name = *v.app.cluster.ClusterArn
	}
	return name
}

// Handle app.Pages switch
func (v *View) handleAppPageSwitch(resourceName string, isJson bool) error {
	kind := v.kind.nextKind()
	if isJson {
		kind = v.kind
	}
	v.showKindPage(kind)
	return nil
}

// Reload current resource
func (v *View) reloadResource() error {
	v.successModal("Reloaded ✅", 1, 20, 5)
	go v.showKindPage(v.kind)
	return nil
}

func (v *View) showKindPage(k Kind) error {
	switch v.secondaryKind {
	case LogsPage:
		v.switchToLogsList()
		return nil
	}
	switch k {
	case ClusterPage:
		return v.app.showClustersPage()
	case ServicePage:
		return v.app.showServicesPage()
	case TaskPage:
		return v.app.showTasksPages()
	case ContainerPage:
		return v.app.showContainersPage()
	default:
		v.app.showClustersPage()
	}
	return nil
}

// Go back page based on current kind
func (v *View) back() {
	toPage := v.kind.prevKind().getAppPageName(v.getClusterArn())
	v.kind = v.kind.prevKind()
	v.app.Pages.SwitchToPage(toPage)
}

// Go current page based on current kind
func (v *View) closeModal() {
	if v.app.cluster == nil {
		v.app.Stop()
		return
	}
	toPage := v.kind.getAppPageName(v.getClusterArn())
	v.app.Pages.SwitchToPage(toPage)
}

func (v *View) addFooterItems() {
	// left resources
	v.footer.footer.AddItem(v.footer.cluster, 13, 0, false).
		AddItem(v.footer.service, 13, 0, false).
		AddItem(v.footer.task, 10, 0, false).
		AddItem(v.footer.container, 15, 0, false)

	// keep middle space
	keysLabel := tview.NewTextView().
		// SetText(footerKeyFmt)
		SetText("")
	keysLabel.SetDynamicColors(true).SetTextAlign(tview.AlignCenter)
	v.footer.footer.
		AddItem(keysLabel, 0, 1, false)

	// right labels
	// name version label
	regionLabel := tview.NewTextView().SetDynamicColors(true).SetText(fmt.Sprintf("[black:yellow:bi] %s ", v.app.Region))
	v.footer.footer.AddItem(regionLabel, len(v.app.Region)+3, 0, false)

	appLabel := fmt.Sprintf("[black:blue:bi] %s:%s ", util.AppName, util.AppVersion)
	t := tview.NewTextView().SetTextAlign(L).SetDynamicColors(true).SetText(appLabel)
	v.footer.footer.AddItem(t, 15, 1, false)
}

func (v *View) handleContentPageSwitch(entity Entity, which string, contentString string) {
	contentTitle := fmt.Sprintf(contentTitleFmt, which, entity.entityName)
	contentPageName := v.kind.getContentPageName(entity.entityName + "." + which)

	contentTextItem := getContentTextItem(contentString, contentTitle)

	// press f toggle json
	contentTextItem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fullScreenContent := getContentTextItem(contentString, contentTitle)

		// full screen json press ESC close full screen json and back to table
		fullScreenContent.SetDoneFunc(func(key tcell.Key) {
			v.handleFullScreenContentDone(key)
			v.handleTableContentDone(key)
		})

		// full screen json press f close full screen json
		fullScreenContent.SetInputCapture(v.handleFullScreenContentInput)

		// contentTextComponent press f open in full screen
		switch event.Rune() {
		case fKey, fKey - upperLowerDiff:
			v.app.Pages.AddPage(contentPageName, fullScreenContent, true, true)
		case bKey, bKey - upperLowerDiff:
			v.openInBrowser()
		}

		switch event.Key() {
		case tcell.KeyCtrlR:
			v.reloadResource()
		}
		return event
	})

	contentTextItem.SetDoneFunc(v.handleTableContentDone)

	v.tablePages.AddPage(contentPageName, contentTextItem, true, true)
}

func (v *View) handleInfoPageSwitch(entity Entity, which string) {
	v.infoPages.SwitchToPage(*entity.cluster.ClusterArn + "json")
}

func getContentTextItem(contentStr string, title string) *tview.TextView {
	contentText := tview.NewTextView().SetDynamicColors(true).SetText(contentStr)
	contentText.SetBorder(true).SetTitle(title).SetBorderPadding(0, 0, 1, 1)
	return contentText
}

// SSH into selected container
func (v *View) ssh(containerName string) {
	if v.kind != ContainerPage {
		return
	}
	if v.app.readonly {
		return
	}

	// catch ctrl+C & SIGTERM
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

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
