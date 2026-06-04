package tiling

import (
	"math"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
)

// sampleTree builds a flat node whose parent size equals the sum of its
// children, matching dux's directory-size invariant.
func sampleTree(sizes ...int64) *files.FileTree {
	var total int64
	root := files.NewFileTree(files.File{Path: "root"})
	for i, s := range sizes {
		total += s
		root.AddChildren(files.NewFileTree(files.File{
			Path: "child" + string(rune('a'+i)),
			Size: s,
		}))
	}
	rootFile := root.File()
	rootFile.Size = total
	out := files.NewFileTree(rootFile)
	out.AddChildren(root.Children()...)
	return out
}

func area(r r2.Rect) float64 {
	return r.X.Length() * r.Y.Length()
}

func overlaps(a, b r2.Rect) bool {
	const eps = 1e-9
	return a.X.Lo < b.X.Hi-eps && b.X.Lo < a.X.Hi-eps &&
		a.Y.Lo < b.Y.Hi-eps && b.Y.Lo < a.Y.Hi-eps
}

// assertSpaceFilling checks that tiles do not overlap and that their combined
// area plus the spillage accounts for the whole input rectangle.
func assertSpaceFilling(t *testing.T, name string, rect r2.Rect, tiles []Tile, spillage r2.Rect) {
	t.Helper()
	for i := range tiles {
		if !rectContains(rect, tiles[i].Rect) {
			t.Errorf("%s: tile %d %+v escapes %+v", name, i, tiles[i].Rect, rect)
		}
		for j := i + 1; j < len(tiles); j++ {
			if overlaps(tiles[i].Rect, tiles[j].Rect) {
				t.Errorf("%s: tiles %d and %d overlap: %+v %+v", name, i, j, tiles[i].Rect, tiles[j].Rect)
			}
		}
	}
	var covered float64
	for _, tile := range tiles {
		covered += area(tile.Rect)
	}
	// Spillage is recorded as negative-extent intervals; its size measures the
	// hidden length along one axis. Account for it as leftover area.
	hidden := math.Abs(spillage.X.Length())*rect.Y.Length() + math.Abs(spillage.Y.Length())*rect.X.Length()
	want := area(rect)
	if got := covered + hidden; math.Abs(got-want) > want*0.02 {
		t.Errorf("%s: covered area %.2f + hidden %.2f = %.2f, want ~%.2f", name, covered, hidden, got, want)
	}
}

func rectContains(outer, inner r2.Rect) bool {
	const eps = 1e-6
	return inner.X.Lo >= outer.X.Lo-eps && inner.X.Hi <= outer.X.Hi+eps &&
		inner.Y.Lo >= outer.Y.Lo-eps && inner.Y.Hi <= outer.Y.Hi+eps
}

func TestLayouts_AreSpaceFilling(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 120})
	tree := sampleTree(50, 30, 20, 15, 10, 8, 5, 3)

	layouts := map[string]Sequential{
		"slice & dice": SliceAndDice(),
		"strip":        Strip(),
		"squarified":   Squarified(),
		"pivot":        Pivot(),
	}
	for name, tiler := range layouts {
		for depth := 0; depth < 3; depth++ {
			tiles, spillage := tiler.Tile(rect, *tree, depth)
			if len(tiles) == 0 {
				t.Errorf("%s (depth %d): produced no tiles", name, depth)
			}
			assertSpaceFilling(t, name, rect, tiles, spillage)
		}
	}
}

// TestSquarified_ImprovesAspectRatio checks that squarified produces tiles
// closer to squares than slice & dice on a skewed input.
func TestSquarified_ImprovesAspectRatio(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 120})
	tree := sampleTree(100, 60, 40, 25, 15, 10)

	worstAspect := func(tiles []Tile) float64 {
		worst := 1.0
		for _, tile := range tiles {
			w, h := tile.Rect.X.Length(), tile.Rect.Y.Length()
			if w == 0 || h == 0 {
				continue
			}
			ar := w / h
			if ar < 1 {
				ar = 1 / ar
			}
			if ar > worst {
				worst = ar
			}
		}
		return worst
	}

	sdTiles, _ := SliceAndDice().Tile(rect, *tree, 0)
	sqTiles, _ := Squarified().Tile(rect, *tree, 0)
	if worstAspect(sqTiles) >= worstAspect(sdTiles) {
		t.Errorf("squarified worst aspect %.2f not better than slice & dice %.2f",
			worstAspect(sqTiles), worstAspect(sdTiles))
	}
}

// TestSequential_SpillageHidesTinyItems verifies the small-file hiding
// behaviour carries over to the functor engine.
func TestSequential_SpillageHidesTinyItems(t *testing.T) {
	rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 100, Y: 100})
	// At depth 0 items stack vertically; the last item is far below MINIMUM_HEIGHT.
	tree := sampleTree(100, 1)
	tiles, spillage := SliceAndDice().Tile(rect, *tree, 0)
	if len(tiles) != 1 {
		t.Fatalf("expected 1 visible tile, got %d", len(tiles))
	}
	if spillage.Y.Length() == 0 {
		t.Errorf("expected non-zero spillage for hidden tiny item, got %+v", spillage)
	}
}
