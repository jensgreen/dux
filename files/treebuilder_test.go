package files_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FirstInsertedIsRoot(t *testing.T) {
	tb := files.NewTreeBuilder()
	f := files.File{Path: "foo"}
	tb.Insert(f)

	got, ok := tb.Root()

	assert.True(t, ok)
	assert.Equal(t, f, got.File())
}

func Test_InsertChild(t *testing.T) {
	tb := files.NewTreeBuilder()
	foo := files.File{Path: "foo"}
	bar := files.File{Path: "foo/bar"}
	tb.Insert(foo)
	tb.Insert(bar)

	root, ok := tb.Root()
	require.True(t, ok)
	child := root.Children()[0]
	rootPath := root.File().Path
	childPath := child.File().Path

	assert.Equal(t, rootPath, "foo")
	assert.Equal(t, childPath, "foo/bar")
}

func Test_InsertErrorOnUncleanPath(t *testing.T) {
	tb := files.NewTreeBuilder()
	// "./foo" has shorter equivalent "foo"
	f := files.File{Path: "./foo"}
	err := tb.Insert(f)
	assert.Error(t, err)
}

func Test_InsertBubblesUpSize(t *testing.T) {
	tests := []struct {
		files []files.File
		want  []int64
	}{
		{
			[]files.File{
				{Path: ".", Size: 1},
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
			tb := files.NewTreeBuilder()
			for _, f := range tt.files {
				err := tb.Insert(f)
				require.NoError(t, err)
			}
			got := make([]int64, len(tt.want))
			for w := range tt.want {
				node, ok := tb.Find(tt.files[w].Path)
				require.True(t, ok)
				got[w] = node.File().Size
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("Input files: %v", tt.files)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
