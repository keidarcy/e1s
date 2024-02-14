package ui

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/keidarcy/e1s/api"
	"github.com/rivo/tview"
	"github.com/sirupsen/logrus"
)

const (
	L = tview.AlignLeft
	C = tview.AlignCenter
	R = tview.AlignRight
)

var logger *logrus.Logger

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

type Option struct {
	// Read only mode indicator
	ReadOnly bool
	// Reload resources in each move
	StaleData bool
	// Basic logger
	Logger *logrus.Logger
}

// tview App
type App struct {
	*tview.Application
	*tview.Pages
	*api.Store
	// Option from cli args
	Option
	Entity
}

func newApp(option Option) (*App, error) {
	store, err := api.NewStore(option.Logger)
	if err != nil {
		return nil, err
	}
	region := store.Config.Region
	if len(region) == 0 {
		region = "unknown"
	}
	return &App{
		Application: tview.NewApplication(),
		Pages:       tview.NewPages(),
		Store:       store,
		Option:      option,
	}, nil
}

// Entry point of the app
func Show(option Option) error {
	logger = option.Logger
	app, err := newApp(option)
	if err != nil {
		return err
	}

	app.initStyles()

	if err := app.showClustersPage(false); err != nil {
		return err
	}
	logger.Debug("Started e1s")

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
