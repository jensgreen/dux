package app

import (
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
	"github.com/jensgreen/dux/dux"
	"github.com/jensgreen/dux/treemap"
)

type TreemapWidget struct {
	width    int
	height   int
	appState dux.State
	treemap  treemap.Z2Treemap
	commands chan<- dux.Command

	view         views.View
	box          *Box
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
	tv.box.Select(tv.isSelected())
	tv.setBoxView(tv.view)

	widgets := make([]*TreemapWidget, len(treemap.Children))
	for i, child := range treemap.Children {
		w := &TreemapWidget{
			width:    child.Rect.X.Hi - child.Rect.X.Lo,
			height:   child.Rect.Y.Hi - child.Rect.Y.Lo,
			appState: tv.appState,
			commands: tv.commands,
			treemap:  *child,
			box:      NewBox(),
			label:    NewFileLabel(),
		}
		w.label.SetFile(w.treemap.File)
		w.label.Select(w.isSelected())
		w.box.Select(w.isSelected())
		w.SetView(tv.view)
		widgets[i] = w
	}
	for _, w := range widgets {
		w.updateWidgets(false)
	}
	tv.childWidgets = widgets
}

func (tv *TreemapWidget) Draw() {
	isRoot := true
	tv.view.Clear()
	tv.treemap = *treemap.NewZ2Treemap(tv.appState.Treemap)
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
			if tv.appState.Selection != nil && tv.appState.Selection.Path() == tv.treemap.Path() {
				tv.commands <- dux.ZoomIn{}
			} else {
				tv.commands <- dux.Select{Path: tv.treemap.Path()}
			}
			return true
		} else {
			return false
		}
	}
	return false
}

func (tv *TreemapWidget) SetView(view views.View) {
	tv.view = view
	tv.setBoxView(view)
	tv.setLabelView(view)
}

func (tv *TreemapWidget) setBoxView(view views.View) {
	rect := tv.treemap.Rect
	tv.box.SetView(views.NewViewPort(view,
		rect.X.Lo,
		rect.Y.Lo,
		rect.X.Length(),
		rect.Y.Length(),
	))
}

func (tv *TreemapWidget) setLabelView(view views.View) {
	x := tv.treemap.Rect.X.Lo + 1
	y := tv.treemap.Rect.Y.Lo
	height := 1

	width, _ := tv.label.Size()
	cornerSize := 1
	availWidth := tv.treemap.Rect.X.Length() - 2*cornerSize
	if width > availWidth {
		width = availWidth
	}

	v := views.NewViewPort(view, x, y, width, height)
	tv.label.SetView(v)
}

func (tv *TreemapWidget) Size() (int, int) {
	return tv.width, tv.height
}

func (tv *TreemapWidget) drawSelf(isRoot bool) {
	tv.box.Draw()
	tv.label.Draw()
}

func (tv *TreemapWidget) isSelected() bool {
	return tv.appState.Selection != nil && tv.appState.Selection.Path() == tv.treemap.Path()
}

func NewTreemapWidget(commands chan<- dux.Command) *TreemapWidget {
	label := NewFileLabel()
	box := NewBox()
	tv := &TreemapWidget{
		commands: commands,
		label:    label,
		box:      box,
	}
	return tv
}
