package dux

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"testing"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
	"github.com/jensgreen/dux/geo/z2"
	"github.com/jensgreen/dux/treemap"
	"github.com/jensgreen/dux/treemap/tiling"
)

func init() {
	// Suppress log output during benchmarks to avoid polluting results.
	log.SetOutput(io.Discard)
}

// BenchmarkPresenterTick measures the cost of a single presenter tick —
// the core loop iteration that rebuilds the treemap from the current FS state.
// This is the critical path: called once per file event during scanning.
func BenchmarkPresenterTick(b *testing.B) {
	tiler := tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Left: 1})

	for _, n := range []int{100, 1000, 5000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			// Pre-populate the FS with n files
			fs := files.NewFS()
			fs.Insert(files.File{Path: "root", Size: 0, IsDir: true})
			for i := 0; i < n; i++ {
				fs.Insert(files.File{
					Path: filepath.Join("root", fmt.Sprintf("f%d", i)),
					Size: int64(i + 1),
				})
			}

			// One more file event is waiting to be processed
			fileEvents := make(chan files.FileEvent, 1)
			stateEvents := make(chan StateEvent, 1)

			initialState := State{
				TreemapSize: z2.Point{X: 200, Y: 60},
				MaxDepth:    0,
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fileEvents <- files.FileEvent{
					File: files.File{
						Path: filepath.Join("root", fmt.Sprintf("bench%d", i)),
						Size: 100,
					},
				}

				pres := NewPresenter(
					context.Background(),
					func() {},
					fileEvents,
					nil,
					stateEvents,
					initialState,
					tiler,
					fs,
				)
				pres.tick()
				<-stateEvents
			}
		})
	}
}

// BenchmarkPresenterPipeline measures the throughput of feeding N file events
// through the presenter loop and collecting the resulting state events.
// This reveals the cumulative cost: treemap is rebuilt on every single event.
func BenchmarkPresenterPipeline(b *testing.B) {
	tiler := tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Left: 1})

	for _, n := range []int{100, 500, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			// Pre-generate file events
			events := make([]files.FileEvent, n+1)
			events[0] = files.FileEvent{File: files.File{Path: "root", Size: 0, IsDir: true}}
			for i := 1; i <= n; i++ {
				events[i] = files.FileEvent{
					File: files.File{
						Path: filepath.Join("root", fmt.Sprintf("f%d", i)),
						Size: int64(i),
					},
				}
			}

			initialState := State{
				TreemapSize: z2.Point{X: 200, Y: 60},
				MaxDepth:    0,
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fileEvents := make(chan files.FileEvent, n+1)
				stateEvents := make(chan StateEvent, n+2)
				commands := make(chan Command, 1)

				for _, e := range events {
					fileEvents <- e
				}
				close(fileEvents)

				fs := files.NewFS()
				pres := NewPresenter(
					context.Background(),
					func() {},
					fileEvents,
					commands,
					stateEvents,
					initialState,
					tiler,
					fs,
				)

				// Process all file events + channel close
				for j := 0; j <= n; j++ {
					pres.tick()
				}

				// Drain state events
				for j := 0; j <= n; j++ {
					<-stateEvents
				}

				// Quit
				commands <- Quit{}
				pres.tick()
			}
		})
	}
}

// BenchmarkPresenterTick_WithSelection measures tick cost when a selection
// is active, which adds a FindNode call to each tick.
func BenchmarkPresenterTick_WithSelection(b *testing.B) {
	tiler := tiling.WithPadding(tiling.SliceAndDice{}, tiling.Padding{Top: 1, Left: 1})

	for _, n := range []int{100, 1000} {
		b.Run(fmt.Sprintf("n=%d", n), func(b *testing.B) {
			fs := files.NewFS()
			fs.Insert(files.File{Path: "root", Size: 0, IsDir: true})
			for i := 0; i < n; i++ {
				fs.Insert(files.File{
					Path: filepath.Join("root", fmt.Sprintf("f%d", i)),
					Size: int64(i + 1),
				})
			}

			// Build an initial treemap to get a valid selection
			rect := r2.RectFromPoints(r2.Point{X: 0, Y: 0}, r2.Point{X: 200, Y: 60})
			root, _ := fs.Root()
			initialTreemap := treemap.NewR2Treemap(*root, rect, tiler, 0)

			var selection *treemap.R2Treemap
			if len(initialTreemap.Children) > 0 {
				selection = initialTreemap.Children[len(initialTreemap.Children)/2]
			}

			fileEvents := make(chan files.FileEvent, 1)
			stateEvents := make(chan StateEvent, 1)

			initialState := State{
				TreemapSize: z2.Point{X: 200, Y: 60},
				MaxDepth:    0,
				Selection:   selection,
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				fileEvents <- files.FileEvent{
					File: files.File{
						Path: filepath.Join("root", fmt.Sprintf("bench%d", i)),
						Size: 100,
					},
				}

				pres := NewPresenter(
					context.Background(),
					func() {},
					fileEvents,
					nil,
					stateEvents,
					initialState,
					tiler,
					fs,
				)
				pres.tick()
				<-stateEvents
			}
		})
	}
}
