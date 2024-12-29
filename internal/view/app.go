package view

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/internal/api"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/ui"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

var theme color.Colors
var ErrNoNeedReload = errors.New("no need reload")

// Entity contains ECS resources to show, use uppercase to make items like app.cluster easy to access
type Entity struct {
	cluster           *types.Cluster
	service           *types.Service
	task              *types.Task
	container         *types.Container
	taskDefinition    *types.TaskDefinition
	events            []types.ServiceEvent
	metrics           *api.MetricsData
	autoScaling       *api.AutoScalingData
	serviceDeployment *types.ServiceDeployment
	serviceRevision   *types.ServiceRevision
	entityName        string
}

type Option struct {
	// Read only mode indicator
	ReadOnly bool
	// Reload resources every x second(s), -1 is stop auto refresh
	Refresh int
	// ECS exec shell
	Shell string
	// Here for help view
	Debug bool
	// Here for help view
	JSON bool
	// Here for help view
	LogFile string
	// Here for help view
	ConfigFile string
	// Here for help view
	Theme string
	// Default cluster name
	Cluster string
}

// tview App
type App struct {
	// tview Application
	*tview.Application
	// Info + table area pages UI for MainScreen
	*tview.Pages
	// Notice text UI in MainScreen footer
	Notice *ui.Notice
	// mainScreen content UI
	mainScreen *tview.Flex
	// API client
	*api.Store
	// Option from cli args
	Option
	// Current screen item content, use uppercase to make items like app.cluster easy to access
	Entity
	// Current page primary kind ex: cluster, service
	kind kind
	// Current secondary kind like json, list
	secondaryKind kind
	// Track back kind when necessary
	backKind kind
	// Port forwarding ssm session Id
	sessions []*PortForwardingSession
	// Current primary kind table row index for auto refresh to keep row selected
	rowIndex int
	// Specify in tview app suspend or not
	isSuspended bool
	// Show selected status tasks
	taskStatus types.DesiredStatus
	// Show resources from cluster
	fromCluster bool
	// AWS region
	region string
	// AWS profile
	profile string
}

func newApp(option Option) (*App, error) {
	store, err := api.NewStore()
	if err != nil {
		return nil, err
	}
	app := tview.NewApplication()
	pages := tview.NewPages()
	footer := tview.NewFlex()

	notice := ui.NewNotice(app, theme)
	footer.AddItem(notice, 0, 1, false)
	main := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(pages, 0, 2, true).
		AddItem(footer, 1, 1, false)

	return &App{
		Application:   app,
		Pages:         pages,
		Notice:        notice,
		mainScreen:    main,
		Store:         store,
		Option:        option,
		kind:          ClusterKind,
		secondaryKind: EmptyKind,
		backKind:      EmptyKind,
		taskStatus:    types.DesiredStatusRunning,
		region:        store.Region,
		profile:       store.Profile,
		Entity: Entity{
			cluster: &types.Cluster{
				ClusterName: aws.String("e1s_default_cluster"),
				ClusterArn:  aws.String("e1s_default_cluster_arn"),
			},
			service: &types.Service{
				ServiceName: aws.String("e1s_default_service"),
				ServiceArn:  aws.String("e1s_default_service arn"),
			},
			task:           &types.Task{},
			container:      &types.Container{},
			taskDefinition: &types.TaskDefinition{},
		},
	}, nil
}

// Entry point of the app
func Start(option Option) error {
	file := utils.GetLogger(option.LogFile, option.JSON, option.Debug)
	defer file.Close()
	slog.Debug(`
****************************************************************
**************** Started e1s
****************************************************************`)
	slog.Debug("e1s start", "option", option)
	theme = color.InitStyles(option.Theme)

	app, err := newApp(option)
	if err != nil {
		return err
	}

	if err := app.start(); err != nil {
		return err
	}

	app.SetInputCapture(app.globalInputHandle)

	if err := app.Application.SetRoot(app.mainScreen, true).Run(); err != nil {
		return err
	}
	app.onClose()
	return nil
}

// Add new page to app.Pages
func (app *App) addAppPage(page *tview.Flex) {
	pageName := app.kind.getAppPageName(app.getPageHandle())

	slog.Debug("app.Pages navigation", "action", "AppPage", "pageName", pageName, "app", app)

	app.Pages.AddPage(pageName, page, true, true)
}

// Switch app.Pages page
func (app *App) switchPage(reload bool) bool {
	pageName := app.kind.getAppPageName(app.getPageHandle())
	if app.Pages.HasPage(pageName) && app.Refresh < 0 && !reload {

		slog.Debug("app.Pages navigation", "action", "SwitchToPage", "pageName", pageName, "app", app)
		app.Pages.SwitchToPage(pageName)
		return true
	}
	return false
}

// Go back page based on current kind
func (app *App) back() {
	app.taskStatus = types.DesiredStatusRunning

	prevKind := app.kind.prevKind()
	if app.backKind != EmptyKind {
		prevKind = app.backKind
		app.backKind = EmptyKind
	}

	if app.fromCluster && prevKind == ServiceKind {
		app.fromCluster = false
		prevKind = ClusterKind
	}

	app.kind = prevKind
	app.secondaryKind = EmptyKind
	pageName := prevKind.getAppPageName(app.getPageHandle())

	slog.Debug("app.Pages navigation", "action", "back", "pageName", pageName, "app", app)

	if prevKind == ClusterKind && app.Option.Cluster != "" {
		app.Option.Cluster = ""
		err := app.showPrimaryKindPage(ClusterKind, false)
		if err != nil {
			app.Notice.Warn("failed to back to cluster list")
		}
		return
	}

	app.Pages.SwitchToPage(pageName)
}

// Get page handler, cluster is empty, other is cluster arn
func (app *App) getPageHandle() string {
	name := ""
	if app.kind != ClusterKind {
		name = *app.cluster.ClusterArn
	}
	// based on different task status different name
	if app.kind == TaskKind {
		name = name + "." + strings.ToLower(string((app.taskStatus)))
	}

	// true when show tasks in cluster
	if app.fromCluster {
		name = name + ".cluster"
	}
	return name
}

func (app *App) start() error {
	var err error
	if app.Option.Cluster == "" {
		err = app.showPrimaryKindPage(ClusterKind, false)
	} else {
		app.cluster.ClusterName = &app.Option.Cluster
		err = app.showPrimaryKindPage(ServiceKind, false)
	}

	if app.Option.Refresh > 0 {
		slog.Debug("Auto refresh rate", "seconds", app.Option.Refresh)
		ticker := time.NewTicker(time.Duration(app.Option.Refresh) * time.Second)

		go func() {
			for {
				<-ticker.C
				if app.secondaryKind == EmptyKind && !app.isSuspended {
					app.showPrimaryKindPage(app.kind, true)
					slog.Debug("Auto refresh")
					app.Application.Draw()
				}
			}
		}()
	}
	return err
}

// Show Primary kind page
func (app *App) showPrimaryKindPage(k kind, reload bool) error {
	var err error
	if k == TaskDefinitionKind {
		app.backKind = app.kind
	}
	app.kind = k
	switch k {
	case ClusterKind:
		err = app.showClustersPage(reload)
	case ServiceKind:
		err = app.showServicesPage(reload)
	case TaskKind:
		err = app.showTasksPages(reload)
	case ContainerKind:
		err = app.showContainersPage(reload)
	case TaskDefinitionKind:
		err = app.showTaskDefinitionPage(reload)
	case ServiceDeploymentKind:
		err = app.showServiceDeploymentPage(reload)
	default:
		app.kind = ClusterKind
		err = app.showClustersPage(reload)
	}
	if err != nil {
		app.Notice.Error(err.Error())
		return err
	}
	if !reload {
		if app.taskStatus != types.DesiredStatusStopped {
			app.Notice.Infof("Viewing %s...", app.kind.String())
		}
	} else {
		slog.Debug("Reload in showPrimaryKindPage")
	}
	return nil
}

// E1s app close hook
func (app *App) onClose() {
	if len(app.sessions) != 0 {
		ids := []*string{}
		for _, s := range app.sessions {
			ids = append(ids, s.sessionId)
		}
		err := app.Store.TerminateSessions(ids)
		if err != nil {
			slog.Error("Failed to terminated port forwarding sessions", "error", err)
		} else {
			slog.Debug("Terminated port forwarding session terminated")
		}
	}

	slog.Debug(`
**************** Exited e1s ************************************`)
}

func (app *App) globalInputHandle(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case '?':
		app.showHelpPage()
	}
	return event
}

func (app *App) LogValue() slog.Value {
	return slog.AnyValue(struct {
		kind          string
		secondaryKind string
		cluster       string
		service       string
	}{
		kind:          app.kind.String(),
		secondaryKind: app.secondaryKind.String(),
		cluster:       *app.cluster.ClusterName,
		service:       *app.service.ServiceName,
	})
}
