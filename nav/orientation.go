package nav

import (
	"log"
	"math"

	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
)

// orientaiton that children in a treemap stretches out in
type orientation int

const (
	// orientationNone indicates that orientation is not meaningful because
	// there are too few children
	orientationNone orientation = iota
	// orientationHorizontal indicates a layout along the horizontal axis (left-to-right)
	orientationHorizontal
	// orientationVertical indicates a  layout along the vertical axis (top-to-bottom)
	orientationVertical
)

// TODO just store the split orientation in the treemap?
func orientationOf[T treemap.Rect](tm *treemap.Treemap[T]) orientation {
	if len(tm.Children) < 2 {
		return orientationNone
	}
	a, b := tm.Children[0], tm.Children[1]

	// hack to get around limitations of type unions in Go 1.18
	// https://github.com/golang/go/issues/51183#issuecomment-1049181719
	// https://stackoverflow.com/a/71378366
	var diff float64
	switch aa := any(a.Rect).(type) {
	case z2.Rect:
		bb := any(b.Rect).(z2.Rect)
		diff = float64(bb.X.Lo - aa.X.Lo)
	case r2.Rect:
		bb := any(b.Rect).(r2.Rect)
		diff = bb.X.Lo - aa.X.Lo
	default:
		log.Panicf("not implemented: %T", aa)
	}

	isVertical := math.Abs(diff) < 1e-9
	if isVertical {
		return orientationVertical
	}
	return orientationHorizontal
}
