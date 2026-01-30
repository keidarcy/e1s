package view

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

const (
	awsCli            = "aws"
	smpCi             = "session-manager-plugin"
	execBannerFmt     = "\n\033[1;31m<<E1S-CONTAINER-SHELL>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mTask\033[0m: \"%s\" \n\033[1;32mContainer\033[0m: \"%s\"\n#######################################\n"
	instanceBannerFmt = "\n\033[1;31m<<E1S-INSTANCE-SHELL>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mInstance\033[0m: \"%s\"\n#######################################\n"
	realtimeLogFmt    = "\n\033[1;31m<<E1S-LOGS-TAIL>>\033[0m: \n#######################################\n\033[1;32mCluster\033[0m: \"%s\" \n\033[1;32mService\033[0m: \"%s\" \n\033[1;32mLogGroup\033[0m: \"%s\"\n\033[1;32mLogStreamNames\033[0m: \"%s\"\n#######################################\n"
)

// Base struct of different views
type view struct {
	app         *App
	table       *tview.Table
	searchLast  *string
	headerPages *tview.Pages
	bodyPages   *tview.Pages
	keys        []keyDescriptionPair
	footer      *footer
	pageKeyMap  secondaryPageKeyMap
	// Filter support
	filterText   string
	filterInput  *tview.InputField
	filterFlex   *tview.Flex // Container for filter input (1 row)
	mainFlex     *tview.Flex // Main page flex (to add/remove filter dynamically)
	filterActive bool        // Whether filter input is currently shown
}

func newView(app *App, keys []keyDescriptionPair, pageKeys secondaryPageKeyMap) *view {
	v := &view{
		app:         app,
		headerPages: tview.NewPages(),
		bodyPages:   tview.NewPages(),
		table:       tview.NewTable(),
		searchLast:  new(string),
		keys:        keys,
		footer:      newFooter(),
		pageKeyMap:  pageKeys,
		filterText:  "",
	}
	v.initFilterInput()
	return v
}

// Initialize filter input component
func (v *view) initFilterInput() {
	v.filterInput = tview.NewInputField().
		SetLabel("🐱 /")
	// SetLabelColor(color.Color(theme.Cyan)).
	// SetFieldWidth(0).
	// SetFieldBackgroundColor(color.Color(theme.BgColor)).
	// SetFieldTextColor(color.Color(theme.FgColor))
	v.filterInput.SetBorderPadding(0, 0, 1, 1)
	// 	SetBorder(true).
	// 	// SetTitle(title).

	v.filterInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// Apply filter and hide input
			v.filterText = v.filterInput.GetText()
			v.hideFilterInput()
			v.applyFilter()
		case tcell.KeyEsc:
			// Clear filter and hide input
			v.filterText = ""
			v.filterInput.SetText("")
			v.hideFilterInput()
			v.applyFilter()
		}
	})

	// Create flex container for filter input (1 row when shown)
	v.filterFlex = tview.NewFlex().SetDirection(tview.FlexColumn)
	v.filterFlex.SetBackgroundColor(color.Color(theme.BgColor))
	v.filterFlex.AddItem(v.filterInput, 0, 1, true)
}

// Apply filter and rebuild table
func (v *view) applyFilter() {
	// Save filter to app for persistence across reloads
	v.app.kindFilters[v.app.kind] = v.filterText
	// Reset row index when filtering to avoid out of bounds
	v.app.rowIndex = 1
	// Trigger reload to rebuild table with filter
	v.reloadResource(false)
}

// Load filter from app storage
func (v *view) loadFilter() {
	if filter, ok := v.app.kindFilters[v.app.kind]; ok {
		v.filterText = filter
		v.filterInput.SetText(filter)
	}
}

// Show filter input and focus it
func (v *view) showFilterInput() {
	if v.filterActive {
		return
	}
	v.filterActive = true
	v.filterInput.SetText(v.filterText)

	// Dynamically insert filter row into main flex (after header, before body)
	// mainFlex layout: [header, body, footer] -> [header, filter, body, footer]
	if v.mainFlex != nil {
		// Remove body and footer, add filter, then re-add body and footer
		v.mainFlex.RemoveItem(v.bodyPages)
		v.mainFlex.RemoveItem(v.footer.footerFlex)
		v.mainFlex.AddItem(v.filterFlex, 1, 0, false)
		v.mainFlex.AddItem(v.bodyPages, 0, 2, true)
		v.mainFlex.AddItem(v.footer.footerFlex, 1, 1, false)
	}

	v.app.SetFocus(v.filterInput)
}

// Hide filter input and return focus to table
func (v *view) hideFilterInput() {
	if !v.filterActive {
		return
	}
	v.filterActive = false

	// Remove filter row from main flex
	if v.mainFlex != nil {
		v.mainFlex.RemoveItem(v.filterFlex)
	}

	v.app.SetFocus(v.table)
}

// Check if a row matches the filter
func (v *view) matchesFilter(row []string) bool {
	if v.filterText == "" {
		return true
	}
	filterLower := strings.ToLower(v.filterText)
	for _, cell := range row {
		if strings.Contains(strings.ToLower(cell), filterLower) {
			return true
		}
	}
	return false
}

// Interface to show each view
type dataView interface {
	headerBuilder() *tview.Pages
	bodyBuilder() *tview.Pages
	footerBuilder() *tview.Flex
}

// Common function to build page for each view
func buildAppPage(v dataView) *tview.Flex {
	// build table reference first
	tablePages := v.bodyBuilder()
	infoPages := v.headerBuilder()
	footer := v.footerBuilder()

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(infoPages, oneColumnCount+2, 1, false).
		AddItem(tablePages, 0, 2, true).
		AddItem(footer, 1, 1, false)
	return flex
}

// Build page with filter support (filter row added dynamically when needed)
func buildAppPageWithFilter(dv dataView, v *view) *tview.Flex {
	// build table reference first
	tablePages := dv.bodyBuilder()
	infoPages := dv.headerBuilder()
	footer := dv.footerBuilder()

	// Build main flex WITHOUT filter row by default
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(infoPages, oneColumnCount+2, 1, false).
		AddItem(tablePages, 0, 2, true).
		AddItem(footer, 1, 1, false)

	// Store reference so we can add/remove filter dynamically
	v.mainFlex = flex

	return flex
}

// Get current table selection and return as entity
func (v *view) getCurrentSelection() (Entity, error) {
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
		v.app.Notice.Warnf("unexpected error in getCurrentSelection: %v (%T)", entity, entity)
		return Entity{}, fmt.Errorf("unexpected error in getCurrentSelection: %v (%T)", entity, entity)
	}
}

// Reload current resource
func (v *view) reloadResource(reloadNotice bool) error {
	if reloadNotice {
		v.app.Notice.Info("Reloaded")
	}
	v.showKindPage(v.app.kind, true)
	return nil
}

// Show kind page including primary kind, secondary kind
func (v *view) showKindPage(k kind, reload bool) {
	if v.app.secondaryKind != EmptyKind {
		v.showSecondaryKindPage(reload)
		return
	}
	v.app.showPrimaryKindPage(k, reload)
}

func (v *view) showSecondaryKindPage(reload bool) {
	switch v.app.secondaryKind {
	case AutoScalingKind:
		v.switchToAutoScalingJson()
	case DescriptionKind:
		v.switchToDescriptionJson()
	case LogKind:
		v.switchToLogsList()
	case ServiceEventsKind:
		v.switchToServiceEventsList()
	case ServiceRevisionKind:
		v.switchToServiceRevisionJson()
	}
	if !reload {
		v.app.Notice.Infof("Viewing %s...", v.app.secondaryKind.String())
	} else {
		slog.Debug("Reload", "showSecondaryKindPage", reload)
	}
}

// Go current page based on current kind
func (v *view) closeModal() {
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
func (v *view) handleSecondaryPageSwitch(entity Entity, colorizedJsonString string, jsonBytes []byte) {
	contentTitle := fmt.Sprintf(color.TableSecondaryTitleFmt, v.app.kind, entity.entityName, v.app.secondaryKind)
	contentPageName := v.app.kind.getSecondaryPageName(entity.entityName + "." + v.app.secondaryKind.String())

	contentTextItem := getSecondaryTextItem(colorizedJsonString, contentTitle)

	// press f toggle json
	contentTextItem.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		fullScreenContent := getSecondaryTextItem(colorizedJsonString, contentTitle)

		// full screen json press ESC close full screen json and back to table
		fullScreenContent.SetDoneFunc(func(key tcell.Key) {
			v.handleFullScreenContentDone()
			v.handleTableContentDone(key)
		})

		// full screen json press f close full screen json
		fullScreenContent.SetInputCapture(v.handleFullScreenContentInput(jsonBytes))

		// contentTextComponent press f open in full screen
		switch event.Rune() {
		case 'f':
			v.app.Pages.AddPage(contentPageName, fullScreenContent, true, true)
		case 'b':
			v.openInBrowser()
		case 'r':
			if v.app.secondaryKind == LogKind {
				v.realtimeAwsLog(entity)
			}
		case 'e':
			if v.app.secondaryKind == DescriptionKind || v.app.secondaryKind == AutoScalingKind || v.app.secondaryKind == ServiceRevisionKind || v.app.secondaryKind == LogKind {
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

	slog.Debug("v.tablePages navigation", "action", "AppPage", "pageName", contentPageName, "app", v.app)

	v.bodyPages.AddPage(contentPageName, contentTextItem, true, true)
}

func (v *view) handleHeaderPageSwitch(entity Entity) {
	pageName := fmt.Sprintf("%s.%s", entity.entityName, v.app.secondaryKind)

	slog.Debug("v.tablePages navigation", "action", "SwitchToPage", "pageName", pageName, "app", v.app)

	v.headerPages.SwitchToPage(pageName)
}

func (v *view) buildHeaderPages(items []headerItem, title, entityName string) {
	infoFlex := v.buildHeaderFlex(title, items, v.keys)
	v.headerPages.AddPage(entityName, infoFlex, true, true)

	for p, k := range v.pageKeyMap {
		infoJsonFlex := v.buildHeaderFlex(title, items, k)
		v.headerPages.AddPage(fmt.Sprintf("%s.%s", entityName, p), infoJsonFlex, true, false)
	}
}

func getSecondaryTextItem(contentStr string, title string) *tview.TextView {
	contentText := tview.NewTextView().SetDynamicColors(true).SetText(contentStr)
	contentText.SetBorder(true).SetTitle(title).SetBorderPadding(0, 0, 1, 1)
	return contentText
}
