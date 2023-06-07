package dux

import (
	"log"
	"sync"

	"github.com/jensgreen/dux/nav"
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
	AppSize     z2.Point
	TreemapSize z2.Point
}

func (cmd Resize) Execute(state State) State {
	log.Printf("Resized app to (%d, %d)", cmd.AppSize.X, cmd.AppSize.Y)
	state.AppSize = cmd.AppSize
	state.TreemapSize = cmd.TreemapSize
	return state
}

type Select struct {
	Path string
}

func (cmd Select) Execute(state State) State {
	if state.Treemap != nil {
		selection, err := state.Treemap.FindNode(cmd.Path)
		if err != nil {
			log.Printf("Could not select %s", cmd.Path)
		} else {
			state.Selection = selection
		}
	}
	return state
}

type Deselect struct{}

func (cmd Deselect) Execute(state State) State {
	state.Selection = nil
	return state
}

type Navigate struct {
	Direction nav.Direction
}

func (cmd Navigate) Execute(state State) State {
	root := state.Treemap
	if state.Selection == nil {
		if cmd.Direction == nav.DirectionOut {
			return state
		}
		state.Selection = root
		return state
	}
	if state.Treemap != nil {
		if cmd.Direction == nav.DirectionOut && state.Selection.Path() == root.Path() {
			state.Selection = nil
			return state
		}
		state.Selection = nav.Navigate(state.Selection, cmd.Direction)
		return state
	}
	return state
}

type TogglePause struct{}

func (cmd TogglePause) Execute(state State) State {
	state.Pause = !state.Pause
	return state
}

type ZoomIn struct{}

// TODO implement a zoom stack
func (cmd ZoomIn) Execute(state State) State {
	state.Zoom = state.Selection
	return state
}

type ZoomOut struct{}

func (cmd ZoomOut) Execute(state State) State {
	if state.Zoom != nil {
		state.Selection = state.Zoom
		state.Zoom = state.Zoom.Parent
	}
	return state
}
