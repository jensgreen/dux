package files

import (
	"fmt"
	"path/filepath"
	"testing"
)

func BenchmarkFSInsert(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("flat/n=%d", n), func(b *testing.B) {
			// Pre-generate file list
			files := make([]File, n+1)
			files[0] = File{Path: "root", Size: 0, IsDir: true}
			for i := 1; i <= n; i++ {
				files[i] = File{
					Path: filepath.Join("root", fmt.Sprintf("f%d", i)),
					Size: int64(i),
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs := NewFS()
				for _, f := range files {
					fs.Insert(f)
				}
			}
		})
	}

	for _, depth := range []int{10, 50, 200} {
		b.Run(fmt.Sprintf("deep/depth=%d", depth), func(b *testing.B) {
			// Pre-generate file list forming a chain
			files := make([]File, depth+1)
			path := "root"
			files[0] = File{Path: path, Size: 0, IsDir: true}
			for i := 1; i <= depth; i++ {
				path = filepath.Join(path, fmt.Sprintf("d%d", i))
				files[i] = File{Path: path, Size: 1, IsDir: true}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs := NewFS()
				for _, f := range files {
					fs.Insert(f)
				}
			}
		})
	}
}

func BenchmarkFSFind(b *testing.B) {
	for _, n := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			fs := populateFS("root", n)
			// Look up the last inserted path
			target := filepath.Join("root", fmt.Sprintf("f%d", n-1))

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fs.Find(target)
			}
		})
	}
}
