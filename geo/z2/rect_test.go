package z2_test

import (
	"testing"

	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/geo/r1"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z1"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/stretchr/testify/assert"
)

func TestSnapRoundRect(t *testing.T) {
	x := r1.Interval{Lo: 0.0, Hi: 0.99}
	y := r1.Interval{Lo: 1.01, Hi: 2.66}
	got := z2.SnapRoundRect(r2.Rect{X: x, Y: y})

	want := z2.Rect{
		X: z1.Interval{Lo: 0, Hi: 0},
		Y: z1.Interval{Lo: 1, Hi: 2},
	}
	assert.True(t, want.Eq(got), "Expected %+v, got %+v", want, got)
}

func Test_Rect_ContainsHalfClosed(t *testing.T) {
	w, h := 2, 2
	tests := []struct {
		name string
		rect z2.Rect
		pt   z2.Point
		want bool
	}{
		{"contains lower bounds", geo.NewRect(0, 0, w, h), geo.NewPoint(0, 0), true},
		{"contains lower bounds", geo.NewRect(0, 0, w, h), geo.NewPoint(0, 1), true},
		{"contains lower bounds", geo.NewRect(0, 0, w, h), geo.NewPoint(1, 0), true},

		{"contains inner", geo.NewRect(0, 0, w, h), geo.NewPoint(1, 1), true},

		{"zero-size contains nothing", geo.NewRect(0, 0, 0, 0), geo.NewPoint(0, 0), false},

		{"does not contain upper bound", geo.NewRect(0, 0, w, h), geo.NewPoint(2, 2), false},
		{"does not contain upper bound", geo.NewRect(0, 0, w, h), geo.NewPoint(0, 2), false},
		{"does not contain upper bound", geo.NewRect(0, 0, w, h), geo.NewPoint(2, 0), false},
		{"does not contain upper bound", geo.NewRect(0, 0, w, h), geo.NewPoint(2, 1), false},

		{"does not contain higher", geo.NewRect(0, 0, w, h), geo.NewPoint(0, 3), false},
		{"does not contain lower", geo.NewRect(0, 0, w, h), geo.NewPoint(1, -1), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rect.ContainsHalfClosed(tt.pt)
			assert.Equal(t, tt.want, got)
		})
	}
}
