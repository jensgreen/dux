package nav

import (
	"fmt"

	"github.com/jensgreen/dux/treemap"
)

func Navigate(tm *treemap.Treemap, direction Direction) *treemap.Treemap {
	var destination *treemap.Treemap

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

func stepLeft(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientationOf(tm)
	switch or {
	case orientationHorizontal:
		sibling := prevSibling(tm, orientationHorizontal)
		if sibling != nil {
			return sibling
		}
		return stepLeft(parent)
	case orientationVertical:
		return prevSibling(tm, orientationHorizontal)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationHorizontal:
			return prevSibling(tm, orientationHorizontal)
		case orientationVertical:
			return stepLeft(parent)
		default:
			return stepLeft(parent)
		}
	}
	return nil
}

func stepRight(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientationOf(tm)
	switch or {
	case orientationHorizontal:
		sibling := nextSibling(tm, orientationHorizontal)
		if sibling != nil {
			return sibling
		}
		return stepRight(parent)
	case orientationVertical:
		return nextSibling(tm, orientationHorizontal)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationHorizontal:
			return nextSibling(tm, orientationHorizontal)
		case orientationVertical:
			return stepRight(parent)
		default:
			return stepRight(parent)
		}
	}
	return nil
}


func stepUp(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientationOf(tm)
	switch or {
	case orientationVertical:
		sibling := prevSibling(tm, orientationVertical)
		if sibling != nil {
			return sibling
		}
		return stepUp(parent)
	case orientationHorizontal:
		return prevSibling(tm, orientationVertical)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationVertical:
			return prevSibling(tm, orientationVertical)
		case orientationHorizontal:
			return stepUp(parent)
		default:
			return stepUp(parent)
		}
	}
	return nil
}

func stepDown(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientationOf(tm)
	switch or {
	case orientationVertical:
		sibling := nextSibling(tm, orientationVertical)
		if sibling != nil {
			return sibling
		}
		return stepDown(parent)
	case orientationHorizontal:
		return nextSibling(tm, orientationVertical)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationVertical:
			return nextSibling(tm, orientationVertical)
		case orientationHorizontal:
			return stepDown(parent)
		default:
			return stepDown(parent)
		}
	}
	return nil
}

func stepIn(tm *treemap.Treemap) *treemap.Treemap {
	if len(tm.Children) > 0 {
		return tm.Children[0]
	}
	return nil
}

func stepOut(tm *treemap.Treemap) *treemap.Treemap {
	if tm.Parent != nil {
		return tm.Parent
	}
	return nil
}

func nextSibling(tm *treemap.Treemap, or orientation) *treemap.Treemap {
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

func prevSibling(tm *treemap.Treemap, or orientation) *treemap.Treemap {
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
