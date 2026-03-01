package tiling

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
)

var benchRect = r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 60})

func buildWideTree(basePath string, n int) *files.FileTree {
	root := files.NewFileTree(files.File{Path: basePath, Size: int64(n), IsDir: true, NumDescendants: n})
	for i := 0; i < n; i++ {
		child := files.NewFileTree(files.File{Path: filepath.Join(basePath, fmt.Sprintf("f%d", i)), Size: 1})
		child.SetParent(root)
		root.AddChildren(child)
	}
	return root
}

func BenchmarkVerticalSplit(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tree := buildWideTree("root", n)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				VerticalSplit{}.Tile(benchRect, *tree, 1)
			}
		})
	}
}

func BenchmarkHorizontalSplit(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tree := buildWideTree("root", n)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				HorizontalSplit{}.Tile(benchRect, *tree, 1)
			}
		})
	}
}

func BenchmarkSliceAndDice(b *testing.B) {
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tree := buildWideTree("root", n)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				SliceAndDice{}.Tile(benchRect, *tree, 1)
			}
		})
	}
}

func BenchmarkWithPadding(b *testing.B) {
	padding := Padding{Top: 1, Right: 0, Bottom: 0, Left: 1}
	for _, n := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			tree := buildWideTree("root", n)
			tiler := WithPadding(SliceAndDice{}, padding)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tiler.Tile(benchRect, *tree, 1)
			}
		})
	}
}
