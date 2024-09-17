package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// modal to show in the middle of screen for any usage
func Modal(p tview.Primitive, width, height, top int, closeFn func()) tview.Primitive {
	m := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, top, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)

	// handle ESC key close modal
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			closeFn()
		case tcell.KeyCtrlZ:
			closeFn()
		}
		return event
	})

	return m
}

func StyledForm(title string) *tview.Form {
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
