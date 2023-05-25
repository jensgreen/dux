package dux

import (
	"log"
	"sync"

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
		state.Selection = state.Treemap.FindSubTreemap(cmd.Path)
	}
	return state
}

type Deselect struct{}

func (cmd Deselect) Execute(state State) State {
	state.Selection = nil
	return state
}

type Navigate struct {
	Direction Direction
}

func (cmd Navigate) Execute(state State) State {
	if state.Selection == nil {
		// TODO
		return state
	}
	if state.Treemap != nil {
		nav := NewNavigator()
		state.Selection = nav.Navigate(state.Selection, cmd.Direction)
	}
	return state
}
