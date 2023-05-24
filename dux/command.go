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
		state.Selection = FindSubTreemap(state.Treemap, cmd.Path)
	}
	return state
}

type Deselect struct{}

func (cmd Deselect) Execute(state State) State {
	state.Selection = nil
	return state
}

type Direction int

const (
	DirectionLeft Direction = iota
	DirectionRight
	DirectionUp
	DirectionDown

	DirectionIn
	DirectionOut
)

type Navigate struct {
	Direction Direction
}

func (cmd Navigate) Execute(state State) State {
	if state.Selection == nil {
		// TODO
		return state
	}
	if state.Treemap != nil {
		state.Selection = cmd.navigate(state.Selection, cmd.Direction)
	}
	return state
}

func (cmd Navigate) navigate(selection *Treemap, direction Direction) *Treemap {
	parent := selection.Parent
	switch direction {
	case DirectionLeft:
		for i := len(parent.Children) - 1; i >= 0; i-- {
			if parent.Children[i] == selection {
				if i > 0 {
					return parent.Children[i-1]
				}
				return selection
			}
		}
		return selection
	case DirectionRight:
		for i := range parent.Children {
			if parent.Children[i] == selection {
				if i < len(parent.Children)-1 {
					return parent.Children[i+1]
				}
				return selection
			}
		}
		return selection
	case DirectionUp:
	case DirectionDown:
	case DirectionIn:
	case DirectionOut:
		break // TODO
	}

	return nil
}
