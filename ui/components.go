package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// modal to show in the middle of screen for any usage
func (v *View) modal(p tview.Primitive, width, height int) tview.Primitive {
	m := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)

	// handle ESC key close modal
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			v.closeModal()
		case tcell.KeyCtrlZ:
			v.closeModal()
		}
		return event
	})

	return m
}

func (v *View) styledForm(title string) *tview.Form {
	f := tview.NewForm()
	// f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tcell.ColorDarkCyan).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldBackgroundColor(tcell.ColorDarkCyan).
		SetFieldTextColor(tcell.ColorOrange).
		SetBorder(true)

	// build form title, input fields
	f.SetTitle(title).SetTitleAlign(tview.AlignLeft)

	return f
}

// // deprecated
// // Replaced by notice
// // Call this function need a new goroutine
// func (v *View) errorModal(text string, duration, width, height int) {
// 	v.flashModal(fmt.Sprintf("[red::b]%s ", text), duration, width, height)
// }

// // deprecated
// // Replaced by notice
// // Call this function need a new goroutine
// func (v *View) successModal(text string, duration, width, height int) {
// 	v.flashModal(fmt.Sprintf("[green::b]%s ", text), duration, width, height)
// }

// deprecated
// Replaced by notice
// show a flash modal in a given time duration
func (v *View) flashModal(text string, duration, width, height int) {
	t := tview.NewTextView().SetDynamicColors(true).SetText(text)
	t.SetBorder(true)
	v.app.Pages.AddPage(text, v.modal(t, width, height), true, true)
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
