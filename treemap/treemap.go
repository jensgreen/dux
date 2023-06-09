package treemap

import (
	"fmt"
	"strings"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo"
	"github.com/jensgreen/dux/r2"
	"github.com/jensgreen/dux/treemap/tiling"
)

type GenericTreemap[T any] struct {
	File     files.File
	Rect     T
	Parent   *GenericTreemap[T]
	Children []*GenericTreemap[T]
	Spillage T
}

func (tm *GenericTreemap[T]) Path() string {
	return tm.File.Path
}

func (tm *GenericTreemap[T]) FindNode(path string) (*GenericTreemap[T], error) {
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


type Treemap struct {
	File     files.File
	Rect     r2.Rect
	Parent   *Treemap
	Children []*Treemap
	Spillage r2.Rect
}

func (tm *Treemap) Path() string {
	return tm.File.Path
}

func (tm *Treemap) FindNode(path string) (*Treemap, error) {
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

func New(root files.FileTree, rect r2.Rect, tiler tiling.Tiler, maxDepth int) *Treemap {
	return newWithParent(nil, root, rect, tiler, maxDepth, 1)
}

func newWithParent(parent *Treemap, tree files.FileTree, rect r2.Rect, tiler tiling.Tiler, maxDepth int, depth int) *Treemap {
	if len(tree.Children) == 0 {
		return &Treemap{Parent: parent, File: tree.File, Rect: rect, Children: []*Treemap{}}
	}

	var childTreemaps []*Treemap
	// items that are too small to display are represented by the spillage box
	// in the bottom right corner
	spillage := r2.Rect{
		X: geo.IntervalFromPoint(rect.X.Hi),
		Y: geo.IntervalFromPoint(rect.Y.Hi),
	}

	treemap := &Treemap{
		Parent:   parent,
		File:     tree.File,
		Rect:     rect,
		Spillage: spillage,
	}

	if maxDepth == 0 || depth < maxDepth {
		var tiles []tiling.Tile
		tiles, spillage = tiler.Tile(rect, tree, depth)
		childTreemaps = make([]*Treemap, len(tiles))
		for i, tile := range tiles {
			childTreemaps[i] = newWithParent(treemap, tile.File, tile.Rect, tiler, maxDepth, depth+1)
		}

		treemap.Children = childTreemaps
		treemap.Spillage = spillage
	}

	return treemap
}
