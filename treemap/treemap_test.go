package treemap

import (
	"fmt"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap/tiling"
	"github.com/stretchr/testify/assert"
)

// buildTree constructs a 3-level file tree (root -> dirs -> files) whose
// directory sizes equal the sum of their children, matching dux's invariant.
func buildTree() files.FileTree {
	dirSizes := [][]int64{
		{50, 30, 20, 15, 12, 10},
		{40, 25, 13, 9, 7},
		{33, 22, 11, 6, 4, 3},
		{18, 9, 5},
	}
	root := files.NewFileTree(files.File{Path: "/root"})
	var total int64
	for d, sizes := range dirSizes {
		dir := files.NewFileTree(files.File{Path: fmt.Sprintf("/root/d%d", d)})
		var ds int64
		for i, s := range sizes {
			ds += s
			dir.AddChildren(files.NewFileTree(files.File{Path: fmt.Sprintf("/root/d%d/f%d", d, i), Size: s}))
		}
		df := dir.File()
		df.Size = ds
		total += ds
		ndir := files.NewFileTree(df)
		ndir.AddChildren(dir.Children()...)
		root.AddChildren(ndir)
	}
	rf := root.File()
	rf.Size = total
	out := files.NewFileTree(rf)
	out.AddChildren(root.Children()...)
	return *out
}

// TestNewR2Treemap_NoSubMinimumNodes guards against the cross-axis sliver bug at
// the level the app actually renders: the recursive treemap with real padding.
// A tile that is thin on the cross axis used to be kept; recursing into it then
// hit padding that emptied its rect, silently dropping the tile's whole subtree.
// Every rendered node must instead be at least the minimum size on both axes.
func TestNewR2Treemap_NoSubMinimumNodes(t *testing.T) {
	tree := buildTree()
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 80, Y: 40})
	pad := tiling.Padding{Top: 1, Right: 1, Bottom: 1, Left: 1}

	var check func(t *testing.T, name string, tm *R2Treemap, root bool)
	check = func(t *testing.T, name string, tm *R2Treemap, root bool) {
		if !root {
			w, h := tm.Rect.X.Length(), tm.Rect.Y.Length()
			if w < tiling.MINIMUM_WIDTH || h < tiling.MINIMUM_HEIGHT {
				t.Errorf("%s: rendered sub-minimum node %s at %.2f x %.2f", name, tm.File.Path, w, h)
			}
		}
		for _, c := range tm.Children {
			check(t, name, c, false)
		}
	}

	for _, l := range tiling.DefaultLayouts(pad) {
		tm := NewR2Treemap(tree, rect, l.Tiler, 0)
		check(t, l.Name, tm, true)
	}
}

func TestTreemapWithTiler_NoChildren(t *testing.T) {
	tree := files.FileTree{}
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	got := NewR2Treemap(tree, rect, tiling.VerticalSplit(), 0)

	expected := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	if !r2.RectApproxEqual(expected, got.Rect) {
		t.Errorf("got %v, expected %v", got, expected)
	}
	if len(got.Children) != 0 {
		t.Errorf("expected no children, got %v", got.Children)
	}
}

func TestTreemapWithTiler_SplitsCorrectly(t *testing.T) {
	fileTree := files.NewFileTree(files.File{Size: 2})
	fileTree.AddChildren(
		files.NewFileTree(files.File{Path: "foo", Size: 1}),
		files.NewFileTree(files.File{Path: "bar", Size: 1}),
	)
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 40, Y: 40})
	got := NewR2Treemap(*fileTree, rect, tiling.VerticalSplit(), 0)

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
	fileTree := files.NewFileTree(files.File{Size: 2})
	fileTree.AddChildren(
		files.NewFileTree(files.File{Size: 1}),
		files.NewFileTree(files.File{Size: 1}),
	)
	got, _ := tiling.VerticalSplit().Tile(rect, *fileTree, 0)

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
	fileTree := files.NewFileTree(files.File{Size: 2})
	fileTree.AddChildren(
		files.NewFileTree(files.File{Size: 1}),
		files.NewFileTree(files.File{Size: 1}),
	)
	got, _ := tiling.HorizontalSplit().Tile(rect, *fileTree, 0)

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
	fileTree := files.NewFileTree(files.File{Size: 2})
	fileTree.AddChildren(
		files.NewFileTree(files.File{Size: 1}),
		files.NewFileTree(files.File{Size: 1}),
	)
	got, _ := tiling.HorizontalSplit().Tile(rect, *fileTree, 0)

	if !r2.RectApproxEqual(got[0].Rect, r2.RectFromPoints(r2.Point{X: 10, Y: 0}, r2.Point{X: 50, Y: 20})) {
		t.Errorf("got %v", got[0].Rect)
	}
	if !r2.RectApproxEqual(got[1].Rect, r2.RectFromPoints(r2.Point{X: 10, Y: 20}, r2.Point{X: 50, Y: 40})) {
		t.Errorf("got %v", got[1].Rect)
	}
}

func TestVerticalSplit_SplitsTwoNoRemainder(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 30, Y: 30})
	fileTree := files.NewFileTree(files.File{Size: 3})
	fileTree.AddChildren(
		files.NewFileTree(files.File{Size: 1}),
		files.NewFileTree(files.File{Size: 2}),
	)
	got, _ := tiling.VerticalSplit().Tile(rect, *fileTree, 0)

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
	root := &R2Treemap{
		File: files.File{Path: "foo"},
	}
	inner := &R2Treemap{
		File: files.File{Path: "foo/bar"},
	}
	leaf := &R2Treemap{
		File: files.File{Path: "foo/bar/baz"},
	}
	root.Children = append(root.Children, inner)
	inner.Children = append(inner.Children, leaf)

	tests := []struct {
		name    string
		argPath string
		want    *R2Treemap
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
