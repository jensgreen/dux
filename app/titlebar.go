package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

func NewTitleBar() *views.TextBar {
	tb := views.NewTextBar()
	style := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)

	tb.SetStyle(style)
	tb.SetLeft("title left", style)
	tb.SetCenter("dux (center)", style)
	tb.SetRight("(right)", style)

	return tb
}
