package dux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap/tiling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTiler struct{}

func (t mockTiler) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []tiling.Tile, spillage r2.Rect) {
	return make([]tiling.Tile, len(fileTree.Children)), r2.Rect{}
}

func cancel() {}

func Test_TickProducesStateEvents(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateEvents := make(chan StateEvent, 1)
	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}
	close(fileEvents)

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, nil)
	pres.tick()

	stateEvent, ok := <-stateEvents
	require.True(t, ok)
	assert.Equal(t, "foo", stateEvent.State.Treemap.Path())
}

func Test_WalkDirConcurrencyIntegration(t *testing.T) {
	fileEvents := make(chan files.FileEvent)
	commands := make(chan Command)
	go func() {
		files.WalkDir(context.Background(), "../testdata/example/inner", fileEvents, os.ReadDir)
		commands <- Quit{}
	}()

	stateEvents := make(chan StateEvent)
	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{})
	go pres.Loop()

	for e := range stateEvents {
		f := e.State.Treemap.File
		assert.Contains(t, f.Path, "../testdata/example/inner")
	}

	_, ok := <-fileEvents
	assert.False(t, ok, "expected closed channel")
}

func Test_EmitsStateEventForRootOnEachFileEvent(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 2)
	stateEvents := make(chan StateEvent, 4)
	commands := make(chan Command, 1)

	parent := files.File{Path: "foo"}
	child := files.File{Path: "foo/bar"}
	fileEvents <- files.FileEvent{File: parent}
	fileEvents <- files.FileEvent{File: child}
	close(fileEvents)

	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{})
	pres.tick() // foo
	pres.tick() // foo/bar
	pres.tick() // closed
	commands <- Quit{}
	pres.tick() // Quit

	for event := range stateEvents {
		path := event.State.Treemap.Path()
		assert.Equal(t, path, "foo", "expected StateEvent for root")
	}
}

func Test_EmitsStateEventOnFileEvent(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateEvents := make(chan StateEvent, 1)

	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, mockTiler{})
	pres.tick()
	_, ok := <-stateEvents
	assert.True(t, ok, "no StateEvent sent")
}

func Test_QuitCommandUpdatesQuitState(t *testing.T) {
	stateEvents := make(chan StateEvent, 1)
	commands := make(chan Command, 1)
	pres := NewPresenter(context.Background(), cancel, nil, commands, stateEvents, State{}, mockTiler{})
	commands <- Quit{}
	pres.tick()

	update, ok := <-stateEvents
	require.True(t, ok)
	assert.True(t, update.State.Quit, "quit flag not set")
}

func Test_FirstAddedIsRoot(t *testing.T) {
	pres := NewPresenter(context.Background(), cancel, nil, nil, nil, State{}, nil)
	f := files.File{Path: "foo"}
	pres.add(f)

	assert.Equal(t, f, pres.root.File)
}

func Test_AddChild(t *testing.T) {
	pres := NewPresenter(context.Background(), cancel, nil, nil, nil, State{}, nil)
	foo := files.File{Path: "foo"}
	bar := files.File{Path: "foo/bar"}
	pres.add(foo)
	pres.add(bar)

	child := pres.root.Children[0]
	rootPath := pres.root.File.Path
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
			pres := NewPresenter(context.Background(), cancel, nil, nil, nil, State{}, nil)
			cleanFiles := make([]files.File, len(tt.files))
			for j, f := range tt.files {
				cleanFiles[j] = normalize(f)
			}
			for _, f := range cleanFiles {
				pres.add(f)
			}
			got := make([]int64, len(tt.want))
			for w := range tt.want {
				cleanPath := filepath.Clean(cleanFiles[w].Path)
				got[w] = pres.pathLookup[cleanPath].File.Size
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("Input files: %v", tt.files)
				t.Logf("Cleaned paths: %v", cleanFiles)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
