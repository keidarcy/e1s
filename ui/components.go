package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// modal to show in the middle of screen for any usage
func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}

func styledForm() *tview.Form {
	f := tview.NewForm()
	// f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorDarkCyan).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldBackgroundColor(tcell.ColorDarkCyan).
		SetFieldTextColor(tcell.ColorOrange).
		SetBorder(true)
	return f
}

func (v *View) errorModal(text string) {
	v.flashModal(fmt.Sprintf("[red::b]%s ", text), 3)
}

func (v *View) successModal(text string) {
	v.flashModal(fmt.Sprintf("[green::b]%s ", text), 3)
}

// show a flash modal in a given time duration
func (v *View) flashModal(text string, duration int) {
	t := tview.NewTextView().SetDynamicColors(true).SetText(text)
	t.SetBorder(true)
	v.app.Pages.AddPage(text, modal(t, 100, 10), true, true)
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		v.closeModal()
		return event
	})
	go func() {
		time.Sleep(time.Duration(duration) * time.Second)
		v.closeModal()
		v.app.Application.Draw()
	}()
}
