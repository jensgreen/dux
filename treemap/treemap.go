package treemap

import (
	"fmt"
	"strings"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap/tiling"
)

type Rect interface {
	geo.Rect[int] | geo.Rect[float64]
}

type Treemap[T Rect] struct {
	File     files.File
	Rect     T
	Parent   *Treemap[T]
	Children []*Treemap[T]
	Spillage T
}

type R2Treemap = Treemap[geo.Rect[float64]]
type Z2Treemap = Treemap[geo.Rect[int]]

func (tm *Treemap[T]) Path() string {
	return tm.File.Path
}

func (tm *Treemap[T]) FindNode(path string) (*Treemap[T], error) {
	if path == tm.Path() {
		return tm, nil
	}

	for _, c := range tm.Children {
		if strings.HasPrefix(path, c.Path()) {
			return c.FindNode(path)
		}
	}

	return nil, fmt.Errorf("no such node: %s", path)
}

func NewR2Treemap(root files.FileTree, rect r2.Rect, tiler tiling.Tiler, maxDepth int) *R2Treemap {
	return newR2Treemap(nil, root, rect, tiler, maxDepth, 1)
}

func newR2Treemap(parent *R2Treemap, tree files.FileTree, rect r2.Rect, tiler tiling.Tiler, maxDepth int, depth int) *R2Treemap {
	if len(tree.Children) == 0 {
		return &R2Treemap{Parent: parent, File: tree.File, Rect: rect, Children: []*R2Treemap{}}
	}

	var childTreemaps []*R2Treemap
	// items that are too small to display are represented by the spillage box
	// in the bottom right corner
	spillage := r2.Rect{
		X: geo.IntervalFromPoint(rect.X.Hi),
		Y: geo.IntervalFromPoint(rect.Y.Hi),
	}

	treemap := &R2Treemap{
		Parent:   parent,
		File:     tree.File,
		Rect:     rect,
		Spillage: spillage,
	}

	if maxDepth == 0 || depth < maxDepth {
		var tiles []tiling.Tile
		tiles, spillage = tiler.Tile(rect, tree, depth)
		childTreemaps = make([]*R2Treemap, len(tiles))
		for i, tile := range tiles {
			childTreemaps[i] = newR2Treemap(treemap, tile.File, tile.Rect, tiler, maxDepth, depth+1)
		}

		treemap.Children = childTreemaps
		treemap.Spillage = spillage
	}

	return treemap
}

func NewZ2Treemap(tm *R2Treemap) *Z2Treemap {
	children := make([]*Z2Treemap, len(tm.Children))
	for i, child := range tm.Children {
		children[i] = NewZ2Treemap(child)
	}

	return &Z2Treemap{
		File:     tm.File,
		Rect:     z2.SnapRoundRect(tm.Rect),
		Children: children,
	}
}
