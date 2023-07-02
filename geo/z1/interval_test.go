package z1_test

import (
	"fmt"
	"testing"

	"github.com/jensgreen/dux/geo/r1"
	"github.com/jensgreen/dux/geo/z1"
	"github.com/stretchr/testify/assert"
)

func Test_Interval_ContainsHalfClosed(t *testing.T) {
	tests := []struct {
		name     string
		interval z1.Interval
		x        int
		want     bool
	}{
		{"contain lower bound", z1.Interval{Lo: 0, Hi: 1}, 0, true},

		{"contains inner", z1.Interval{Lo: -1, Hi: 1}, 0, true},

		{"zero-length contains nothing", z1.Interval{Lo: 0, Hi: 0}, 0, false},

		{"does not contain upper bound", z1.Interval{Lo: 0, Hi: 1}, 1, false},

		{"does not contain higher or lower", z1.Interval{Lo: 0, Hi: 1}, 2, false},
		{"does not contain higher or lower", z1.Interval{Lo: 0, Hi: 1}, -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.interval.ContainsHalfClosed(tt.x)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Interval_Length(t *testing.T) {
	tests := []struct {
		interval z1.Interval
		want     int
	}{
		{z1.Interval{Lo: 0, Hi: 0}, 0},
		{z1.Interval{Lo: 0, Hi: 1}, 1},
		{z1.Interval{Lo: 0, Hi: 2}, 2},
		{z1.Interval{Lo: 0, Hi: -1}, -1},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v is %v", tt.interval, tt.want), func(t *testing.T) {
			got := tt.interval.Length()
			if got != tt.want {
				t.Errorf("\ngot  %v,\nwant %v", got, tt.want)
			}
		})
	}
}

func Test_SnapRoundInterval(t *testing.T) {
	tests := []struct {
		r1Interval r1.Interval
		want       z1.Interval
	}{
		{r1.Interval{Lo: 0.0, Hi: 1.99}, z1.Interval{Lo: 0, Hi: 1}},
		{r1.Interval{Lo: 0.01, Hi: 1.99}, z1.Interval{Lo: 0, Hi: 1}},
		{r1.Interval{Lo: 0.0, Hi: 2.5}, z1.Interval{Lo: 0, Hi: 2}},
		{r1.Interval{Lo: 2.5, Hi: 3.0}, z1.Interval{Lo: 2, Hi: 3}},
		{r1.Interval{Lo: 0.0, Hi: 0.99}, z1.Interval{Lo: 0, Hi: 0}},
		{r1.Interval{Lo: 0.99, Hi: 1.99}, z1.Interval{Lo: 0, Hi: 1}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v -> %v", tt.r1Interval, tt.want), func(t *testing.T) {
			got := z1.SnapRoundInterval(tt.r1Interval)
			assert.Equal(t, tt.want, got)
		})
	}
}
