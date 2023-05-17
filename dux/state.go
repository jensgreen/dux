package dux

import (
	"sync"

	"github.com/jensgreen/dux/z2"
)

type State struct {
	Treemap        Treemap
	Quit           bool
	MaxDepth       int
	TreemapSize    z2.Point
	AppSize        z2.Point
	TotalFiles     int
	IsWalkingFiles bool
	Refresh        *sync.Once
	Selection      *string
}

type StateEvent struct {
	State  State
	Errors []error
}
