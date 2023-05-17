package z2

import (
	"fmt"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
)

type Rect struct {
	X, Y Interval
}

func (rect Rect) Eq(other Rect) bool {
	return rect.X.Eq(other.X) && rect.Y.Eq(other.Y)
}

func (rect Rect) Lo() Point {
	return Point{X: rect.X.Lo, Y: rect.Y.Lo}
}

func (rect Rect) Hi() Point {
	return Point{X: rect.X.Hi, Y: rect.Y.Hi}
}

func (rect Rect) ContainsPoint(x, y int) bool {
	return rect.X.Contains(x) && rect.Y.Contains(y)
}

func (rect Rect) AsR2() r2.Rect {
	return r2.Rect{
		X: r1.Interval{Lo: float64(rect.X.Lo), Hi: float64(rect.X.Hi)},
		Y: r1.Interval{Lo: float64(rect.Y.Lo), Hi: float64(rect.Y.Hi)},
	}
}

func (rect Rect) String() string {
	return fmt.Sprintf("[Lo(%d, %d), Hi(%d, %d)]", rect.X.Lo, rect.Y.Lo, rect.X.Hi, rect.Y.Hi)
}

func SnapRoundRect(rect r2.Rect) Rect {
	return Rect{
		X: snapRoundInterval(rect.X),
		Y: snapRoundInterval(rect.Y),
	}
}

func NewRect(x int, y int, width int, height int) Rect {
	return Rect{
		X: Interval{Lo: x, Hi: width},
		Y: Interval{Lo: y, Hi: height},
	}
}
