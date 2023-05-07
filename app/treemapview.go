package app

import (
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/z2"
)

type TreemapView struct {
	width  int
	height int
	state  dux.State
	view   views.View

	views.WidgetWatchers
}

func (tv *TreemapView) SetState(state dux.State) {
	tv.state = state
}

// Draw is called to inform the widget to draw itself.  A containing
// Widget will generally call this during the application draw loop.
func (tv *TreemapView) Draw() {
	log.Printf("tv size   w=%v, h=%v", tv.width, tv.height)
	vw, vh := tv.view.Size()
	log.Printf("view size w=%v h=%v", vw, vh)
	// for i := 0; i < tv.height; i++ {
	// 	tv.view.SetContent(i%4, i, 'X', nil, tcell.StyleDefault)
	// }
	tv.view.Clear()
	tv.drawTreemapPane(tv.state)
}

// Resize is called in response to a resize of the View.  Unlike with
// other events, Resize performed by parents first, and they must
// then call their children.  This is because the children need to
// see the updated sizes from the parents before they are called.
// In general this is done *after* the views have updated.
func (tv *TreemapView) Resize() {
	w, h := tv.view.Size()
	log.Printf("Resize() from %v,%v to %v,%v", tv.width, tv.height, w, h)
	tv.width = w
	tv.height = h
	tv.PostEventWidgetResize(tv)
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
	// log.Printf("TreemapView event: %T", ev)
	// switch x := ev.(type) {
	// case *views.EventWidgetResize:
	// 	log.Printf("event resize: %T %+v", x, x.Widget())
	// case *tcell.EventMouse:
	// 	log.Printf("event mouse: %+v", x)
	// }
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

func (tv *TreemapView) Update(state dux.State) {
	log.Printf("Updating state %+v", state)
	tv.state = state
	log.Printf("Posted PostEventWidgetContent event")
	tv.Draw()
	tv.PostEventWidgetContent(tv)
}

func (tv *TreemapView) closeHalfOpen(rect z2.Rect) z2.Rect {
	// Hi is exclusive (half-open range)
	if rect.X.Hi > rect.X.Lo {
		rect.X.Hi--
	}
	if rect.Y.Hi > rect.Y.Lo {
		rect.Y.Hi--
	}
	return rect
}

func (tv *TreemapView) drawBox(rect z2.Rect) {
	lo := rect.Lo()
	hi := rect.Hi()

	// Upper edge
	style := tcell.StyleDefault
	for x := lo.X + 1; x < hi.X; x++ {
		tv.view.SetContent(x, lo.Y, tcell.RuneHLine, nil, style)
	}
	// Lower edge
	for x := lo.X + 1; x < hi.X; x++ {
		tv.view.SetContent(x, hi.Y, tcell.RuneHLine, nil, style)
	}
	// Left edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		tv.view.SetContent(lo.X, y, tcell.RuneVLine, nil, style)
	}
	// Right edge
	for y := lo.Y + 1; y < hi.Y; y++ {
		tv.view.SetContent(hi.X, y, tcell.RuneVLine, nil, style)
	}
	// Corners, clockwise from lower right
	tv.view.SetContent(hi.X, hi.Y, tcell.RuneLRCorner, nil, style)
	tv.view.SetContent(lo.X, hi.Y, tcell.RuneLLCorner, nil, style)
	tv.view.SetContent(lo.X, lo.Y, tcell.RuneULCorner, nil, style)
	tv.view.SetContent(hi.X, lo.Y, tcell.RuneURCorner, nil, style)
}

func (tv *TreemapView) drawTreemapPane(state dux.State) {
	itm := snapRoundTreemap(state.Treemap)
	isRoot := true
	tv.drawTreemap(state, itm, isRoot)
}

func (tv *TreemapView) drawTreemap(state dux.State, tm intTreemap, isRoot bool) {
	f := tm.File
	rect := tv.closeHalfOpen(tm.Rect)
	log.Printf("Drawing %s at %v (rect: %+v)", f.Path, rect, tm.Rect)
	tv.drawBox(rect)
	tv.drawLabel(rect, f, isRoot)

	for _, child := range tm.Children {
		tv.drawTreemap(state, child, false)
	}
}

func (tv *TreemapView) drawLabel(rect z2.Rect, f files.File, isRoot bool) {
	xrangeLabel := z2.Interval{
		Lo: rect.X.Lo + 1, // don't draw on left corner
		Hi: rect.X.Hi,
	}

	fname := f.Name()
	if isRoot {
		fname = f.Path
	}
	style := tcell.StyleDefault
	if f.IsDir {
		if !strings.HasSuffix(fname, "/") {
			// avoid showing for example "/" as "//"
			fname += "/"
		}
		style = style.Foreground(tcell.ColorBlue)
	}

	// apply styling to name part only
	xrangeName := xrangeLabel
	if len(fname) < xrangeName.Hi-xrangeName.Lo {
		xrangeName.Hi = xrangeName.Lo + len(fname)
	}
	// draw disk usage in renaming range, if possible
	xrangeSize := z2.Interval{
		Lo: xrangeName.Hi,
		Hi: xrangeLabel.Hi,
	}
	y := rect.Lo().Y
	tv.drawString(xrangeName, y, fname, nil, style) // different when dir
	humanSize := " " + files.HumanizeIEC(f.Size)
	if xrangeSize.Hi >= xrangeSize.Lo+len(humanSize) {
		tv.drawString(xrangeSize, y, humanSize, nil, tcell.StyleDefault)
	}
}

// drawString draws the provided string on row y and in cols given by the open interval xrange,
// truncating the string if necessary
func (tv *TreemapView) drawString(xrange z2.Interval, y int, s string, combc []rune, style tcell.Style) {
	i := 0
	for x := xrange.Lo; x < xrange.Hi; x++ {
		if i == len(s) {
			break
		}
		tv.view.SetContent(x, y, rune(s[i]), combc, style)
		i++
	}
}

func NewTreemapView() *TreemapView {
	tv := &TreemapView{}
	return tv
}

// snapRoundTreemap rounds float coordinates in a Treemap to
// discrete coordinates in an IntTreemap
func snapRoundTreemap(tm dux.Treemap) intTreemap {
	children := make([]intTreemap, len(tm.Children))
	for i, child := range tm.Children {
		children[i] = snapRoundTreemap(child)
	}

	return intTreemap{
		File:     tm.File,
		Rect:     z2.SnapRoundRect(tm.Rect),
		Children: children,
	}
}

type intTreemap struct {
	File     files.File
	Rect     z2.Rect
	Children []intTreemap
}
