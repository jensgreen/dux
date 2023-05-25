package dux

import (
	"fmt"
	"math"

	"github.com/jensgreen/dux/treemap"
)

type Orientation int

const (
	// OrientationNone indicates that orientation is not meaningful because
	// there are too few children
	OrientationNone Orientation = iota
	// OrientationHorizontal indicates a left-to-right layout
	OrientationHorizontal
	// OrientationVertical indicates a top-to-bottom layout
	OrientationVertical
)

func orientation(treemap *treemap.Treemap) Orientation {
	if len(treemap.Children) < 2 {
		return OrientationNone
	}
	a, b := treemap.Children[0], treemap.Children[1]
	isVertical := math.Abs(b.Rect.X.Lo-a.Rect.X.Lo) < 1e-9
	if isVertical {
		return OrientationVertical
	}
	return OrientationHorizontal
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

type Navigator interface {
	Navigate(treemap *treemap.Treemap, direction Direction) *treemap.Treemap
}

type navigator struct {
}

func (n *navigator) Navigate(tm *treemap.Treemap, direction Direction) *treemap.Treemap {
	switch direction {
	case DirectionLeft:
		sibling := n.westNeighbor(tm)
		if sibling != nil {
			return sibling
		}
	case DirectionRight:
		neighbor := n.eastNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionUp:
		neighbor := n.northNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionDown:
		neighbor := n.southNeighbor(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionIn:
		neighbor := n.stepIn(tm)
		if neighbor != nil {
			return neighbor
		}
	case DirectionOut:
		neighbor := n.stepOut(tm)
		if neighbor != nil {
			return neighbor
		}
	}

	return tm
}

func (n *navigator) eastNeighbor(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientation(tm)
	switch or {
	case OrientationHorizontal:
		sibling := n.nextSibling(tm, OrientationHorizontal)
		if sibling != nil {
			return sibling
		}
		return n.eastNeighbor(parent)
	case OrientationVertical:
		return n.nextSibling(tm, OrientationHorizontal)
	case OrientationNone:
		switch orientation(parent) {
		case OrientationHorizontal:
			return n.nextSibling(tm, OrientationHorizontal)
		case OrientationVertical:
			return n.eastNeighbor(parent)
		default:
			return n.eastNeighbor(parent)
		}
	}
	return nil
}

func (n *navigator) westNeighbor(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientation(tm)
	switch or {
	case OrientationHorizontal:
		sibling := n.prevSibling(tm, OrientationHorizontal)
		if sibling != nil {
			return sibling
		}
		return n.westNeighbor(parent)
	case OrientationVertical:
		return n.prevSibling(tm, OrientationHorizontal)
	case OrientationNone:
		switch orientation(parent) {
		case OrientationHorizontal:
			return n.prevSibling(tm, OrientationHorizontal)
		case OrientationVertical:
			return n.westNeighbor(parent)
		default:
			return n.westNeighbor(parent)
		}
	}
	return nil
}

func (n *navigator) northNeighbor(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientation(tm)
	switch or {
	case OrientationVertical:
		sibling := n.prevSibling(tm, OrientationVertical)
		if sibling != nil {
			return sibling
		}
		return n.northNeighbor(parent)
	case OrientationHorizontal:
		return n.prevSibling(tm, OrientationVertical)
	case OrientationNone:
		switch orientation(parent) {
		case OrientationVertical:
			return n.prevSibling(tm, OrientationVertical)
		case OrientationHorizontal:
			return n.northNeighbor(parent)
		default:
			return n.northNeighbor(parent)
		}
	}
	return nil
}

func (n *navigator) southNeighbor(tm *treemap.Treemap) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}

	or := orientation(tm)
	switch or {
	case OrientationVertical:
		sibling := n.nextSibling(tm, OrientationVertical)
		if sibling != nil {
			return sibling
		}
		return n.southNeighbor(parent)
	case OrientationHorizontal:
		return n.nextSibling(tm, OrientationVertical)
	case OrientationNone:
		switch orientation(parent) {
		case OrientationVertical:
			return n.nextSibling(tm, OrientationVertical)
		case OrientationHorizontal:
			return n.southNeighbor(parent)
		default:
			return n.southNeighbor(parent)
		}
	}
	return nil
}

func (n *navigator) stepIn(tm *treemap.Treemap) *treemap.Treemap {
	if (len(tm.Children) > 0) {
		return tm.Children[0]
	}
	return nil
}

func (n *navigator) stepOut(tm *treemap.Treemap) *treemap.Treemap {
	if (tm.Parent != nil) {
		return tm.Parent
	}
	return nil
}

func (n *navigator) nextSibling(tm *treemap.Treemap, or Orientation) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}
	if orientation(parent) != or {
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

func (n *navigator) prevSibling(tm *treemap.Treemap, or Orientation) *treemap.Treemap {
	parent := tm.Parent
	if parent == nil {
		return nil
	}
	if orientation(parent) != or {
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

func NewNavigator() Navigator {
	return &navigator{}
}
