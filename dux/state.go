package dux

import (
	"sync"

	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
)

type State struct {
	Treemap        *treemap.Treemap
	Selection      *treemap.Treemap
	Zoom           *treemap.Treemap
	Quit           bool
	MaxDepth       int
	TreemapSize    z2.Point
	AppSize        z2.Point
	TotalFiles     int
	IsWalkingFiles bool
	Pause          bool
	Refresh        *sync.Once
}

type StateEvent struct {
	State  State
	Errors []error
}
