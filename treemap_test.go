package dux

import (
	"testing"

	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
)

func TestTreemapWithTiler_NoChildren(t *testing.T) {
	tree := files.FileTree{}
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	got := TreemapWithTiler(tree, rect, VerticalSplit{}, -1, 0)

	expected := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	if !expected.ApproxEqual(got.Rect) {
		t.Errorf("got %v, expected %v", got, expected)
	}
	if len(got.Children) != 0 {
		t.Errorf("expected no children, got %v", got.Children)
	}
}

func TestTreemapWithTiler_SplitsCorrectly(t *testing.T) {
	tree := files.FileTree{
		File: files.File{Size: 2},
		Children: []files.FileTree{
			{File: files.File{Size: 1}},
			{File: files.File{Size: 1}},
		},
	}
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	got := TreemapWithTiler(tree, rect, VerticalSplit{}, -1, 0)

	if len(got.Children) != 2 {
		t.Errorf("expected 2 children, got %v", len(got.Children))
	}
	if !got.Children[0].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 20, Y: 40})) {
		t.Errorf("got %v", got.Children[0].Rect)
	}
	if !got.Children[1].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 20, Y: 0}, r2.Point{X: 40, Y: 40})) {
		t.Errorf("got %v", got.Children[1].Rect)
	}
}

func TestVerticalSplit_SplitsTwoEqualWeightsInHalfVertically(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	fileTree := files.FileTree{
		File: files.File{Size: 2},
		Children: []files.FileTree{
			{File: files.File{Size: 1}},
			{File: files.File{Size: 1}},
		},
	}
	got, _ := VerticalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !got[0].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 20, Y: 40})) {
		t.Errorf("got %v", got[0])
	}
	if !got[1].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 20, Y: 0}, r2.Point{X: 40, Y: 40})) {
		t.Errorf("got %v", got[1])
	}
}

func TestHorizontalSplit_SplitsTwoEqualWeightsInHalfHorizontally(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	fileTree := files.FileTree{
		File: files.File{Size: 2},
		Children: []files.FileTree{
			{File: files.File{Size: 1}},
			{File: files.File{Size: 1}},
		},
	}
	got, _ := HorizontalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !got[0].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 20})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !got[1].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 0, Y: 20}, r2.Point{X: 40, Y: 40})) {
		t.Errorf("got %v", got[1].Rect)
	}
}

func TestHorizontalSplit_WorksWithNonZeroX(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 50, Y: 40})
	fileTree := files.FileTree{
		File: files.File{Size: 2},
		Children: []files.FileTree{
			{File: files.File{Size: 1}},
			{File: files.File{Size: 1}},
		},
	}
	got, _ := HorizontalSplit{}.Tile(rect, fileTree, 0)

	if !got[0].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 50, Y: 20})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !got[1].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 10, Y: 20}, r2.Point{X: 50, Y: 40})) {
		t.Errorf("got %v", got[1].Rect)
	}
}

func TestVerticalSplit_SplitsTwoNoRemainder(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 30, Y: 30})
	fileTree := files.FileTree{
		File: files.File{Size: 3},
		Children: []files.FileTree{
			{File: files.File{Size: 1}},
			{File: files.File{Size: 2}},
		},
	}
	got, _ := VerticalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !got[0].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 10, Y: 30})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !got[1].Rect.ApproxEqual(r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 30, Y: 30})) {
		t.Errorf("got %v", got[1].Rect)
	}
}

func TestTreemapWithTiler_CapsDepth(t *testing.T) {
	// TODO
}

func TestTreemapWithTiler_MaxDepthNegativeOneNoCap(t *testing.T) {
	// TODO
}
