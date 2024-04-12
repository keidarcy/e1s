package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

const (
	titleFmt        = "[aqua::b]%s[aqua::-]([purple::b]%d[aqua::-]) "
	nsTitleFmt      = " [aqua::-]<[purple::b]%s[aqua::-]>" + titleFmt
	contentTitleFmt = " [blue]%s([purple::b]%s[blue:-:-]) "
	infoTitleFmt    = " [blue]info([purple::b]%s[blue:-:-]) "
	keyFmt          = " [purple::b]<%s> [green:-:-]%s "
	infoItemFmt     = " %s:[aqua::b] %s "
	clusterTasksFmt = "[blue]%d Pending[-] | [green]%d Running"
	serviceTasksFmt = "%d/%d Tasks running"
	colorJSONFmt    = `%s"[steelblue::b]%s[-:-:-]": %s`

	backToPrevious                 = "Back"
	describe                       = "Describe"
	describeServiceEvents          = "Describe service events"
	describeAutoScaling            = "Describe service auto scaling"
	showTaskDefinitions            = "Show task definitions"
	showMetrics                    = "Show metrics(CPU/Memory)"
	showLogs                       = "Show cloudwatch logs(Only support awslogs logDriver)"
	realtimeLog                    = "Realtime log streaming(Only support one log group)"
	toggleFullScreen               = "Toggle full screen"
	updateService                  = "Update service"
	openInEditor                   = "Open in default editor"
	openInBrowser                  = "Open in browser"
	reloadResource                 = "Reload resources"
	portForwarding                 = "Start port forwarding session"
	terminatePortForwardingSession = "Terminate port forwarding session"
	sshContainer                   = "SSH container"
	exitContainer                  = "Exit from container"

	shell          = "/bin/sh"
	awsCli         = "aws"
	smpCi          = "session-manager-plugin"
	sshBannerFmt   = "\033[1;31m<<E1S-ECS-EXEC>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mTask\033[0m: \"%s\" \n\033[1;32mContainer\033[0m: \"%s\"\n#######################################\n"
	realtimeLogFmt = "\033[1;31m<<E1S-LOGS-TAIL>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mLogGroup\033[0m: \"%s\"\n#######################################\n"
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
	wKey  = 'w'
	FKey  = 'F'
	TKey  = 'T'
	UKey  = 'U'
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
		v.app.Notice.Info("Reloaded")
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
	v.app.showPrimaryKindPage(k, reload)
}

func (v *View) showSecondaryKindPage(reload bool) {
	switch v.app.secondaryKind {
	case AutoScalingKind:
		v.switchToAutoScalingJson()
	case DescriptionKind:
		v.switchToDescriptionJson()
	case LogKind:
		v.switchToLogsList()
	case ServiceEventsKind:
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
	v.app.secondaryKind = EmptyKind
	if v.app.cluster == nil {
		v.app.Stop()
		return
	}
	// v.app.secondaryKind = EmptyKind
	toPage := v.app.kind.getAppPageName(v.app.getPageHandle())
	v.app.Pages.SwitchToPage(toPage)
}

// Content page builder
func (v *View) handleContentPageSwitch(entity Entity, colorizedJsonString string, jsonBytes []byte) {
	contentTitle := fmt.Sprintf(contentTitleFmt, v.app.kind, entity.entityName)
	contentPageName := v.app.kind.getContentPageName(entity.entityName + "." + v.app.secondaryKind.String())

	contentTextItem := getContentTextItem(colorizedJsonString, contentTitle)

	// press f toggle json
	contentTextItem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fullScreenContent := getContentTextItem(colorizedJsonString, contentTitle)

		// full screen json press ESC close full screen json and back to table
		fullScreenContent.SetDoneFunc(func(key tcell.Key) {
			v.handleFullScreenContentDone()
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
			if v.app.secondaryKind == LogKind {
				v.realtimeAwsLog(entity)
			}
		case eKey:
			if v.app.secondaryKind == DescriptionKind || v.app.secondaryKind == AutoScalingKind {
				v.openInEditor(jsonBytes)
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

func (v *View) buildInfoPages(items []InfoItem, title, entityName string) {
	infoFlex := v.buildInfoFlex(title, items, v.keys)
	v.infoPages.AddPage(entityName, infoFlex, true, true)

	for p, k := range v.pageKeyMap {
		infoJsonFlex := v.buildInfoFlex(title, items, k)
		v.infoPages.AddPage(fmt.Sprintf("%s.%s", entityName, p), infoJsonFlex, true, false)
	}
}
