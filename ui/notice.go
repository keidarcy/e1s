package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/rivo/tview"
)

const (
	infoFmt  = "✅ [green::]"
	warnFmt  = "😔 [orange::]"
	errorFmt = "💥 [red::]"
	closeFmt = "[-:-:-]"
)

type Notice struct {
	*tview.TextView
	app *tview.Application
}

func newNotice(app *tview.Application) *Notice {
	t := tview.NewTextView().
		SetTextAlign(C).
		SetDynamicColors(true)
	return &Notice{
		TextView: t,
		app:      app,
	}
}

func (n *Notice) sendMessage(s string) {
	update := func() {
		t := strings.TrimSpace(n.GetText(false))
		if t != infoFmt+reloadText+closeFmt {
			n.SetText(s)
		}
	}
	clear := func() {
		n.app.QueueUpdate(func() {
			n.SetText("")
		})
	}
	forceClear := func() {
		n.SetText("")
		n.app.Draw()
	}
	go n.app.QueueUpdateDraw(update)
	time.AfterFunc(2*time.Second, clear)
	time.AfterFunc(8*time.Second, forceClear)
}

func (n *Notice) Info(s string) {
	m := infoFmt + s + closeFmt
	n.sendMessage(m)
}

func (n *Notice) Warn(s string) {
	m := warnFmt + s + closeFmt
	n.sendMessage(m)
}
func (n *Notice) Error(s string) {
	m := errorFmt + s + closeFmt
	n.sendMessage(m)
}

func (n *Notice) Infof(s string, args ...interface{}) {
	m := fmt.Sprintf(infoFmt+s+closeFmt, args...)
	n.sendMessage(m)
}

func (n *Notice) Warnf(s string, args ...interface{}) {
	m := fmt.Sprintf(warnFmt+s+closeFmt, args...)
	n.sendMessage(m)
}
func (n *Notice) Errorf(s string, args ...interface{}) {
	m := fmt.Sprintf(errorFmt+s+closeFmt, args...)
	n.sendMessage(m)
}
