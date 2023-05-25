package nav

import (
	"math"

	"github.com/jensgreen/dux/treemap"
)
type orientation int

const (
	// orientationNone indicates that orientation is not meaningful because
	// there are too few children
	orientationNone orientation = iota
	// orientationHorizontal indicates a left-to-right layout
	orientationHorizontal
	// orientationVertical indicates a top-to-bottom layout
	orientationVertical
)

func orientationOf(treemap *treemap.Treemap) orientation {
	if len(treemap.Children) < 2 {
		return orientationNone
	}
	a, b := treemap.Children[0], treemap.Children[1]
	isVertical := math.Abs(b.Rect.X.Lo-a.Rect.X.Lo) < 1e-9
	if isVertical {
		return orientationVertical
	}
	return orientationHorizontal
}
