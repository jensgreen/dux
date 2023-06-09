package geo

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

// Interval

type Interval[T Number] struct {
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

func (i Interval[T]) Contains(x T) bool {
	return i.Lo <= x && x < i.Hi
}

func (i Interval[T]) String() string {
	return fmt.Sprintf("[Lo(%v), Hi(%v)]", i.Lo, i.Hi)
}

func IntervalFromPoint[T Number](pt T) Interval[T] {
	return Interval[T]{Lo: pt, Hi: pt}
}

// Point

type Point[T Number] struct {
	X, Y T
}

// Rect

type Rect[T Number] struct {
	X, Y Interval[T]
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

func (r Rect[T]) ContainsPoint(x, y T) bool {
	return r.X.Contains(x) && r.Y.Contains(y)
}

func (r Rect[T]) String() string {
	return fmt.Sprintf("[Lo(%v, %v), Hi(%v, %v)]", r.X.Lo, r.Y.Lo, r.X.Hi, r.Y.Hi)
}
