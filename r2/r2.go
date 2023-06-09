package r2

import (
	"math"

	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/r1"
)

type Point = geo.Point[float64]
type Rect = geo.Rect[float64]

func RectFromPoints(ul, lr Point) Rect {
	return Rect{
		X: r1.Interval{Lo: ul.X, Hi: lr.X},
		Y: r1.Interval{Lo: ul.Y, Hi: lr.Y},
	}
}

const epsilon = 1e-15

func RectApproxEqual(a, b Rect) bool {
	return math.Abs(a.X.Lo-b.X.Lo) < epsilon &&
		math.Abs(a.X.Hi-b.X.Hi) < epsilon &&
		math.Abs(a.Y.Lo-b.Y.Lo) < epsilon &&
		math.Abs(a.Y.Hi-b.Y.Hi) < epsilon
}
