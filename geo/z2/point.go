package z2

import (
	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/geo/r2"
)

type Point = geo.Point[int]

func PointAsR2(pt Point) r2.Point {
	return r2.Point{
		X: float64(pt.X),
		Y: float64(pt.Y),
	}
}
