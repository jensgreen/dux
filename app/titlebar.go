package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
)

type TitleBar struct {
	textBar        *views.TextBar
	path           string
	size           int64
	numFiles       int
	maxDepth       int
	isWalkingFiles bool
	style          tcell.Style
	spinner        *Spinner

	views.WidgetWatchers
}

func (tb *TitleBar) SetState(state dux.State) {
	tb.path = state.Treemap.File.Path
	tb.size = state.Treemap.File.Size
	tb.numFiles = state.TotalFiles
	tb.maxDepth = state.MaxDepth
	tb.isWalkingFiles = state.IsWalkingFiles

	tb.spinner.Tick()
	tb.updateText()
	tb.PostEventWidgetContent(tb)
}

func (tb *TitleBar) updateText() {
	width, _ := tb.Size()
	s := fmt.Sprintf(" %s %s (%d files)", tb.path, files.HumanizeIEC(tb.size), tb.numFiles)
	if tb.isWalkingFiles {
		s = fmt.Sprintf("%s %s", s, tb.spinner.String())
	}
	s = fmt.Sprintf("%s (%d)", s, tb.maxDepth)
	s = fmt.Sprintf("%-*v", width-1, s)

	tb.textBar.SetLeft(s, tb.style)
}

func (tb *TitleBar) Draw() {
	tb.textBar.Draw()
}

func (tb *TitleBar) Resize() {
	tb.textBar.Resize()
}

func (tb *TitleBar) HandleEvent(ev tcell.Event) bool {
	return tb.textBar.HandleEvent(ev)
}

func (tb *TitleBar) SetView(view views.View) {
	tb.textBar.SetView(view)
}

func (tb *TitleBar) Size() (int, int) {
	return tb.textBar.Size()
}

func NewTitleBar() *TitleBar {
	style := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
	text := views.NewTextBar()
	text.SetStyle(style)

	spinner := NewSpinner()

	return &TitleBar{
		textBar: text,
		style:   style,
		spinner: spinner,
	}
}
