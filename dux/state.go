package dux

import (
	"sync"

	"github.com/golang/geo/r2"
)

type State struct {
	Treemap        Treemap
	Quit           bool
	MaxDepth       int
	TreemapRect    r2.Rect
	TotalFiles     int
	IsWalkingFiles bool
	Refresh        *sync.Once
}

type StateEvent struct {
	State  State
	Errors []error
}
