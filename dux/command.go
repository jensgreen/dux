package dux

import (
	"log"

	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/nav"
)

type Command interface {
	Execute(State) (State, Action)
}

func state(s State) (State, Action) {
	return s, ActionNone
}

type Quit struct{}

func (Quit) Execute(state State) (State, Action) {
	state.Quit = true
	return state, ActionNone
}

type Refresh struct{}

func (Refresh) Execute(state State) (State, Action) {
	return state, ActionNone
}

type SendToBackground struct{}

func (SendToBackground) Execute(state State) (State, Action) {
	return state, ActionBackground
}

type IncreaseMaxDepth struct{}

func (IncreaseMaxDepth) Execute(state State) (State, Action) {
	state.MaxDepth++
	return state, ActionNone
}

type DecreaseMaxDepth struct{}

func (DecreaseMaxDepth) Execute(state State) (State, Action) {
	if state.MaxDepth > 0 {
		state.MaxDepth--
	}
	return state, ActionNone
}

type Resize struct {
	AppSize     z2.Point
	TreemapSize z2.Point
}

func (cmd Resize) Execute(state State) (State, Action) {
	log.Printf("Resized app to (%d, %d)", cmd.AppSize.X, cmd.AppSize.Y)
	state.AppSize = cmd.AppSize
	state.TreemapSize = cmd.TreemapSize
	return state, ActionNone
}

type Select struct {
	Path string
}

func (cmd Select) Execute(state State) (State, Action) {
	if state.Treemap != nil {
		selection, err := state.Treemap.FindNode(cmd.Path)
		if err != nil {
			log.Printf("Could not select %s", cmd.Path)
		} else {
			state.Selection = selection
		}
	}
	return state, ActionNone
}

type Deselect struct{}

func (cmd Deselect) Execute(state State) (State, Action) {
	state.Selection = nil
	return state, ActionNone
}

type Navigate struct {
	Direction nav.Direction
}

func (cmd Navigate) Execute(state State) (State, Action) {
	root := state.Treemap
	if state.Selection == nil {
		if cmd.Direction == nav.DirectionOut {
			return state, ActionNone
		}
		state.Selection = root
		return state, ActionNone
	}
	if state.Treemap != nil {
		if cmd.Direction == nav.DirectionOut && state.Selection.Path() == root.Path() {
			state.Selection = nil
			return state, ActionNone
		}
		state.Selection = nav.Navigate(state.Selection, cmd.Direction)
		return state, ActionNone
	}
	return state, ActionNone
}

type TogglePause struct{}

func (cmd TogglePause) Execute(state State) (State, Action) {
	state.Pause = !state.Pause
	return state, ActionNone
}

type ZoomIn struct{}

// TODO implement a zoom stack
func (cmd ZoomIn) Execute(state State) (State, Action) {
	state.Zoom = state.Selection
	return state, ActionNone
}

type ZoomOut struct{}

func (cmd ZoomOut) Execute(state State) (State, Action) {
	if state.Zoom != nil {
		state.Selection = state.Zoom
		state.Zoom = state.Zoom.Parent
	}
	return state, ActionNone
}
