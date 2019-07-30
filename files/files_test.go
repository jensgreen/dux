package files

import (
	"errors"
	"os"
	"testing"
)

// Contents of testdata/ dir:
//		testdata
//	 	└── example
//	 	    ├── inner
//	 	    │   ├── a.txt
//	 	    │   ├── b.txt
//	 	    │   └── nested
//	 	    │       └── innermost.txt
//	 	    └── outer.txt

func TestWalkDir_PropagatesReadDirError(t *testing.T) {
	ch := make(chan FileEvent, 1)
	errorReadDir := func(dirname string) ([]os.DirEntry, error) {
		return nil, errors.New("foo")
	}

	WalkDir("", ch, errorReadDir)

	if event := <-ch; event.Error == nil {
		t.Error("expected error")
	}
}

func TestWalkDir_ClosesChannelWhenDone(t *testing.T) {
	ch := make(chan FileEvent, 1)
	emptyReadDir := func(dirname string) ([]os.DirEntry, error) {
		return make([]os.DirEntry, 0), nil
	}

	WalkDir("", ch, emptyReadDir)

	if _, ok := <-ch; !ok {
		t.Fatalf("expected root dir")
	}
	if _, ok := <-ch; ok {
		t.Error("expected closed channel")
	}
}

func TestWalkDir_ProducesFileEventsIncludingRoot(t *testing.T) {
	ch := make(chan FileEvent, 10)

	WalkDir("../testdata/example/inner/nested", ch, os.ReadDir)

	tests := []struct {
		err  error
		path string
	}{
		{err: nil, path: "../testdata/example/inner/nested"},
		{err: nil, path: "../testdata/example/inner/nested/innermost.txt"},
	}

	for _, test := range tests {
		event, ok := <-ch
		if !ok {
			t.Fatalf("expected open channel")
		}
		if event.Error != test.err {
			t.Fatalf("got error %+v", event.Error)
		}
		if event.File.Path != test.path {
			t.Fatalf("expected %s got %s", test.path, event.File.Path)
		}
	}
}

func TestWalkDir_ProducesFileEventsBreadthFirst(t *testing.T) {
	ch := make(chan FileEvent, 10)

	WalkDir("../testdata/example/inner", ch, os.ReadDir)

	tests := []string{
		"../testdata/example/inner",
		"../testdata/example/inner/a.txt",
		"../testdata/example/inner/b.txt",
		"../testdata/example/inner/nested",
		"../testdata/example/inner/nested/innermost.txt",
	}
	for _, test := range tests {
		result, ok := <-ch
		if !ok {
			t.Fatalf("channel is closed")
		}
		if result.Error != nil {
			t.Fatalf("got error: %+v", result.Error)
		}
		if result.File.Path != test {
			t.Errorf("expected %s, got %s", test, result.File.Path)
		}
	}
}

func TestWalkDir_SetsFileSize(t *testing.T) {
	ch := make(chan FileEvent, 10)

	WalkDir("../testdata/example/inner/nested", ch, os.ReadDir)

	tests := []struct {
		path string
		size int64
	}{
		{"../testdata/example/inner/nested", 0},
		{"../testdata/example/inner/nested/innermost.txt", 15},
	}

	for _, test := range tests {
		event, ok := <-ch
		f, err := event.File, event.Error
		if !ok {
			t.Fatalf("expected closed channel")
		}
		if err != nil {
			t.Fatalf("got error %+v", err)
		}
		if !(f.Path == test.path && f.Size == test.size) {
			t.Errorf("expected (%s %d), got (%s %d)", test.path, test.size, f.Path, f.Size)
		}
	}
}
