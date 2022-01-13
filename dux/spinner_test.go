package dux

import (
	"testing"
	"time"
)

func Test_TickAdvancesFrame(t *testing.T) {
	spinner := newSpinner()
	want := spinner.frame + 1
	spinner.Tick()
	got := spinner.frame
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func Test_Throttle(t *testing.T) {
	spinner := spinner{
		lastUpdate: time.Now().Add(24 * time.Hour),
	}
	want := spinner.frame
	spinner.Tick()
	got := spinner.frame
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}
