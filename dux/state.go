package dux

import (
	"sync"

	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/z2"
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
