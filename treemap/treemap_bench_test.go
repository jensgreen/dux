package treemap

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap/tiling"
)

var benchRect = r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 60})

func BenchmarkNewR2Treemap(b *testing.B) {
	tiler := tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Left: 1, Bottom: 0, Right: 0})

	cases := []struct {
		name      string
		branching int
		depth     int
	}{
		{"balanced/b=10_d=2", 10, 2},    // 100 leaves
		{"balanced/b=10_d=3", 10, 3},    // 1000 leaves
		{"balanced/b=10_d=4", 10, 4},    // 10000 leaves
		{"wide/n=100", 100, 1},          // 100 children, flat
		{"wide/n=1000", 1000, 1},        // 1000 children, flat
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tree := buildBalancedTree("root", tc.branching, tc.depth)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewR2Treemap(*tree, benchRect, tiler, 0)
			}
		})
	}
}

func BenchmarkNewR2Treemap_MaxDepth(b *testing.B) {
	tiler := tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Left: 1, Bottom: 0, Right: 0})
	tree := buildBalancedTree("root", 10, 4) // 10000 leaves

	for _, maxDepth := range []int{2, 3, 4, 0} {
		name := fmt.Sprintf("maxDepth=%d", maxDepth)
		if maxDepth == 0 {
			name = "maxDepth=unlimited"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewR2Treemap(*tree, benchRect, tiler, maxDepth)
			}
		})
	}
}

func BenchmarkNewZ2Treemap(b *testing.B) {
	tiler := tiling.SliceAndDice{}

	cases := []struct {
		name      string
		branching int
		depth     int
	}{
		{"b=10_d=2", 10, 2},
		{"b=10_d=3", 10, 3},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			tree := buildBalancedTree("root", tc.branching, tc.depth)
			r2tm := NewR2Treemap(*tree, benchRect, tiler, 0)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewZ2Treemap(r2tm)
			}
		})
	}
}

func BenchmarkFindNode(b *testing.B) {
	tiler := tiling.SliceAndDice{}

	for _, depth := range []int{2, 3, 4} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			tree := buildBalancedTree("root", 5, depth)
			tm := NewR2Treemap(*tree, benchRect, tiler, 0)

			// Build a path to a deep leaf: root/d0/d0/.../d0
			target := "root"
			for i := 0; i < depth; i++ {
				target = filepath.Join(target, "d0")
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tm.FindNode(target)
			}
		})
	}
}

// buildBalancedTree creates a synthetic FileTree for benchmarking.
func buildBalancedTree(basePath string, branching, depth int) *files.FileTree {
	size := leafCount(branching, depth)
	root := files.NewFileTree(files.File{Path: basePath, Size: int64(size), IsDir: true, NumDescendants: size - 1})
	if depth > 0 {
		addChildren(root, basePath, branching, depth-1)
	}
	return root
}

func addChildren(parent *files.FileTree, parentPath string, branching, remainingDepth int) {
	for i := 0; i < branching; i++ {
		childPath := filepath.Join(parentPath, fmt.Sprintf("d%d", i))
		if remainingDepth == 0 {
			child := files.NewFileTree(files.File{Path: childPath, Size: 1})
			child.SetParent(parent)
			parent.AddChildren(child)
		} else {
			size := leafCount(branching, remainingDepth)
			child := files.NewFileTree(files.File{Path: childPath, Size: int64(size), IsDir: true, NumDescendants: size - 1})
			child.SetParent(parent)
			parent.AddChildren(child)
			addChildren(child, childPath, branching, remainingDepth-1)
		}
	}
}

func leafCount(branching, depth int) int {
	n := 1
	for i := 0; i < depth; i++ {
		n *= branching
	}
	return n
}
