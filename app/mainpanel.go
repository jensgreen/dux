package app

import "github.com/gdamore/tcell/v2/views"

type MainPanel struct {
	views.Panel
}

func NewMainPanel(content views.Widget) *MainPanel {
	m := &MainPanel{}

	m.SetTitle(NewTitleBar())

	sb := views.NewText()
	sb.SetText("statusbar")
	m.SetStatus(sb)

	m.SetContent(content)

	return m
}
