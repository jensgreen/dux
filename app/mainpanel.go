package app

import (
	"github.com/gdamore/tcell/v2/views"
)

type MainPanel struct {
	tv        *TreemapView
	views.Panel
}

func NewMainPanel(title *TitleBar, tv *TreemapView) *MainPanel {
	m := &MainPanel{}

	sb := views.NewText()
	sb.SetText("statusbar")
	m.SetStatus(sb)

	m.SetTitle(title)

	m.tv = tv
	m.SetContent(m.tv)

	return m
}
