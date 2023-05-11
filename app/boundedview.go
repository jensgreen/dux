package app

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/views"
)

type BoundedView interface {
	GetAbsolute() (int, int, int, int)
	GetRelative() (int, int, int, int)
	views.View
}

type boundedView struct {
	absx   int
	absy   int
	relx   int
	rely   int
	width  int
	height int
	parent BoundedView
}

func (v *boundedView) Clear() {
	v.Fill(' ', tcell.StyleDefault)
}

func (v *boundedView) Fill(ch rune, style tcell.Style) {
	if v.parent != nil {
		for y := 0; y < v.height; y++ {
			for x := 0; x < v.width; x++ {
				v.parent.SetContent(x+v.relx, y+v.rely, ch, nil, style)
			}
		}
	}
}

func (v *boundedView) Size() (int, int) {
	return v.width, v.height
}

func (v *boundedView) SetContent(x, y int, ch rune, comb []rune, s tcell.Style) {
	if v.parent == nil {
		return
	}
	if x < 0 {
		log.Panicf("SetContent: x=%d < 0", x)
	}
	if y < 0 {
		log.Panicf("SetContent: y=%d < 0", y)
	}
	if x >= (v.relx + v.width) {
		log.Panicf("SetContent: x=%d >= relx=%d + width=%d", x, v.relx, v.width)
	}
	if y >= (v.rely + v.height) {
		log.Panicf("SetContent: y=%d >= rely=%d + height=%d", y, v.rely, v.height)
	}
	v.parent.SetContent(x+v.relx, y+v.rely, ch, comb, s)
}

// GetRelative returns the upper left and lower right coordinates of the visible
// content in the coordinate space of the parent.  This is the physical
// coordinates of the screen, if the screen is the view's parent.
func (v *boundedView) GetRelative() (int, int, int, int) {
	x1 := v.relx
	y1 := v.rely
	x2 := v.relx + v.width - 1
	y2 := v.rely + v.height - 1
	return x1, y1, x2, y2
}

// GetAbsolute returns the upper left and lower right coordinates of the visible
// content in the coordinate space of the screen.
func (v *boundedView) GetAbsolute() (int, int, int, int) {
	x1 := v.absx
	y1 := v.absy
	x2 := v.absx + v.width - 1
	y2 := v.absy + v.height - 1
	return x1, y1, x2, y2
}

// Resize is called by the enclosing view to change the size of the BoundedView,
// usually in response to a window resize event.  The x, y refer are the
// BoundedView's location relative to the parent View.  A negative value for either
// width or height will cause the BoundedView to expand to fill to the end of parent
// View in the relevant dimension.
func (v *boundedView) Resize(x, y, width, height int) {
	if v.parent == nil {
		return
	}
	px, py := v.parent.Size()
	pabsx, pabsy, _, _ := v.parent.GetAbsolute()
	if x >= 0 && x < px {
		v.relx = x
		v.absx = x + pabsx
	}
	if y >= 0 && y < py {
		v.rely = y
		v.absy = y + pabsy
	}
	if width < 0 || width > px-x {
		width = px - x
	}
	if height < 0 || height > py-y {
		height = py - y
	}

	v.width = width
	v.height = height
}

// SetView is called during setup, to provide the parent View.
func (v *boundedView) SetView(view BoundedView) {
	v.parent = view
}

// NewBoundedView returns a new BoundedView (and hence also a View).
// The x and y coordinates are an offset relative to the parent.
// The origin 0,0 represents the upper left.  The width and height
// indicate a width and height. If the value -1 is supplied, then the
// dimension is calculated from the parent.
func NewBoundedView(view BoundedView, x, y, width, height int) BoundedView {
	v := &boundedView{parent: view}
	v.Resize(x, y, width, height)
	return v
}
