package z2

import "github.com/golang/geo/r2"

type Point struct {
	X, Y int
}

func (pt Point) AsR2() r2.Point {
	return r2.Point{
		X: float64(pt.X),
		Y: float64(pt.Y),
	}
}
