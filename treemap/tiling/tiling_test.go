package tiling

import (
	"reflect"
	"testing"

	"github.com/golang/geo/r1"
	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
)

func TestPadding_PadsAllSides(t *testing.T) {
	rect := r2.Rect{
		X: r1.Interval{Lo: 0.0, Hi: 3.0},
		Y: r1.Interval{Lo: 0.0, Hi: 3.0},
	}
	expected := r2.Rect{
		X: r1.Interval{Lo: 1.0, Hi: 2.0},
		Y: r1.Interval{Lo: 1.0, Hi: 2.0},
	}
	got := Padding{padding: 1.0}.pad(rect)
	if !expected.ApproxEqual(got) {
		t.Errorf("got %+v", got)
	}
}

func TestPadding_ClampToEmpty(t *testing.T) {
	rect := r2.Rect{
		X: r1.Interval{Lo: 0.0, Hi: 2.0},
		Y: r1.Interval{Lo: 1.0, Hi: 1.5},
	}
	got := Padding{padding: 1.0}.pad(rect)
	if !got.IsEmpty() {
		t.Errorf("%+v is not empty", got)
	}
	if !got.IsValid() {
		t.Errorf("%+v is not valid", got)
	}
}

func TestPadding_ClampsToNonEmptyZeroWidth(t *testing.T) {
	rect := r2.Rect{
		X: r1.Interval{Lo: 0.0, Hi: 2.0},
		Y: r1.Interval{Lo: 0.0, Hi: 3.0},
	}
	got := Padding{padding: 1.0}.pad(rect)
	if got.IsEmpty() {
		t.Errorf("%+v is empty", got)
	}
	if got.X.Length() != 0.0 {
		t.Errorf("%+v != 0.0", got.X.Length())
	}
}

func TestSliceAndDice_SliceOrientationDependsOnDepth(t *testing.T) {
	square := r2.RectFromPoints(
		r2.Point{X: 0.0, Y: 0.0},
		r2.Point{X: 100.0, Y: 100.0},
	)
	fileTree := files.FileTree{
		File: files.File{Size: 2},
		Children: []files.FileTree{
			{File: files.File{Path: "foo", IsDir: false, Size: 1}, Children: nil},
			{File: files.File{Path: "bar", IsDir: false, Size: 1}, Children: nil},
		},
	}

	wantV := []r2.Rect{
		r2.RectFromPoints(
			r2.Point{X: 0.0, Y: 0.0},
			r2.Point{X: 50.0, Y: 100.0},
		),
		r2.RectFromPoints(
			r2.Point{X: 50.0, Y: 0.0},
			r2.Point{X: 100.0, Y: 100.0},
		),
	}
	wantH := []r2.Rect{
		r2.RectFromPoints(
			r2.Point{X: 0.0, Y: 0.0},
			r2.Point{X: 100.0, Y: 50.0},
		),
		r2.RectFromPoints(
			r2.Point{X: 0.0, Y: 50.0},
			r2.Point{X: 100.0, Y: 100.0},
		),
	}

	tests := []struct {
		name  string
		depth int
		want  []r2.Rect
	}{
		{"depth 0 should split horizontally", 0, wantH},
		{"depth 1 should split vertically", 1, wantV},
		{"depth 2 should split horizontally", 2, wantH},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tiler := SliceAndDice{}
			gotTiles, gotSpillage := tiler.Tile(square, fileTree, tt.depth)
			if len(gotTiles) == 0 {
				t.Errorf("no tiles")
			}
			if gotSpillage.Size().Norm() != 0.0 {
				t.Errorf("non-zero spillage: %+v", gotSpillage)
			}
			for i, tile := range gotTiles {
				if !reflect.DeepEqual(tile.Rect, tt.want[i]) {
					t.Errorf("\ngot  %v,\nwant %v", tile.Rect, tt.want[i])
				}
			}
		})
	}
}
