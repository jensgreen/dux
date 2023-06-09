package nav

import (
	"fmt"

	"github.com/jensgreen/dux/treemap"
)

// Navigate returns an adjacent Treemap in the given direction,
// stepping up the tree if necessary. Returns input Treemap if no
// valid destination exists (e.g. for the root).
func Navigate[T treemap.Rect](tm *treemap.Treemap[T], direction Direction) *treemap.Treemap[T] {
	var destination *treemap.Treemap[T]

	switch direction {
	case DirectionLeft:
		destination = stepLeft(tm)
	case DirectionRight:
		destination = stepRight(tm)
	case DirectionUp:
		destination = stepUp(tm)
	case DirectionDown:
		destination = stepDown(tm)
	case DirectionIn:
		destination = stepIn(tm)
	case DirectionOut:
		destination = stepOut(tm)
	}

	if destination == nil {
		return tm
	}
	return destination
}

func stepLeft[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	return stepHorizontal[T](tm, prevSibling[T])
}

func stepRight[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	return stepHorizontal[T](tm, nextSibling[T])
}

func stepUp[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	return stepVertical[T](tm, prevSibling[T])
}

func stepDown[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	return stepVertical[T](tm, nextSibling[T])
}

func stepHorizontal[T treemap.Rect](tm *treemap.Treemap[T], getSibling adjacentSibling[T]) *treemap.Treemap[T] {
	parent := tm.Parent
	isRoot := parent == nil
	if isRoot {
		return nil
	}

	switch orientationOf(tm) {
	case orientationHorizontal:
		sibling := getSibling(tm, orientationHorizontal)
		if sibling != nil {
			return sibling
		}
		return stepHorizontal(parent, getSibling)
	case orientationVertical:
		return getSibling(tm, orientationHorizontal)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationHorizontal:
			return getSibling(tm, orientationHorizontal)
		default:
			return stepHorizontal(parent, getSibling)
		}
	}
	return nil
}

func stepVertical[T treemap.Rect](tm *treemap.Treemap[T], getSibling adjacentSibling[T]) *treemap.Treemap[T] {
	parent := tm.Parent
	isRoot := parent == nil
	if isRoot {
		return nil
	}

	switch orientationOf(tm) {
	case orientationVertical:
		sibling := getSibling(tm, orientationVertical)
		if sibling != nil {
			return sibling
		}
		return stepVertical(parent, getSibling)
	case orientationHorizontal:
		return getSibling(tm, orientationVertical)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationVertical:
			return getSibling(tm, orientationVertical)
		default:
			return stepVertical(parent, getSibling)
		}
	}
	return nil
}

func stepIn[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	if len(tm.Children) > 0 {
		return tm.Children[0]
	}
	return nil
}

func stepOut[T treemap.Rect](tm *treemap.Treemap[T]) *treemap.Treemap[T] {
	if tm.Parent != nil {
		return tm.Parent
	}
	return nil
}

// get the next/previous sibling in the given orientation
type adjacentSibling[T treemap.Rect] func(*treemap.Treemap[T], orientation) *treemap.Treemap[T]

// implements adjacentSibling
func nextSibling[T treemap.Rect](tm *treemap.Treemap[T], or orientation) *treemap.Treemap[T] {
	parent := tm.Parent
	if parent == nil {
		return nil
	}
	if orientationOf(parent) != or {
		return nil
	}
	for i := range parent.Children {
		if parent.Children[i] == tm {
			if i < len(parent.Children)-1 {
				sibling := parent.Children[i+1]
				return sibling
			} else {
				return nil
			}
		}
	}
	panic(fmt.Sprintf("%s not a child of %s", tm.Path(), parent.Path()))
}

// implements adjacentSibling
func prevSibling[T treemap.Rect](tm *treemap.Treemap[T], or orientation) *treemap.Treemap[T] {
	parent := tm.Parent
	if parent == nil {
		return nil
	}
	if orientationOf(parent) != or {
		return nil
	}
	for i := len(parent.Children) - 1; i >= 0; i-- {
		if parent.Children[i] == tm {
			if i > 0 {
				sibling := parent.Children[i-1]
				return sibling
			} else {
				return nil
			}
		}
	}
	panic(fmt.Sprintf("%s not a child of %s", tm.Path(), parent.Path()))
}
