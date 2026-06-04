package tiling

import (
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
)

// Tiles that are too small to be meaningfully represented individually are put in a
// special "Spillage" bucket and hidden from normal display.
const (
	MINIMUM_HEIGHT float64 = 3.0
	MINIMUM_WIDTH  float64 = 7.0
)

// Tiler arranges rectangular area into smaller rects with adjoining edges. The
// number of output tiles must match len(weights), and the area of each rect
// should depend on its relative weight.
type Tiler interface {
	Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect)
}

type Tile struct {
	File files.FileTree
	Rect r2.Rect
}

type Padding struct {
	Top, Right, Bottom, Left float64
}

func (p Padding) pad(rect r2.Rect) r2.Rect {
	rect.Y.Lo += p.Top
	rect.X.Hi -= p.Right
	rect.Y.Hi -= p.Bottom
	rect.X.Lo += p.Left

	if rect.X.IsEmpty() || rect.Y.IsEmpty() {
		return r2.Rect{}
	}
	return rect
}

type paddingTiler struct {
	tiler   Tiler
	padding Padding
}

func (p paddingTiler) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	return p.tiler.Tile(p.padding.pad(rect), fileTree, depth)
}

func WithPadding(tiler Tiler, widths Padding) Tiler {
	return paddingTiler{
		tiler:   tiler,
		padding: widths,
	}
}
