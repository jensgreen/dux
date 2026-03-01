package nav

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/treemap/tiling"
)

var benchRect = r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 60})

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

func BenchmarkNavigate(b *testing.B) {
	tiler := tiling.SliceAndDice{}

	directions := []struct {
		name string
		dir  Direction
	}{
		{"Left", DirectionLeft},
		{"Right", DirectionRight},
		{"Up", DirectionUp},
		{"Down", DirectionDown},
		{"In", DirectionIn},
		{"Out", DirectionOut},
	}

	for _, size := range []struct {
		name      string
		branching int
		depth     int
	}{
		{"small/b=5_d=2", 5, 2},
		{"medium/b=10_d=3", 10, 3},
	} {
		tree := buildBalancedTree("root", size.branching, size.depth)
		tm := treemap.NewR2Treemap(*tree, benchRect, tiler, 0)

		// Navigate from a middle child to test sibling traversal
		var start *treemap.R2Treemap
		if len(tm.Children) > 1 {
			start = tm.Children[len(tm.Children)/2]
		} else {
			start = tm
		}

		for _, dir := range directions {
			b.Run(fmt.Sprintf("%s/%s", size.name, dir.name), func(b *testing.B) {
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					Navigate(start, dir.dir)
				}
			})
		}
	}
}

func BenchmarkNavigate_ManySiblings(b *testing.B) {
	tiler := tiling.VerticalSplit{}

	for _, n := range []int{10, 100, 500} {
		b.Run(fmt.Sprintf("siblings=%d", n), func(b *testing.B) {
			root := files.NewFileTree(files.File{Path: "root", Size: int64(n), IsDir: true, NumDescendants: n})
			for i := 0; i < n; i++ {
				child := files.NewFileTree(files.File{
					Path: filepath.Join("root", fmt.Sprintf("f%d", i)),
					Size: 1,
				})
				child.SetParent(root)
				root.AddChildren(child)
			}
			// Use a wide rect so children don't fall below MINIMUM_WIDTH
			wideRect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: float64(n) * 10, Y: 60})
			tm := treemap.NewR2Treemap(*root, wideRect, tiler, 0)

			if len(tm.Children) == 0 {
				b.Skip("no visible children at this size")
			}
			// Start from the middle child
			mid := tm.Children[len(tm.Children)/2]

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Navigate(mid, DirectionRight)
			}
		})
	}
}
