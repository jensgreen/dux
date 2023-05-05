package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type TreemapView struct {
	width  int
	height int
	view   views.View

	views.WidgetWatchers
}

// Draw is called to inform the widget to draw itself.  A containing
// Widget will generally call this during the application draw loop.
func (tv *TreemapView) Draw() {
	for i := 0; i < tv.height; i++ {
		tv.view.SetContent(i%4, i, 'X', nil, tcell.StyleDefault)
	}
}

// Resize is called in response to a resize of the View.  Unlike with
// other events, Resize performed by parents first, and they must
// then call their children.  This is because the children need to
// see the updated sizes from the parents before they are called.
// In general this is done *after* the views have updated.
func (tv *TreemapView) Resize() {
	w, h := tv.view.Size()
	tv.width = w
	tv.height = h
}

// HandleEvent is called to ask the widget to handle any events.
// If the widget has consumed the event, it should return true.
// Generally, events are handled by the lower layers first, that
// is for example, a button may have a chance to handle an event
// before the enclosing window or panel.
//
// Its expected that Resize events are consumed by the outermost
// Widget, and the turned into a Resize() call.
func (tv *TreemapView) HandleEvent(ev tcell.Event) bool {
	return false
}

// SetView is used by callers to set the visual context of the
// Widget.  The Widget should use the View as a context for
// drawing.
func (tv *TreemapView) SetView(view views.View) {
	tv.view = view
}

// Size returns the size of the widget (content size) as width, height
// in columns.  Layout managers should attempt to ensure that at least
// this much space is made available to the View for this Widget.  Extra
// space may be allocated on as an needed basis.
func (tv *TreemapView) Size() (int, int) {
	return tv.width, tv.height
}

func NewTreemapView() *TreemapView {
	tv := &TreemapView{}
	return tv
}
