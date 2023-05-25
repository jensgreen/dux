package dux

import (
	"strings"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
)

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

func (tm *Treemap) FindSubTreemap(path string) *Treemap {
	if path == tm.Path() {
		return tm
	}

	for _, c := range tm.Children {
		if strings.HasPrefix(path, c.Path()) {
			return c.FindSubTreemap(path)
		}
	}

	return nil
}

func TreemapWithTiler(root files.FileTree, rect r2.Rect, tiler Tiler, maxDepth int, depth int) *Treemap {
	return treemapWithTiler(nil, root, rect, tiler, maxDepth, depth)
}

func treemapWithTiler(parent *Treemap, tree files.FileTree, rect r2.Rect, tiler Tiler, maxDepth int, depth int) *Treemap {
	if len(tree.Children) == 0 {
		return &Treemap{Parent: parent, File: tree.File, Rect: rect, Children: []*Treemap{}}
	}

	var childTreemaps []*Treemap
	// items that are too small to display are represented by the spillage box
	// in the bottom right corner
	spillage := r2.Rect{
		X: r1.IntervalFromPoint(rect.X.Hi),
		Y: r1.IntervalFromPoint(rect.Y.Hi),
	}

	treemap := &Treemap{
		Parent:   parent,
		File:     tree.File,
		Rect:     rect,
		Spillage: spillage,
	}

	if maxDepth == -1 || depth < maxDepth {
		var tiles []Tile
		tiles, spillage = tiler.Tile(rect, tree, depth)
		childTreemaps = make([]*Treemap, len(tiles))
		for i, tile := range tiles {
			childTreemaps[i] = treemapWithTiler(treemap, tile.File, tile.Rect, tiler, maxDepth, depth+1)
		}

		treemap.Children = childTreemaps
		treemap.Spillage = spillage
	}

	return treemap
}
