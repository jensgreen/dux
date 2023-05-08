package app

import (
	"github.com/gdamore/tcell/v2/views"
)

type MainPanel struct {
	views.Panel
}

func NewMainPanel(title *TitleBar, tv *TreemapView) *MainPanel {
	m := &MainPanel{}
	m.SetTitle(title)
	m.SetContent(tv)
	return m
}
