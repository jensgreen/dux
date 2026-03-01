package files

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	// Suppress log output during benchmarks to avoid polluting results.
	log.SetOutput(io.Discard)
}

// mockDirEntry implements os.DirEntry for benchmark use.
type mockDirEntry struct {
	name  string
	isDir bool
	size  int64
}

func (m mockDirEntry) Name() string              { return m.name }
func (m mockDirEntry) IsDir() bool               { return m.isDir }
func (m mockDirEntry) Type() fs.FileMode         { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return mockFileInfo{size: m.size}, nil }

type mockFileInfo struct {
	size int64
}

func (m mockFileInfo) Name() string       { return "" }
func (m mockFileInfo) Size() int64        { return m.size }
func (m mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m mockFileInfo) ModTime() time.Time { return time.Time{} }
func (m mockFileInfo) IsDir() bool        { return false }
func (m mockFileInfo) Sys() interface{}   { return nil }

// mockReadDir returns a function that serves synthetic directory entries
// from an in-memory map, avoiding all disk I/O.
func mockReadDir(dirs map[string][]os.DirEntry) ReadDir {
	return func(dirname string) ([]os.DirEntry, error) {
		entries, ok := dirs[dirname]
		if !ok {
			return nil, nil
		}
		return entries, nil
	}
}

// buildMockDirMap creates a flat directory with n files for use with mockReadDir.
func buildMockDirMap(basePath string, n int) map[string][]os.DirEntry {
	dirs := make(map[string][]os.DirEntry)
	entries := make([]os.DirEntry, n)
	for i := 0; i < n; i++ {
		entries[i] = mockDirEntry{
			name:  fmt.Sprintf("f%d", i),
			isDir: false,
			size:  int64(i + 1),
		}
	}
	dirs[basePath] = entries
	return dirs
}

// buildMockDirMapNested creates a nested directory tree for use with mockReadDir.
func buildMockDirMapNested(basePath string, branching, depth int) map[string][]os.DirEntry {
	dirs := make(map[string][]os.DirEntry)
	buildMockLevel(dirs, basePath, branching, depth)
	return dirs
}

func buildMockLevel(dirs map[string][]os.DirEntry, path string, branching, remainingDepth int) {
	entries := make([]os.DirEntry, branching)
	for i := 0; i < branching; i++ {
		name := fmt.Sprintf("d%d", i)
		if remainingDepth == 0 {
			entries[i] = mockDirEntry{name: name, isDir: false, size: 1}
		} else {
			entries[i] = mockDirEntry{name: name, isDir: true}
			childPath := filepath.Join(path, name)
			buildMockLevel(dirs, childPath, branching, remainingDepth-1)
		}
	}
	dirs[path] = entries
}

// BenchmarkWalkDir measures the file walker throughput using a mock ReadDir
// that serves entries from memory, isolating channel + tree insertion costs.
// A temp directory is created so os.Stat succeeds on the root path.
func BenchmarkWalkDir(b *testing.B) {
	for _, n := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("flat/n=%d", n), func(b *testing.B) {
			basePath := b.TempDir()
			dirs := buildMockDirMap(basePath, n)
			readDir := mockReadDir(dirs)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				fileEvents := make(chan FileEvent, n+2)
				ctx := context.Background()
				b.StartTimer()

				go WalkDir(ctx, basePath, fileEvents, readDir)
				for range fileEvents {
				}
			}
		})
	}
}

// BenchmarkWalkDir_Nested measures walker throughput on a nested directory tree.
func BenchmarkWalkDir_Nested(b *testing.B) {
	cases := []struct {
		name      string
		branching int
		depth     int
	}{
		{"b=5_d=3", 5, 3},   // 125 leaves
		{"b=10_d=3", 10, 3}, // 1000 leaves
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			basePath := b.TempDir()
			dirs := buildMockDirMapNested(basePath, tc.branching, tc.depth)
			readDir := mockReadDir(dirs)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				fileEvents := make(chan FileEvent, 10000)
				ctx := context.Background()
				b.StartTimer()

				go WalkDir(ctx, basePath, fileEvents, readDir)
				for range fileEvents {
				}
			}
		})
	}
}

// BenchmarkWalkDirToFS measures the combined cost of walking + inserting
// into the FS — the real producer side of the pipeline.
func BenchmarkWalkDirToFS(b *testing.B) {
	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			basePath := b.TempDir()
			dirs := buildMockDirMap(basePath, n)
			readDir := mockReadDir(dirs)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				fileEvents := make(chan FileEvent, n+2)
				ctx := context.Background()
				fsys := NewFS()
				b.StartTimer()

				go WalkDir(ctx, basePath, fileEvents, readDir)
				for event := range fileEvents {
					if event.Error == nil {
						fsys.Insert(event.File)
					}
				}
			}
		})
	}
}
