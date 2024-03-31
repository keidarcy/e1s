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
	"github.com/sirupsen/logrus"
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

	editService                    = "Edit Service"
	editTaskDefinition             = "Edit Task Definition"
	reloadResource                 = "Reload Resources"
	openInBrowser                  = "Open in browser"
	sshContainer                   = "SSH container"
	portForwarding                 = "Port forwarding session"
	terminatePortForwardingSession = "Terminate port forwarding session"
	toggleFullScreen               = "Content Toggle full screen"
	realtimeLog                    = "Cloudwatch realtime logs(only support one log group)"
	backToPrevious                 = "Back"

	// shell        = "/bin/sh -c \"if [ -x /bin/bash ]; then exec /bin/bash; else exec /bin/sh; fi\""
	shell          = "/bin/sh"
	awsCli         = "aws"
	smpCi          = "session-manager-plugin"
	sshBannerFmt   = "\033[1;31m<<E1S-ECS-EXEC>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mTask\033[0m: \"%s\" \n\033[1;32mContainer\033[0m: \"%s\"\n#######################################\n"
	realtimeLogFmt = "\033[1;31m<<E1S-LOGS-TAIL>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mLogGroup\033[0m: \"%s\"\n#######################################\n"

	reloadText = "Reloaded"
)

const (
	aKey  = 'a'
	bKey  = 'b'
	dKey  = 'd'
	eKey  = 'e'
	fKey  = 'f'
	lKey  = 'l'
	mKey  = 'm'
	rKey  = 'r'
	tKey  = 't'
	vKey  = 'v'
	wKey  = 'w'
	FKey  = 'F'
	TKey  = 'T'
	ctrlR = "ctrl-r"
	ctrlZ = "ctrl-z"
)

var basicKeyInputs = []KeyInput{
	{key: string(bKey), description: openInBrowser},
	{key: string(dKey), description: describe},
	{key: ctrlR, description: reloadResource},
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
	app        *App
	table      *tview.Table
	infoPages  *tview.Pages
	tablePages *tview.Pages
	keys       []KeyInput
	footer     *Footer
	pageKeyMap secondaryPageKeyMap
}

func newView(app *App, keys []KeyInput, pageKeys secondaryPageKeyMap) *View {
	return &View{
		app:        app,
		infoPages:  tview.NewPages(),
		tablePages: tview.NewPages(),
		table:      tview.NewTable(),
		keys:       keys,
		footer:     newFooter(),
		pageKeyMap: pageKeys,
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
		logger.Warnf("Unexpected error in getCurrentSelection: %v (%T)", entity, entity)
		v.app.Notice.Warnf("Unexpected error in getCurrentSelection: %v (%T)", entity, entity)
		return Entity{}, fmt.Errorf("unexpected error in getCurrentSelection: %v (%T)", entity, entity)
	}
}

// Reload current resource
func (v *View) reloadResource(reloadNotice bool) error {
	if reloadNotice {
		v.app.Notice.Info(reloadText)
	}
	v.showKindPage(v.app.kind, true)
	return nil
}

// Show kind page including primary kind, secondary kind
func (v *View) showKindPage(k Kind, reload bool) {
	if v.app.secondaryKind != EmptyKind {
		v.showSecondaryKindPage(reload)
		return
	}
	rowIndex, _ := v.table.GetSelection()
	v.app.showPrimaryKindPage(k, reload, rowIndex)
}

func (v *View) showSecondaryKindPage(reload bool) {
	switch v.app.secondaryKind {
	case AutoScalingPage:
		v.switchToAutoScalingJson()
	case DescriptionPage:
		v.switchToDescriptionJson()
	case LogPage:
		v.switchToLogsList()
	case TaskDefinitionPage:
		v.switchToTaskDefinitionJson()
	case TaskDefinitionRevisionsPage:
		v.switchToTaskDefinitionRevisionsJson()
	case ServiceEventsPage:
		v.switchToServiceEventsList()
	}
	if !reload {
		v.app.Notice.Infof("Viewing %s...", v.app.secondaryKind.String())
	} else {
		logger.Debugf("Reload in showSecondaryKindPage: %v", reload)
	}
}

// Go current page based on current kind
func (v *View) closeModal() {
	if v.app.cluster == nil {
		v.app.Stop()
		return
	}
	// v.app.secondaryKind = EmptyKind
	toPage := v.app.kind.getAppPageName(v.app.getPageHandle())
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

// Content page builder
func (v *View) handleContentPageSwitch(entity Entity, contentString string) {
	contentTitle := fmt.Sprintf(contentTitleFmt, v.app.kind, entity.entityName)
	contentPageName := v.app.kind.getContentPageName(entity.entityName + "." + v.app.secondaryKind.String())

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
		case fKey:
			v.app.Pages.AddPage(contentPageName, fullScreenContent, true, true)
		case bKey:
			v.openInBrowser()
		case rKey:
			if v.app.secondaryKind == LogPage {
				v.realtimeAwsLog(entity)
			}

		}

		switch event.Key() {
		case tcell.KeyCtrlR:
			v.reloadResource(true)
		case tcell.KeyCtrlZ:
			v.handleTableContentDone(0)
		}
		return event
	})

	contentTextItem.SetDoneFunc(v.handleTableContentDone)

	logger.WithFields(logrus.Fields{
		"Action":        "AddPage",
		"PageName":      contentPageName,
		"Kind":          v.app.kind.String(),
		"SecondaryKind": v.app.secondaryKind.String(),
		"Cluster":       *v.app.cluster.ClusterName,
		"Service":       *v.app.service.ServiceName,
	}).Debug("AddPage v.tablePages")

	v.tablePages.AddPage(contentPageName, contentTextItem, true, true)
}

func (v *View) handleInfoPageSwitch(entity Entity) {
	pageName := fmt.Sprintf("%s.%s", entity.entityName, v.app.secondaryKind)

	logger.WithFields(logrus.Fields{
		"Action":        "SwitchToPage",
		"PageName":      pageName,
		"Kind":          v.app.kind.String(),
		"SecondaryKind": v.app.secondaryKind.String(),
		"Cluster":       *v.app.cluster.ClusterName,
		"Service":       *v.app.service.ServiceName,
	}).Debug("SwitchToPage v.infoPages")

	v.infoPages.SwitchToPage(pageName)
}

func getContentTextItem(contentStr string, title string) *tview.TextView {
	contentText := tview.NewTextView().SetDynamicColors(true).SetText(contentStr)
	contentText.SetBorder(true).SetTitle(title).SetBorderPadding(0, 0, 1, 1)
	return contentText
}

// SSH into selected container
func (v *View) ssh(containerName string) {
	if v.app.kind != ContainerPage {
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
		shell,
	}

	logger.Infof("Exec: `%s %s`", awsCli, strings.Join(args, " "))

	v.app.Suspend(func() {
		cmd := exec.Command(bin, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		// ignore the stderr from ssh server
		_, err = cmd.Stdout.Write([]byte(fmt.Sprintf(sshBannerFmt, *v.app.cluster.ClusterName, *v.app.service.ServiceName, util.ArnToName(v.app.task.TaskArn), containerName)))
		err = cmd.Run()
		// return signal
		signal.Stop(interrupt)
		close(interrupt)
	})
}

func (v *View) buildInfoPages(items []InfoItem, title, entityName string) {
	infoFlex := v.buildInfoFlex(title, items, v.keys)
	v.infoPages.AddPage(entityName, infoFlex, true, true)

	for p, k := range v.pageKeyMap {
		infoJsonFlex := v.buildInfoFlex(title, items, k)
		v.infoPages.AddPage(fmt.Sprintf("%s.%s", entityName, p), infoJsonFlex, true, false)
	}
}
