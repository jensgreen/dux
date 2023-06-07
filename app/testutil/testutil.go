package testutil

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func InitSimScreen(t *testing.T, width int, height int) tcell.SimulationScreen {
	screen := tcell.NewSimulationScreen("")
	err := screen.Init()
	if err != nil {
		t.Fatal(err)
	}
	screen.SetSize(width, height)
	return screen
}

func ScreenToString(screen tcell.SimulationScreen) string {
	screen.Show()
	cells, width, _ := screen.GetContents()
	sb := strings.Builder{}
	for i, c := range cells {
		for _, r := range c.Runes {
			sb.WriteRune(r)
		}
		if (i+1)%width == 0 && i != len(cells)-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}
