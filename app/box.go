package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/z2"
)

type Box struct {
	view views.View

	views.WidgetWatchers
}

func (b *Box) Draw() {
	w, h := b.view.Size()
	lo := z2.Point{X: 0, Y: 0}
	hi := z2.Point{X: w - 1, Y: h - 1}
	view := b.view

	// Upper edge
	style := tcell.StyleDefault
	for x := lo.X + 1; x < hi.X; x++ {
		view.SetContent(x, lo.Y, tcell.RuneHLine, nil, style)
	}
	// Lower edge
	for x := lo.X + 1; x < hi.X; x++ {
		view.SetContent(x, hi.Y, tcell.RuneHLine, nil, style)
	}
	// Left edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		view.SetContent(lo.X, y, tcell.RuneVLine, nil, style)
	}
	// Right edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		view.SetContent(hi.X, y, tcell.RuneVLine, nil, style)
	}
	// Corners, clockwise from lower right
	view.SetContent(hi.X, hi.Y, tcell.RuneLRCorner, nil, style)
	view.SetContent(lo.X, hi.Y, tcell.RuneLLCorner, nil, style)
	view.SetContent(lo.X, lo.Y, tcell.RuneULCorner, nil, style)
	view.SetContent(hi.X, lo.Y, tcell.RuneURCorner, nil, style)
}

func (b *Box) Resize() {
	b.PostEventWidgetResize(b)
}

func (b *Box) HandleEvent(ev tcell.Event) bool {
	return false
}

func (b *Box) SetView(view views.View) {
	b.view = view
}

func (b *Box) Size() (int, int) {
	return b.view.Size()
}

func NewBox() *Box {
	return &Box{}
}
