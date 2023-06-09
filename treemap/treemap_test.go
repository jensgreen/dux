package treemap

import (
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap/tiling"
	"github.com/stretchr/testify/assert"
)

func TestTreemapWithTiler_NoChildren(t *testing.T) {
	tree := files.FileTree{}
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	got := New(tree, rect, tiling.VerticalSplit{}, 0)

	expected := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	if !r2.RectApproxEqual(expected, got.Rect) {
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
	got := New(tree, rect, tiling.VerticalSplit{}, 0)

	if len(got.Children) != 2 {
		t.Errorf("expected 2 children, got %v", len(got.Children))
	}
	if !r2.RectApproxEqual(got.Children[0].Rect, r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 20, Y: 40})) {
		t.Errorf("got %v", got.Children[0].Rect)
	}
	if !r2.RectApproxEqual(got.Children[1].Rect, r2.RectFromPoints(r2.Point{X: 20, Y: 0}, r2.Point{X: 40, Y: 40})) {
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
	got, _ := tiling.VerticalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !r2.RectApproxEqual(got[0].Rect, r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 20, Y: 40})) {
		t.Errorf("got %v", got[0])
	}
	if !r2.RectApproxEqual(got[1].Rect, r2.RectFromPoints(r2.Point{X: 20, Y: 0}, r2.Point{X: 40, Y: 40})) {
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
	got, _ := tiling.HorizontalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !r2.RectApproxEqual(got[0].Rect, r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 20})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !r2.RectApproxEqual(got[1].Rect, r2.RectFromPoints(r2.Point{X: 0, Y: 20}, r2.Point{X: 40, Y: 40})) {
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
	got, _ := tiling.HorizontalSplit{}.Tile(rect, fileTree, 0)

	if !r2.RectApproxEqual(got[0].Rect, r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 50, Y: 20})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !r2.RectApproxEqual(got[1].Rect, r2.RectFromPoints(r2.Point{X: 10, Y: 20}, r2.Point{X: 50, Y: 40})) {
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
	got, _ := tiling.VerticalSplit{}.Tile(rect, fileTree, 0)

	if len(got) != 2 {
		t.Errorf("expected 2 children, got %v", len(got))
	}
	if !r2.RectApproxEqual(got[0].Rect, r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 10, Y: 30})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !r2.RectApproxEqual(got[1].Rect, r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 30, Y: 30})) {
		t.Errorf("got %v", got[1].Rect)
	}
}

func TestTreemap_FindNode(t *testing.T) {
	root := &Treemap{
		File: files.File{Path: "foo"},
	}
	inner := &Treemap{
		File: files.File{Path: "foo/bar"},
	}
	leaf := &Treemap{
		File: files.File{Path: "foo/bar/baz"},
	}
	root.Children = append(root.Children, inner)
	inner.Children = append(inner.Children, leaf)

	tests := []struct {
		name    string
		argPath string
		want    *Treemap
		wantErr bool
	}{
		{
			name:    "finds valid inner node",
			argPath: "foo/bar",
			want:    inner,
			wantErr: false,
		},
		{
			name:    "finds valid leaf node",
			argPath: "foo/bar/baz",
			want:    leaf,
			wantErr: false,
		},
		{
			name:    "error on miss",
			argPath: "foo/i_am_not_in_tree",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := root.FindNode(tt.argPath)
			if tt.wantErr {
				assert.Error(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
