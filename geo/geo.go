package geo

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type numeric interface {
	constraints.Integer | constraints.Float
}

// Interval

type Interval[T numeric] struct {
	Lo, Hi T
}

func (i Interval[T]) Eq(other Interval[T]) bool {
	return i.Lo == other.Lo && i.Hi == other.Hi
}

func (i Interval[T]) Length() T {
	return i.Hi - i.Lo
}

func (r Interval[T]) IsEmpty() bool {
	return r.Lo == r.Hi
}

// ContainsHalfClosed is true if x is in the half-closed interval [lo, hi)
func (i Interval[T]) ContainsHalfClosed(x T) bool {
	return i.Lo <= x && x < i.Hi
}

// ContainsClosed is true if x is in the closed interval [lo, hi]
func (i Interval[T]) ContainsClosed(x T) bool {
	return i.Lo <= x && x <= i.Hi
}

func IntervalFromPoint[T numeric](pt T) Interval[T] {
	return Interval[T]{Lo: pt, Hi: pt}
}

func (i Interval[T]) String() string {
	return fmt.Sprintf("[Lo(%v), Hi(%v)]", i.Lo, i.Hi)
}

// Point

type Point[T numeric] struct {
	X, Y T
}

func NewPoint[T numeric](x, y T) Point[T] {
	return Point[T]{X: x, Y: y}
}

func (p Point[T]) String() string {
	return fmt.Sprintf("[X(%v), Hi(%v)]", p.X, p.Y)
}

// Rect

type Rect[T numeric] struct {
	X, Y Interval[T]
}

func NewRect[T numeric](x, y, width, height T) Rect[T] {
	return Rect[T]{
		X: Interval[T]{Lo: x, Hi: width},
		Y: Interval[T]{Lo: y, Hi: height},
	}
}

func (r Rect[T]) Eq(other Rect[T]) bool {
	return r.X.Eq(other.X) && r.Y.Eq(other.Y)
}

func (r Rect[T]) Lo() Point[T] {
	return Point[T]{X: r.X.Lo, Y: r.Y.Lo}
}

func (r Rect[T]) Hi() Point[T] {
	return Point[T]{X: r.X.Hi, Y: r.Y.Hi}
}

func (r Rect[T]) IsEmpty() bool {
	return r.X.IsEmpty() && r.Y.IsEmpty()
}

func (r Rect[T]) Size() Point[T] {
	return Point[T]{
		X: r.X.Hi - r.X.Lo,
		Y: r.Y.Hi - r.Y.Lo,
	}

}

func (r Rect[T]) ContainsHalfClosed(pt Point[T]) bool {
	return r.X.ContainsHalfClosed(pt.X) && r.Y.ContainsHalfClosed(pt.Y)
}

func (r Rect[T]) ContainsClosed(pt Point[T]) bool {
	return r.X.ContainsClosed(pt.X) && r.Y.ContainsClosed(pt.Y)
}

func (r Rect[T]) String() string {
	return fmt.Sprintf("[Lo(%v, %v), Hi(%v, %v)]", r.X.Lo, r.Y.Lo, r.X.Hi, r.Y.Hi)
}
