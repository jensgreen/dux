package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/z2"
)

type boxChars struct {
	HLine,
	LLCorner,
	LRCorner,
	Plus,
	TTee,
	RTee,
	LTee,
	BTee,
	ULCorner,
	URCorner,
	VLine rune
}

var boxCharsLight = boxChars{
	HLine:    '─',
	LLCorner: '└',
	LRCorner: '┘',
	Plus:     '┼',
	TTee:     '┬',
	RTee:     '┤',
	LTee:     '├',
	BTee:     '┴',
	ULCorner: '┌',
	URCorner: '┐',
	VLine:    '│',
}

var boxCharsHeavy = boxChars{
	HLine:    '━',
	LLCorner: '┗',
	LRCorner: '┛',
	Plus:     '╋',
	TTee:     '┳',
	RTee:     '┫',
	LTee:     '┣',
	BTee:     '┻',
	ULCorner: '┏',
	URCorner: '┓',
	VLine:    '┃',
}

var boxCharsDouble = boxChars{
	HLine:    '═',
	LLCorner: '╚',
	LRCorner: '╝',
	Plus:     '╬',
	TTee:     '╦',
	RTee:     '╣',
	LTee:     '╠',
	BTee:     '╩',
	ULCorner: '╔',
	URCorner: '╗',
	VLine:    '║',
}

type Box struct {
	isSelected bool
	view       views.View

	views.WidgetWatchers
}

func (b *Box) Draw() {
	w, h := b.view.Size()
	lo := z2.Point{X: 0, Y: 0}
	hi := z2.Point{X: w - 1, Y: h - 1}
	view := b.view

	// Upper edge
	style := b.style()
	c := b.boxChars()
	for x := lo.X + 1; x < hi.X; x++ {
		view.SetContent(x, lo.Y, c.HLine, nil, style)
	}
	// Lower edge
	for x := lo.X + 1; x < hi.X; x++ {
		view.SetContent(x, hi.Y, c.HLine, nil, style)
	}
	// Left edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		view.SetContent(lo.X, y, c.VLine, nil, style)
	}
	// Right edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		view.SetContent(hi.X, y, c.VLine, nil, style)
	}
	// Corners, clockwise from lower right
	view.SetContent(hi.X, hi.Y, c.LRCorner, nil, style)
	view.SetContent(lo.X, hi.Y, c.LLCorner, nil, style)
	view.SetContent(lo.X, lo.Y, c.ULCorner, nil, style)
	view.SetContent(hi.X, lo.Y, c.URCorner, nil, style)
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

func (b *Box) Select(isSelected bool) {
	b.isSelected = isSelected
}

func (b *Box) style() tcell.Style {
	if b.isSelected {
		return tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	return tcell.StyleDefault
}

func (b *Box) boxChars() boxChars {
	return boxCharsLight
}

func NewBox() *Box {
	return &Box{}
}
