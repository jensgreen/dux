package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/z2"
)

type TreemapWidget struct {
	width    int
	height   int
	appState dux.State
	treemap  intTreemap
	commands chan<- dux.Command

	view         views.View
	label        *FileLabel
	childWidgets []*TreemapWidget

	views.WidgetWatchers
}

func (tv *TreemapWidget) SetState(state dux.State) {
	tv.appState = state
}

func (tv *TreemapWidget) updateWidgets(isRoot bool) {
	treemap := tv.treemap
	
	tv.label.SetFile(treemap.File)
	tv.label.SetIsRoot(isRoot)
	tv.label.Select(tv.isSelected())
	tv.setLabelView(tv.view)

	widgets := make([]*TreemapWidget, len(treemap.Children))
	for i, child := range treemap.Children {
		label := NewFileLabel()
		w := &TreemapWidget{
			width:    child.Rect.X.Hi - child.Rect.X.Lo,
			height:   child.Rect.Y.Hi - child.Rect.Y.Lo,
			appState: tv.appState,
			commands: tv.commands,
			treemap:  child,
			label:    label,
		}
		w.label.SetFile(w.treemap.File)
		w.label.Select(w.isSelected())
		w.SetView(tv.view)
		widgets[i] = w
	}
	for _, w := range widgets {
		w.updateWidgets(isRoot)
	}
	tv.childWidgets = widgets
}

func (tv *TreemapWidget) Draw() {
	isRoot := true
	tv.view.Clear()
	tv.treemap = snapRoundTreemap(tv.appState.Treemap)
	tv.updateWidgets(isRoot)
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
	tv.label.Resize()
	tv.PostEventWidgetResize(tv)
}

func (tv *TreemapWidget) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *EventMouseLocal:
		mx, my := ev.LocalPosition()
		if tv.treemap.Rect.ContainsPoint(mx, my) {
			for _, w := range tv.childWidgets {
				if w.HandleEvent(ev) {
					return true
				}
			}
			// I contain the point, but none of my children do: stop looking
			tv.commands <- dux.Select{Path: tv.treemap.Path()}
			return true
		} else {
			return false
		}
	}
	return false
}

func (tv *TreemapWidget) SetView(view views.View) {
	tv.view = view
	tv.setLabelView(view)
}

func (tv *TreemapWidget) setLabelView(view views.View) {
	x := tv.treemap.Rect.X.Lo + 1
	y := tv.treemap.Rect.Y.Lo
	height := 1

	width, _ := tv.label.Size()
	availWidth := tv.treemap.Rect.X.Length() - 2 - 1 // FIXME -1 because not half-open any more
	if width > availWidth {
		width = availWidth
	}

	v := views.NewViewPort(view, x, y, width, height)
	tv.label.SetView(v)
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
	// log.Printf("Drawing %s at %v (rect: %+v)", f.Path, rect, tm.Rect)
	rect := tv.closeHalfOpen(tv.treemap.Rect)
	tv.drawBox(rect)
	tv.label.Draw()
}

func (tv *TreemapWidget) isSelected() bool {
	return tv.appState.Selection != nil && tv.appState.Selection.Path() == tv.treemap.Path()
}

func NewTreemapWidget(commands chan<- dux.Command) *TreemapWidget {
	label := NewFileLabel() 
	tv := &TreemapWidget{
		commands: commands,
		label:    label,
	}
	return tv
}

// snapRoundTreemap rounds float coordinates in a Treemap to
// discrete coordinates in an IntTreemap
func snapRoundTreemap(tm *treemap.Treemap) intTreemap {
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

func (itm *intTreemap) Path() string {
	return itm.File.Path
}
