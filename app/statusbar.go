package app

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
)

type StatusBar struct {
	text     *views.Text
	commands chan<- dux.Command

	views.WidgetWatchers
}

func (sb *StatusBar) Draw() {
	sb.text.Draw()
}

func (sb *StatusBar) Resize() {
	sb.text.Resize()
}

func (sb *StatusBar) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *EventMouseLocal:
		width, height := sb.Size()
		mx, my := ev.LocalPosition()
		sb.commands <- dux.Deselect{}
		return mx >= 0 && mx < width && my >= 0 && my < height
	}
	return false
}

func (sb *StatusBar) SetView(view views.View) {
	sb.text.SetView(view)
}

func (sb *StatusBar) Size() (int, int) {
	return sb.text.Size()
}

func NewStatusBar(commands chan<- dux.Command) *StatusBar {
	style := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)
	text := views.NewText()
	text.SetStyle(style)
	text.SetAlignment(views.AlignEnd)

	help := strings.Join([]string{
		"<←↓↑→/hjkl> navigate",
		"<enter/bs> up/down",
		"<io> zoom",
		"<+-> depth",
		"<q> quit",
		// "<?> help",
	}, " | ")
	help += " "
	text.SetText(help)

	return &StatusBar{
		text:     text,
		commands: commands,
	}
}
