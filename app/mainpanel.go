package app

import (
	"github.com/gdamore/tcell/v2/views"
)

type MainPanel struct {
	tv        *TreemapView
	views.Panel
}

func NewMainPanel(tv *TreemapView) *MainPanel {
	m := &MainPanel{}

	m.SetTitle(NewTitleBar())

	sb := views.NewText()
	sb.SetText("statusbar")
	m.SetStatus(sb)

	m.tv = tv
	m.SetContent(m.tv)

	return m
}
