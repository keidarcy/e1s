package ui

import (
	"flag"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/api"
	"github.com/keidarcy/e1s/util"
	"github.com/rivo/tview"
)

const (
	L = tview.AlignLeft
	C = tview.AlignCenter
	R = tview.AlignRight
)

var (
	logger   = util.Logger
	readonly = false
	version  = false
)

// Entity contains ECS resources to show
type Entity struct {
	cluster                 *types.Cluster
	service                 *types.Service
	task                    *types.Task
	container               *types.Container
	taskDefinition          *types.TaskDefinition
	events                  []types.ServiceEvent
	taskDefinitionRevisions api.TaskDefinitionRevision
	metrics                 *api.MetricsData
	autoScaling             *api.AutoScalingData
	entityName              string
}

// tview App
type App struct {
	*tview.Application
	*tview.Pages
	*api.Store
	Region string
	Entity
	readonly bool
}

func init() {
	flag.BoolVar(&readonly, "readonly", false, "Enable readonly mode")
	flag.BoolVar(&version, "version", false, "Print e1s version")
}

func newApp() *App {
	store := api.NewStore()
	region := store.Config.Region
	if len(region) == 0 {
		region = "unknown"
	}
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		Store:       store,
		Region:      region,
		readonly:    readonly,
	}
}

// Entry point of the app
func Show() error {
	flag.Parse()
	if version {
		fmt.Println("v" + util.AppVersion)
		return nil
	}
	app := newApp()
	app.initStyles()

	if err := app.showClustersPage(); err != nil {
		return err
	}

	if err := app.Application.SetRoot(app.Pages, true).Run(); err != nil {
		return err
	}
	return nil
}

// Init basic tview styles
func (app App) initStyles() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorBlack
	tview.Styles.ContrastBackgroundColor = tcell.ColorBlack
	tview.Styles.PrimaryTextColor = tcell.ColorWhite
	tview.Styles.BorderColor = tcell.ColorDarkCyan
}
