package dux

import (
	"context"
	"os"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
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

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, nil, files.NewFS(), false)
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
	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{}, files.NewFS(), false)
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

	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, State{}, mockTiler{}, files.NewFS(), false)
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

	pres := NewPresenter(context.Background(), cancel, fileEvents, nil, stateEvents, State{}, mockTiler{}, files.NewFS(), false)
	pres.tick()
	_, ok := <-stateEvents
	assert.True(t, ok, "no StateEvent sent")
}

func Test_ExitAfterScanSetsQuitOnChannelClose(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateEvents := make(chan StateEvent, 5)
	commands := make(chan Command, 1)

	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}
	close(fileEvents)

	initState := State{IsWalkingFiles: true}
	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, initState, mockTiler{}, files.NewFS(), true)
	pres.tick() // processes file event

	// Simulate the initial resize from tcell
	commands <- Resize{
		AppSize:     z2.Point{X: 80, Y: 24},
		TreemapSize: z2.Point{X: 80, Y: 23},
	}
	pres.tick() // processes resize, sets screenReady
	pres.tick() // processes channel close
	pres.tick() // screenReady + !IsWalkingFiles -> Quit

	var lastEvent StateEvent
	for event := range stateEvents {
		lastEvent = event
	}
	assert.True(t, lastEvent.State.Quit, "expected Quit after scan completes with exitAfterScan")
	assert.False(t, lastEvent.State.IsWalkingFiles, "expected IsWalkingFiles to be false")
}

func Test_ExitAfterScanWaitsForResize(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateEvents := make(chan StateEvent, 4)
	commands := make(chan Command, 1)

	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}
	close(fileEvents)

	// TreemapSize starts at zero (no resize yet)
	initState := State{IsWalkingFiles: true}
	pres := NewPresenter(context.Background(), cancel, fileEvents, commands, stateEvents, initState, mockTiler{}, files.NewFS(), true)
	pres.tick() // processes file event
	pres.tick() // processes channel close, sets IsWalkingFiles=false

	// Drain events so far — none should have Quit
	event1 := <-stateEvents
	assert.False(t, event1.State.Quit)
	event2 := <-stateEvents
	assert.False(t, event2.State.Quit, "should not quit before resize")

	// Now send a resize command
	commands <- Resize{
		AppSize:     z2.Point{X: 80, Y: 24},
		TreemapSize: z2.Point{X: 80, Y: 23},
	}
	pres.tick() // processes resize, sets screenReady
	pres.tick() // screenReady + !IsWalkingFiles -> Quit

	event3 := <-stateEvents
	assert.False(t, event3.State.Quit, "resize tick should not quit yet")
	event4 := <-stateEvents
	assert.True(t, event4.State.Quit, "should quit after resize when screenReady")
}

func Test_QuitCommandUpdatesQuitState(t *testing.T) {
	stateEvents := make(chan StateEvent, 1)
	commands := make(chan Command, 1)
	pres := NewPresenter(context.Background(), cancel, nil, commands, stateEvents, State{}, mockTiler{}, files.NewFS(), false)
	commands <- Quit{}
	pres.tick()

	update, ok := <-stateEvents
	require.True(t, ok)
	assert.True(t, update.State.Quit, "quit flag not set")
}
