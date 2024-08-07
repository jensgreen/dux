package dux

import (
	"context"
	"os"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/treemap/tiling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTiler struct{}

func (t mockTiler) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []tiling.Tile, spillage r2.Rect) {
	return make([]tiling.Tile, len(fileTree.Children())), r2.Rect{}
}

func cancel() {}

func Test_TickProducesStateEvents(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateEvents := make(chan StateEvent, 1)
	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}
	close(fileEvents)

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, nil, files.NewFS())
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
	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{}, files.NewFS())
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

	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{}, files.NewFS())
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

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, mockTiler{}, files.NewFS())
	pres.tick()
	_, ok := <-stateEvents
	assert.True(t, ok, "no StateEvent sent")
}

func Test_QuitCommandUpdatesQuitState(t *testing.T) {
	stateEvents := make(chan StateEvent, 1)
	commands := make(chan Command, 1)
	pres := NewPresenter(context.Background(), cancel, nil, commands, stateEvents, State{}, mockTiler{}, files.NewFS())
	commands <- Quit{}
	pres.tick()

	update, ok := <-stateEvents
	require.True(t, ok)
	assert.True(t, update.State.Quit, "quit flag not set")
}
