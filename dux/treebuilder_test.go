package dux

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/stretchr/testify/assert"
)

func Test_FirstAddedIsRoot(t *testing.T) {
	tb := NewTreeBuilder()
	f := files.File{Path: "foo"}
	tb.Add(f)

	assert.Equal(t, f, tb.root.File)
}

func Test_AddChild(t *testing.T) {
	tb := NewTreeBuilder()
	foo := files.File{Path: "foo"}
	bar := files.File{Path: "foo/bar"}
	tb.Add(foo)
	tb.Add(bar)

	child := tb.root.Children[0]
	rootPath := tb.root.File.Path
	childPath := child.File.Path

	assert.Equal(t, rootPath, "foo")
	assert.Equal(t, childPath, "foo/bar")
}

func Test_AddBubblesUpSize(t *testing.T) {
	tests := []struct {
		files []files.File
		want  []int64
	}{
		{
			[]files.File{
				{Path: ".", Size: 1},
				{Path: "./foo", Size: 2},
				{Path: "./foo/bar", Size: 4},
			},
			[]int64{7, 6, 4},
		},
		{
			[]files.File{
				{Path: "", Size: 1},
				{Path: "foo", Size: 2},
				{Path: "foo/bar", Size: 4},
			},
			[]int64{7, 6, 4},
		},
		{
			[]files.File{
				{Path: "/", Size: 1},
				{Path: "/foo", Size: 2},
				{Path: "/foo/bar", Size: 4},
			},
			[]int64{7, 6, 4},
		},
		{
			[]files.File{
				{Path: ".git", Size: 1},
				{Path: ".git/foo", Size: 2},
				{Path: ".git/foo/bar", Size: 4},
			},
			[]int64{7, 6, 4},
		},
		{
			[]files.File{
				{Path: "..", Size: 1},
				{Path: "../foo", Size: 2},
				{Path: "../foo/bar", Size: 4},
			},
			[]int64{7, 6, 4},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			tb := NewTreeBuilder()
			cleanFiles := make([]files.File, len(tt.files))
			for j, f := range tt.files {
				cleanFiles[j] = normalize(f)
			}
			for _, f := range cleanFiles {
				tb.Add(f)
			}
			got := make([]int64, len(tt.want))
			for w := range tt.want {
				cleanPath := filepath.Clean(cleanFiles[w].Path)
				got[w] = tb.pathLookup[cleanPath].File.Size
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("Input files: %v", tt.files)
				t.Logf("Cleaned paths: %v", cleanFiles)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
