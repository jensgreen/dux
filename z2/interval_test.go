package z2

import (
	"fmt"
	"testing"

	"github.com/jensgreen/dux/r1"
)

func Test_Interval_Contains(t *testing.T) {
	tests := []struct {
		interval Interval
		x        int
		want     bool
	}{
		// contain lower bound
		{Interval{0, 1}, 0, true},
		// contains inner
		{Interval{-1, 1}, 0, true},

		// zero-length contains nothing
		{Interval{0, 0}, 0, false},

		// does not contain upper bound
		{Interval{0, 1}, 1, false},
		// does not contain higher or lower
		{Interval{0, 1}, 2, false},
		{Interval{0, 1}, -1, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v is %v", tt.interval, tt.want), func(t *testing.T) {
			got := tt.interval.Contains(tt.x)
			if got != tt.want {
				t.Errorf("\ngot  %v,\nwant %v", got, tt.want)
			}
		})
	}
}

func Test_Interval_Length(t *testing.T) {
	tests := []struct {
		interval Interval
		want     int
	}{
		{Interval{0, 0}, 0},
		{Interval{0, 1}, 1},
		{Interval{0, 2}, 2},
		{Interval{0, -1}, -1},
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

func TestSnapRoundInterval_1(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 0.0, Hi: 1.99})
	expected := Interval{0, 1}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestSnapRoundInterval_2(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 0.01, Hi: 1.99})
	expected := Interval{0, 1}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestSnapRoundInterval_3(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 0.0, Hi: 2.5})
	expected := Interval{0, 2}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}
func TestSnapRoundInterval_4(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 2.5, Hi: 3.0})
	expected := Interval{2, 3}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestSnapRoundInterval_RoundDownWhenWithinAnInt(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 0.0, Hi: 0.99})
	expected := Interval{0, 0}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}

func TestSnapRoundInterval_RoundUpWhenCrossingInt(t *testing.T) {
	got := snapRoundInterval(r1.Interval{Lo: 0.99, Hi: 1.99})
	expected := Interval{0, 1}
	if !expected.Eq(got) {
		t.Errorf("Expected %+v, got %+v", expected, got)
	}
}
