package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
)

type TitleBar struct {
	textBar  *views.TextBar
	style    tcell.Style
	spinner  *Spinner
	commands chan<- dux.Command

	views.WidgetWatchers
}

func (tb *TitleBar) SetState(state dux.State) {
	tb.spinner.Tick()
	tb.updateText(state)
	tb.PostEventWidgetContent(tb)
}

func (tb *TitleBar) updateText(state dux.State) {
	width, _ := tb.Size()

	var f files.File
	if state.Selection == nil {
		f = state.Treemap.File
	} else {
		f = state.Selection.File
	}

	s := fmt.Sprintf(
		" %s %s (%d files)",
		f.Path,
		files.HumanizeIEC(f.Size),
		state.TotalFiles,
	)
	if state.IsWalkingFiles {
		s = fmt.Sprintf("%s %s", s, tb.spinner.String())
	}
	s = fmt.Sprintf("%s (%d)", s, state.MaxDepth)
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
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		width, height := tb.Size()
		mx, my := ev.Position()
		isClicked := mx >= 0 && mx < width && my >= 0 && my < height
		if isClicked {
			tb.commands <- dux.Deselect{}
			return true
		}
	}
	return tb.textBar.HandleEvent(ev)
}

func (tb *TitleBar) SetView(view views.View) {
	tb.textBar.SetView(view)
}

func (tb *TitleBar) Size() (int, int) {
	return tb.textBar.Size()
}

func NewTitleBar(commands chan<- dux.Command) *TitleBar {
	style := tcell.StyleDefault.Background(tcell.ColorGreen).Foreground(tcell.ColorBlack)
	text := views.NewTextBar()
	text.SetStyle(style)
	// force initial non-zero size
	text.SetLeft(" ", style)

	spinner := NewSpinner()

	return &TitleBar{
		textBar:  text,
		style:    style,
		spinner:  spinner,
		commands: commands,
	}
}
