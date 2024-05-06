package ui

import (
	"fmt"
	"time"

	"github.com/keidarcy/e1s/internal/color"
	"github.com/rivo/tview"
)

type Notice struct {
	*tview.TextView
	app        *tview.Application
	timer      *time.Timer
	forceTimer *time.Timer
}

func NewNotice(app *tview.Application, theme color.Colors) *Notice {
	t := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true)
	t.SetBackgroundColor(color.Color(theme.BgColor))
	return &Notice{
		TextView:   t,
		app:        app,
		timer:      time.NewTimer(time.Second * 3),
		forceTimer: time.NewTimer(time.Second * 8),
	}
}

func (n *Notice) sendMessage(s string) {
	update := func() {
		n.SetText(s)
	}
	go n.app.QueueUpdateDraw(update)
	n.clear()
	n.forceClear()
}

func (n *Notice) clear() {
	n.timer.Reset(time.Second * 3)
	go func() {
		<-n.timer.C
		n.app.QueueUpdate(func() {
			n.Clear()
		})
	}()
}

func (n *Notice) forceClear() {
	n.forceTimer.Reset(time.Second * 8)
	go func() {
		<-n.forceTimer.C
		n.Clear()
		n.app.Draw()
	}()
}

func (n *Notice) Info(s string) {
	m := fmt.Sprintf(color.NoticeInfoFmt, s)
	n.sendMessage(m)
}

func (n *Notice) Warn(s string) {
	m := fmt.Sprintf(color.NoticeWarnFmt, s)
	n.sendMessage(m)
}
func (n *Notice) Error(s string) {
	m := fmt.Sprintf(color.NoticeErrorFmt, s)
	n.sendMessage(m)
}

func (n *Notice) Infof(s string, args ...interface{}) {
	m := fmt.Sprintf(fmt.Sprintf(color.NoticeInfoFmt, s), args...)
	n.sendMessage(m)
}

func (n *Notice) Warnf(s string, args ...interface{}) {
	m := fmt.Sprintf(fmt.Sprintf(color.NoticeWarnFmt, s), args...)
	n.sendMessage(m)
}
func (n *Notice) Errorf(s string, args ...interface{}) {
	m := fmt.Sprintf(fmt.Sprintf(color.NoticeErrorFmt, s), args...)
	n.sendMessage(m)
}
