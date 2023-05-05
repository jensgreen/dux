package dux

import (
	"log"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/z2"
)

type Command interface {
	Execute(State) State
}

type Quit struct{}

func (Quit) Execute(state State) State {
	state.Quit = true
	return state
}

type IncreaseMaxDepth struct{}

func (IncreaseMaxDepth) Execute(state State) State {
	state.MaxDepth++
	return state
}

type DecreaseMaxDepth struct{}

func (DecreaseMaxDepth) Execute(state State) State {
	if state.MaxDepth > 0 {
		state.MaxDepth--
	}
	return state
}

type Resize struct {
	WindowWidth  int
	WindowHeight int
}

func (cmd Resize) Execute(state State) State {
	log.Printf("Window resized to (%d, %d)", cmd.WindowWidth, cmd.WindowHeight)
	_, _, treemapR2 := cmd.calcAreas(cmd.WindowWidth, cmd.WindowHeight)
	state.TreemapRect = treemapR2
	return state
}

// calcAreas sets up the different coordinate spaces basen on terminal windows size
func (cmd Resize) calcAreas(screenWidth int, screenHeight int) (screenSpaceZ2 z2.Rect, statusbarSpaceZ2 z2.Rect, treemapSpaceR2 r2.Rect) {

	// Screen coordinate space
	// z2, lo at top left, hi limited by screen size (half-open intervals)
	screenSpaceZ2 = z2.Rect{
		X: z2.Interval{Lo: 0, Hi: screenWidth + 1},
		Y: z2.Interval{Lo: 0, Hi: screenHeight + 1},
	}

	// Status bar coordinate space
	// z2, lo at top left, hi limited by (screen width, 1)
	statusbarSpaceZ2 = screenSpaceZ2
	statusbarSpaceZ2.Y.Hi = 1

	// treemap coordinate system
	// r2, lo at top left, hi limited by (screen width, screen height - (statusbar height))
	// origin in screen space coords is at (0, statusbar height) = (0, 1)
	TreemapSpaceZ2 := screenSpaceZ2
	TreemapSpaceZ2.Y.Hi -= statusbarSpaceZ2.Y.Length()
	treemapSpaceR2 = r2.Rect{
		X: r1.Interval{Lo: float64(TreemapSpaceZ2.X.Lo), Hi: float64(TreemapSpaceZ2.X.Hi - 1)},
		Y: r1.Interval{Lo: float64(TreemapSpaceZ2.Y.Lo), Hi: float64(TreemapSpaceZ2.Y.Hi - 1)},
	}

	return screenSpaceZ2, statusbarSpaceZ2, treemapSpaceR2
}
