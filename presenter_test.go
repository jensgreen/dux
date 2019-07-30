package dux

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/geo/r2"
	"github.com/jensgreen/dux/files"
)

type mockTiler struct{}

func (t mockTiler) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	return make([]Tile, len(fileTree.Children)), r2.EmptyRect()
}

func Test_TickProducesStateUpdateEvent(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateUpdates := make(chan StateUpdate, 1)
	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}
	close(fileEvents)

	pres := NewPresenter(fileEvents, nil, stateUpdates, State{}, nil)
	pres.tick()

	stateUpdate, ok := <-stateUpdates
	if !ok {
		t.Errorf("expected StateUpdate")
	}
	gotPath := stateUpdate.State.Treemap.File.Path
	if "foo" != gotPath {
		t.Errorf("expected StateUpdate for %v, got %v", "foo", gotPath)
	}
}

func Test_WalkDirConcurrencyIntegration(t *testing.T) {
	fileEvents := make(chan files.FileEvent)
	commands := make(chan Command)
	go func() {
		files.WalkDir("testdata/example/inner", fileEvents, os.ReadDir)
		commands <- Quit{}
	}()

	stateUpdates := make(chan StateUpdate)
	pres := NewPresenter(fileEvents, commands, stateUpdates, State{}, mockTiler{})
	go pres.Loop()

	for e := range stateUpdates {
		f := e.State.Treemap.File
		if !strings.Contains(f.Path, "testdata/example/inner") {
			t.Fail()
		}
	}

	if _, ok := <-fileEvents; ok {
		t.Errorf("expected closed channel")
	}
}

func Test_EmitsStateUpdateForRootOnEachFileEvent(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 2)
	stateUpdates := make(chan StateUpdate, 3)
	commands := make(chan Command, 1)

	parent := files.File{Path: "foo"}
	child := files.File{Path: "foo/bar"}
	fileEvents <- files.FileEvent{File: parent}
	fileEvents <- files.FileEvent{File: child}
	close(fileEvents)

	pres := NewPresenter(fileEvents, commands, stateUpdates, State{}, mockTiler{})
	pres.tick() // foo
	pres.tick() // foo/bar
	pres.tick() // closed
	commands <- Quit{}
	pres.tick() // Quit

	for event := range stateUpdates {
		path := event.State.Treemap.File.Path
		if path != "foo" {
			t.Errorf("expected StateUpdate for root %v, got %v", "foo", path)
		}
	}
}

func Test_EmitsStateUpdateOnFileEvent(t *testing.T) {
	fileEvents := make(chan files.FileEvent, 1)
	stateUpdates := make(chan StateUpdate, 1)

	fileEvents <- files.FileEvent{File: files.File{Path: "foo"}}

	pres := NewPresenter(fileEvents, nil, stateUpdates, State{}, mockTiler{})
	pres.tick()
	if _, ok := <-stateUpdates; !ok {
		t.Fail()
	}
}

func Test_QuitCommandUpdatesQuitState(t *testing.T) {
	stateUpdates := make(chan StateUpdate, 1)
	commands := make(chan Command, 1)
	pres := NewPresenter(nil, commands, stateUpdates, State{}, mockTiler{})
	commands <- Quit{}
	pres.tick()

	update, ok := <-stateUpdates
	if !ok {
		t.Errorf("No StateUpdate sent")
	}
	if update.State.Quit != true {
		t.Errorf("Quit flag not set")
	}
}

func Test_FirstAddedIsRoot(t *testing.T) {
	pres := NewPresenter(nil, nil, nil, State{}, nil)
	f := files.File{Path: "foo"}
	pres.add(f)

	want := f
	got := pres.root.File
	if want != got {
		t.Errorf("got %v", got)
	}
}

func Test_AddChild(t *testing.T) {
	pres := NewPresenter(nil, nil, nil, State{}, nil)
	foo := files.File{Path: "foo"}
	bar := files.File{Path: "foo/bar"}
	pres.add(foo)
	pres.add(bar)

	child := pres.root.Children[0]
	rootPath := pres.root.File.Path
	childPath := child.File.Path
	if rootPath != "foo" {
		t.Errorf("got root %v", rootPath)
	}
	if childPath != "foo/bar" {
		t.Errorf("got %v", childPath)
	}
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
			pres := NewPresenter(nil, nil, nil, State{}, nil)
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
				t.Errorf("\ngot  %v,\nwant %v", got, tt.want)
			}
		})
	}
}
