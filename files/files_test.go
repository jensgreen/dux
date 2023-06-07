package files

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	event := <-ch
	assert.Error(t, event.Error)
}

func TestWalkDir_ClosesChannelWhenDone(t *testing.T) {
	ch := make(chan FileEvent, 1)
	emptyReadDir := func(dirname string) ([]os.DirEntry, error) {
		return make([]os.DirEntry, 0), nil
	}

	WalkDir("", ch, emptyReadDir)

	_, ok := <-ch
	assert.True(t, ok, "expected root dir")
	_, ok = <-ch
	assert.False(t, ok, "expected closed channel")
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
		require.True(t, ok)
		assert.Equal(t, test.err, event.Error)
		assert.Equal(t, test.path, event.File.Path)
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
		require.True(t, ok)
		if assert.NoError(t, result.Error) {
			assert.Equal(t, test, result.File.Path)
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
		require.True(t, ok)
		f, err := event.File, event.Error
		if assert.NoError(t, err) {
			assert.Equal(t, test.path, f.Path)
			assert.Equal(t, test.size, f.Size)
		}
	}
}
