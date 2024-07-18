package dux

import (
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
)

type State struct {
	Treemap        *treemap.R2Treemap
	Selection      *treemap.R2Treemap
	Zoom           *treemap.R2Treemap
	Quit           bool
	MaxDepth       int
	TreemapSize    z2.Point
	AppSize        z2.Point
	TotalFiles     int
	IsWalkingFiles bool
	Pause          bool
}

type Action int

const (
	ActionNone = iota
	ActionRefresh
	ActionBackground
	ActionResize
)

type StateEvent struct {
	State  State
	Action Action
	Errors []error
}
