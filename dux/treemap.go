package dux

import (
	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
)

type Treemap struct {
	File     files.File
	Rect     r2.Rect
	Children []Treemap
	Spillage r2.Rect
}

func TreemapWithTiler(tree files.FileTree, rect r2.Rect, tiler Tiler, maxDepth int, depth int) Treemap {
	if len(tree.Children) == 0 {
		return Treemap{File: tree.File, Rect: rect, Children: []Treemap{}}
	}

	var childTreemaps []Treemap
	// items that are too small to display are represented by the spillage box
	// in the bottom right corner
	spillage := r2.Rect{
		X: r1.IntervalFromPoint(rect.X.Hi),
		Y: r1.IntervalFromPoint(rect.Y.Hi),
	}
	if maxDepth == -1 || depth < maxDepth {
		var tiles []Tile
		tiles, spillage = tiler.Tile(rect, tree, depth)
		childTreemaps = make([]Treemap, len(tiles))
		for i, tile := range tiles {
			childTreemaps[i] = TreemapWithTiler(tile.File, tile.Rect, tiler, maxDepth, depth+1)
		}
	}

	return Treemap{File: tree.File, Rect: rect, Children: childTreemaps, Spillage: spillage}
}
