package z2

import (
	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/geo/r1"
	"github.com/jensgreen/dux/geo/r2"
)

type Rect = geo.Rect[int]

func RectAsR2[T geo.Number](rect geo.Rect[T]) r2.Rect {
	return r2.Rect{
		X: r1.Interval{Lo: float64(rect.X.Lo), Hi: float64(rect.X.Hi)},
		Y: r1.Interval{Lo: float64(rect.Y.Lo), Hi: float64(rect.Y.Hi)},
	}
}

func SnapRoundRect(rect r2.Rect) Rect {
	return Rect{
		X: snapRoundInterval(rect.X),
		Y: snapRoundInterval(rect.Y),
	}
}

func NewRect(x int, y int, width int, height int) Rect {
	return Rect{
		X: geo.Interval[int]{Lo: x, Hi: width},
		Y: geo.Interval[int]{Lo: y, Hi: height},
	}
}
