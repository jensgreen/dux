package dux

import (
	"log"
	"sync"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
)

type Command interface {
	Execute(State) State
}

type Quit struct{}

func (Quit) Execute(state State) State {
	state.Quit = true
	return state
}

type Refresh struct{}

func (Refresh) Execute(state State) State {
	state.Refresh = &sync.Once{}
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
	Width  int
	Height int
}

func (cmd Resize) Execute(state State) State {
	log.Printf("Treemap area resized to (%d, %d)", cmd.Width, cmd.Height)
	state.TreemapRect = cmd.calcAreas(cmd.Width, cmd.Height)
	return state
}

func (cmd Resize) calcAreas(screenWidth int, screenHeight int) r2.Rect {
	return  r2.Rect{
		X: r1.Interval{Lo: 0, Hi: float64(screenWidth)},
		Y: r1.Interval{Lo: 0, Hi: float64(screenHeight)},
	}
}
