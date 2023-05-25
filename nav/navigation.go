package nav

import (
	"fmt"

	"github.com/jensgreen/dux/treemap"
)

func Navigate(tm *treemap.Treemap, direction Direction) *treemap.Treemap {
	switch direction {
	case DirectionLeft:
		sibling := westNeighbor(tm)
		if sibling != nil {
			return sibling
		}
	case DirectionRight:
		neighbor := eastNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionUp:
		neighbor := northNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionDown:
		neighbor := southNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionIn:
		neighbor := stepIn(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionOut:
		neighbor := stepOut(tm)
		if neighbor != nil {
			return neighbor
		}
	}

	return tm
}

func eastNeighbor(tm *treemap.Treemap) *treemap.Treemap {
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
		return eastNeighbor(parent)
	case orientationVertical:
		return nextSibling(tm, orientationHorizontal)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationHorizontal:
			return nextSibling(tm, orientationHorizontal)
		case orientationVertical:
			return eastNeighbor(parent)
		default:
			return eastNeighbor(parent)
		}
	}
	return nil
}

func westNeighbor(tm *treemap.Treemap) *treemap.Treemap {
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
		return westNeighbor(parent)
	case orientationVertical:
		return prevSibling(tm, orientationHorizontal)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationHorizontal:
			return prevSibling(tm, orientationHorizontal)
		case orientationVertical:
			return westNeighbor(parent)
		default:
			return westNeighbor(parent)
		}
	}
	return nil
}

func northNeighbor(tm *treemap.Treemap) *treemap.Treemap {
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
		return northNeighbor(parent)
	case orientationHorizontal:
		return prevSibling(tm, orientationVertical)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationVertical:
			return prevSibling(tm, orientationVertical)
		case orientationHorizontal:
			return northNeighbor(parent)
		default:
			return northNeighbor(parent)
		}
	}
	return nil
}

func southNeighbor(tm *treemap.Treemap) *treemap.Treemap {
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
		return southNeighbor(parent)
	case orientationHorizontal:
		return nextSibling(tm, orientationVertical)
	case orientationNone:
		switch orientationOf(parent) {
		case orientationVertical:
			return nextSibling(tm, orientationVertical)
		case orientationHorizontal:
			return southNeighbor(parent)
		default:
			return southNeighbor(parent)
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
