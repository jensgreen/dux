package tiling

import (
	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
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

type VerticalSplit struct{}

func (VerticalSplit) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	tiles = []Tile{}

	totalWeight := float64(fileTree.File.Size)
	nextMinX := rect.Lo().X
	for _, f := range fileTree.Children {
		weightFactor := float64(f.File.Size) / totalWeight
		size := rect.Size()
		dx := weightFactor * float64(size.X)
		candidate := r2.Rect{
			X: r1.Interval{Lo: nextMinX, Hi: nextMinX + dx},
			Y: rect.Y,
		}

		if dx < MINIMUM_WIDTH {
			// Don't show this tile, it's too small.
			// Grow spillage from the right.
			spillage.X.Lo -= candidate.X.Length()
		} else {
			tiles = append(tiles, Tile{File: f, Rect: candidate})
			nextMinX = candidate.X.Hi
		}
	}
	return tiles, spillage
}

type HorizontalSplit struct{}

func (HorizontalSplit) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	tiles = []Tile{}

	totalWeight := float64(fileTree.File.Size)
	nextMinY := rect.Lo().Y
	for _, f := range fileTree.Children {
		weightFactor := float64(f.File.Size) / totalWeight
		size := rect.Size()
		dy := weightFactor * float64(size.Y)
		candidate := r2.Rect{
			X: rect.X,
			Y: r1.Interval{Lo: nextMinY, Hi: nextMinY + dy},
		}
		if dy < MINIMUM_HEIGHT {
			// Don't show this tile, it's too small.
			// grow spillage from the bottom.
			spillage.Y.Lo -= candidate.Y.Length()
		} else {
			tiles = append(tiles, Tile{File: f, Rect: candidate})
			nextMinY = candidate.Y.Hi
		}
	}
	return tiles, spillage
}

// SliceAndDice alternates between HorizontalSplit and VerticalSplit based on depth
type SliceAndDice struct{}

func (sd SliceAndDice) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	var tiler Tiler
	if depth%2 == 0 {
		tiler = HorizontalSplit{}
	} else {
		tiler = VerticalSplit{}
	}
	return tiler.Tile(rect, fileTree, depth)
}

type Padding struct {
	tiler   Tiler
	padding float64
}

func (p Padding) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	return p.tiler.Tile(p.pad(rect), fileTree, depth)
}

func (p Padding) pad(rect r2.Rect) r2.Rect {
	// Returns EmptyRect() if resulting rect is of size <= 0. The empty rect/interval's
	// position is unexpected:
	// "The empty interval is considered to be positioned arbitrarily on the real line"
	return rect.ExpandedByMargin(-p.padding)
}

func WithPadding(tiler Tiler, padding float64) Tiler {
	return Padding{
		tiler:   tiler,
		padding: padding,
	}
}
