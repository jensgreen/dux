package app

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/z2"
)

type TreemapWidget struct {
	width    int
	height   int
	appState dux.State
	treemap  intTreemap
	commands chan<- dux.Command

	view         views.View
	childWidgets []*TreemapWidget

	views.WidgetWatchers
}

func (tv *TreemapWidget) SetState(state dux.State) {
	tv.appState = state
}

func (tv *TreemapWidget) updateWidgets(treemap intTreemap) {
	widgets := make([]*TreemapWidget, len(treemap.Children))
	for i, child := range treemap.Children {
		w := &TreemapWidget{
			width:    child.Rect.X.Hi - child.Rect.X.Lo,
			height:   child.Rect.Y.Hi - child.Rect.Y.Lo,
			appState: tv.appState,
			commands: tv.commands,
			treemap:  child,
		}
		w.SetView(tv.view)
		widgets[i] = w
	}
	for _, w := range widgets {
		w.updateWidgets(w.treemap)
	}
	tv.treemap = treemap
	tv.childWidgets = widgets
}

func (tv *TreemapWidget) Draw() {
	tv.view.Clear()
	treemap := snapRoundTreemap(tv.appState.Treemap)
	tv.updateWidgets(treemap)
	isRoot := true
	tv.draw(isRoot)
}

func (tv *TreemapWidget) draw(isRoot bool) {
	tv.drawSelf(isRoot)
	for _, c := range tv.childWidgets {
		c.draw(false)
	}
}

func (tv *TreemapWidget) Resize() {
	w, h := tv.view.Size()
	tv.width = w
	tv.height = h
	tv.PostEventWidgetResize(tv)
}

func (tv *TreemapWidget) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventMouse:
		mx, my := ev.Position()
		if tv.treemap.Rect.ContainsPoint(mx, my) {
			for _, w := range tv.childWidgets {
				if w.HandleEvent(ev) {
					return true
				}
			}
			// I contain the point, but none of my children do: stop looking
			tv.commands <- dux.Select{Path: tv.treemap.File.Path}
			return true
		}
		return false
	}
	return false
}

func (tv *TreemapWidget) SetView(view views.View) {
	tv.view = view
}

func (tv *TreemapWidget) Size() (int, int) {
	return tv.width, tv.height
}

func (tv *TreemapWidget) closeHalfOpen(rect z2.Rect) z2.Rect {
	// Hi is exclusive (half-open range)
	if rect.X.Hi > rect.X.Lo {
		rect.X.Hi--
	}
	if rect.Y.Hi > rect.Y.Lo {
		rect.Y.Hi--
	}
	return rect
}

func (tv *TreemapWidget) drawBox(rect z2.Rect) {
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

func (tv *TreemapWidget) drawSelf(isRoot bool) {
	f := tv.treemap.File
	rect := tv.closeHalfOpen(tv.treemap.Rect)
	// log.Printf("Drawing %s at %v (rect: %+v)", f.Path, rect, tm.Rect)
	tv.drawBox(rect)
	tv.drawLabel(rect, f, isRoot)
}

func (tv *TreemapWidget) drawLabel(rect z2.Rect, f files.File, isRoot bool) {
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
func (tv *TreemapWidget) drawString(xrange z2.Interval, y int, s string, combc []rune, style tcell.Style) {
	i := 0
	for x := xrange.Lo; x < xrange.Hi; x++ {
		if i == len(s) {
			break
		}
		tv.view.SetContent(x, y, rune(s[i]), combc, style)
		i++
	}
}

func NewTreemapWidget(commands chan<- dux.Command) *TreemapWidget {
	tv := &TreemapWidget{
		commands: commands,
	}
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
