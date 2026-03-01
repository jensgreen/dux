# Benchmarks

This document describes how to run benchmarks, profile performance, and
track regressions in dux.

## Running Benchmarks

Run all benchmarks:

```sh
go test -bench=. -benchmem ./...
```

Run benchmarks for a specific package:

```sh
go test -bench=. -benchmem ./treemap/...
go test -bench=. -benchmem ./files/...
go test -bench=. -benchmem ./nav/...
go test -bench=. -benchmem ./dux/...
```

Run a specific benchmark:

```sh
go test -bench=BenchmarkNewR2Treemap -benchmem ./treemap/...
go test -bench=BenchmarkPresenterTick -benchmem ./dux/...
```

## Benchmark Categories

### Micro-benchmarks

| Benchmark | Package | What it measures |
|-----------|---------|------------------|
| `BenchmarkFSInsert` | `files` | FS.Insert with flat and deep trees |
| `BenchmarkFSFind` | `files` | Map-based path lookup |
| `BenchmarkNewR2Treemap` | `treemap` | Float64 treemap construction |
| `BenchmarkNewR2Treemap_MaxDepth` | `treemap` | Effect of depth limiting |
| `BenchmarkNewZ2Treemap` | `treemap` | Float-to-int coordinate snapping |
| `BenchmarkFindNode` | `treemap` | Recursive path-based node lookup |
| `BenchmarkVerticalSplit` | `tiling` | Vertical split tiling |
| `BenchmarkHorizontalSplit` | `tiling` | Horizontal split tiling |
| `BenchmarkSliceAndDice` | `tiling` | Alternating split tiling |
| `BenchmarkWithPadding` | `tiling` | Padded tiling wrapper |
| `BenchmarkNavigate` | `nav` | Directional navigation |
| `BenchmarkNavigate_ManySiblings` | `nav` | Sibling scan with many children |

### Integration benchmarks

| Benchmark | Package | What it measures |
|-----------|---------|------------------|
| `BenchmarkPresenterTick` | `dux` | Single tick cost (treemap rebuild) |
| `BenchmarkPresenterTick_WithSelection` | `dux` | Tick cost with active selection |
| `BenchmarkPresenterPipeline` | `dux` | End-to-end N events through presenter |
| `BenchmarkWalkDir` | `files` | Walker throughput (mock I/O) |
| `BenchmarkWalkDir_Nested` | `files` | Walker with nested directories |
| `BenchmarkWalkDirToFS` | `files` | Walker + FS insertion combined |

## Profiling with pprof

### CPU profiling

```sh
# Profile treemap construction
go test -bench=BenchmarkNewR2Treemap -cpuprofile cpu.prof -benchmem ./treemap/

# Profile the presenter tick
go test -bench=BenchmarkPresenterTick -cpuprofile cpu.prof -benchmem ./dux/

# Analyze
go tool pprof cpu.prof
# In the pprof shell: top, list NewR2Treemap, web (opens flamegraph)
```

### Memory profiling

```sh
# Profile allocations during treemap construction
go test -bench=BenchmarkNewR2Treemap -memprofile mem.prof -benchmem ./treemap/

# Analyze by allocation space
go tool pprof -alloc_space mem.prof
# In the pprof shell: top, list newR2Treemap
```

### Execution tracing

Use `go tool trace` to visualize goroutine scheduling, channel
contention, and GC pauses:

```sh
go test -bench=BenchmarkPresenterPipeline -trace trace.out ./dux/
go tool trace trace.out
```

### What to look for

- **CPU**: Proportion of time in `newR2Treemap` vs `Tile` vs `FindNode`
- **Memory**: Allocation rate per tick — each tick creates a new treemap
- **Trace**: Channel blocking between walker and presenter goroutines
- **GC pressure**: Treemaps rebuilt every tick become garbage immediately

## Regression Tracking with benchstat

Install benchstat:

```sh
go install golang.org/x/perf/cmd/benchstat@latest
```

Collect baseline and compare after changes:

```sh
# Before changes
go test -bench=. -benchmem -count=10 ./... > old.txt

# After changes
go test -bench=. -benchmem -count=10 ./... > new.txt

# Compare
benchstat old.txt new.txt
```

Use `-count=10` (or more) for statistically significant results.
benchstat will report the delta and p-value for each benchmark.
