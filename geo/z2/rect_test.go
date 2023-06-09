package z2

import (
	"testing"

	"github.com/jensgreen/dux/geo/r1"
	"github.com/jensgreen/dux/geo/r2"
)

func TestSnapRoundRect_1(t *testing.T) {
	x := r1.Interval{Lo: 0.0, Hi: 0.99}
	y := r1.Interval{Lo: 1.01, Hi: 2.66}
	got := SnapRoundRect(r2.Rect{X: x, Y: y})

	expected := Rect{
		X: Interval{0, 0},
		Y: Interval{1, 2},
	}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestRect_ContainsCenter(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := rect.ContainsPoint(1, 1)
	if !ok {
		t.Error()
	}
}

func TestRect_ContainsLoX(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := rect.ContainsPoint(0, 0) && rect.ContainsPoint(0, 1)
	if !ok {
		t.Error()
	}
}
func TestRect_ContainsLoY(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := rect.ContainsPoint(0, 0) && rect.ContainsPoint(1, 0)
	if !ok {
		t.Error()
	}
}

func TestRect_NotContainsHiX(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := !(rect.ContainsPoint(2, 0) && rect.ContainsPoint(2, 1) && rect.ContainsPoint(2, 2))
	if !ok {
		t.Error()
	}
}
func TestRect_NotContainsHiY(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := !(rect.ContainsPoint(0, 2) && rect.ContainsPoint(1, 2) && rect.ContainsPoint(2, 2))
	if !ok {
		t.Error()
	}
}

func TestRect_NotContainsLower(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := !(rect.ContainsPoint(-1, 0) && rect.ContainsPoint(0, -1) && rect.ContainsPoint(-1, -1))
	if !ok {
		t.Error()
	}
}

func TestRect_NotContainsHigher(t *testing.T) {
	rect := Rect{
		X: Interval{0, 2},
		Y: Interval{0, 2},
	}
	ok := !(rect.ContainsPoint(3, 0) && rect.ContainsPoint(0, 3) && rect.ContainsPoint(3, 3))
	if !ok {
		t.Error()
	}
}
