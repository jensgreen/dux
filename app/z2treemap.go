package app

import (
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/z2"
)

// treemap with discrete z2-coordinates
type z2treemap struct {
	File     files.File
	Rect     z2.Rect
	Children []z2treemap
}

func (zt *z2treemap) Path() string {
	return zt.File.Path
}

func newZ2Treemap(tm *treemap.Treemap) z2treemap {
	children := make([]z2treemap, len(tm.Children))
	for i, child := range tm.Children {
		children[i] = newZ2Treemap(child)
	}

	return z2treemap{
		File:     tm.File,
		Rect:     z2.SnapRoundRect(tm.Rect),
		Children: children,
	}
}
