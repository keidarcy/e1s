package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/keidarcy/e1s/internal/api"
	"github.com/keidarcy/e1s/internal/color"
	"github.com/keidarcy/e1s/internal/utils"
	"github.com/rivo/tview"
)

// Static splash: implied motion from density bars and spacing, no animation.
const splashASCII = `
 ░▒▓██▓▒░░▒▓██▓▒░░▒▓██▓▒░░▒▓██▓▒░░

   ███████╗     ██╗     ███████╗
   ██╔════╝     ██║     ██╔════╝
   █████╗       ██║     ███████╗
   ██╔══╝       ██║     ╚════██║
   ███████╗     ██║     ███████║
   ╚══════╝     ╚═╝     ╚══════╝
 ░▒▓██▓▒░░▒▓██▓▒░░▒▓██▓▒░░▒▓██▓▒░░
`

func (app *App) buildSplashPage() *tview.Flex {
	logo := strings.TrimLeft(splashASCII, "\n")
	logoTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[%s::b]%s[-:-:-]", theme.Cyan, tview.Escape(logo)))

	verTV := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[%s::b]version[-:-:-] [%s::b]%s[-:-:-]",
			theme.Magenta, theme.Yellow, tview.Escape(utils.AppVersion)))

	hint := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[%s]Loading…[-:-:-]", theme.Gray))

	col := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(logoTV, 0, 1, false).
		AddItem(verTV, 1, 0, false).
		AddItem(hint, 1, 0, false)

	// Vertical center: top/bottom spacers and middle share flex space; col must
	// have proportion > 0 — (0, 0) gives zero height so the logo never shows.
	return tview.NewFlex().
		AddItem(tview.NewBox().SetBackgroundColor(color.Color(theme.BgColor)), 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(tview.NewBox().SetBackgroundColor(color.Color(theme.BgColor)), 0, 1, false).
			AddItem(col, 0, 1, false).
			AddItem(tview.NewBox().SetBackgroundColor(color.Color(theme.BgColor)), 0, 1, false), 0, 3, true).
		AddItem(tview.NewBox().SetBackgroundColor(color.Color(theme.BgColor)), 0, 1, false)
}

func (app *App) runSplashBootstrap() {
	start := time.Now()
	store, err := api.NewStore(globalProfile, globalRegion)
	var clusters []types.Cluster
	var services []types.Service
	if err == nil {
		if app.Option.Cluster == "" {
			clusters, err = store.ListClusters()
		} else {
			cn := app.Option.Cluster
			services, err = store.ListServices(&cn)
		}
	}
	elapsed := time.Since(start)
	if err == nil && elapsed < time.Second {
		time.Sleep(time.Second - elapsed)
	}
	app.QueueUpdateDraw(func() {
		if err != nil {
			app.splashStartupErr = err
			app.Stop()
			return
		}
		app.Store = store
		if app.Option.Cluster == "" {
			app.bootstrapClusters = clusters
		} else {
			app.bootstrapServices = services
		}
		app.SetRoot(app.mainScreen, true)
		if err := app.start(); err != nil {
			app.Notice.Error(err.Error())
		}
	})
}
