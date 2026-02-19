# CLAUDE.md

This file provides guidance for AI assistants working with the **dux** codebase.

## Project Overview

dux is a disk usage analyzer that displays results as an interactive
terminal treemap. It is a Go application built with `tcell/v2` for
terminal UI rendering. The user runs `dux [DIRECTORY]` and gets a
live, navigable treemap of disk usage.

## Build & Test Commands

```sh
go build -v ./...     # Build all packages
go test -v ./...      # Run all tests
go test ./files/...   # Run tests for a specific package
go test -run TestName ./...      # Run a specific test
```

There is no Makefile or task runner. Standard `go` commands are used
throughout.

## Project Structure

```
main.go                 Entry point: wires up goroutines and channels
app/                    Terminal UI layer (tcell screen, widgets, event loop)
dux/                    Core presenter: state management and command processing
files/                  File system walking, tree building, size formatting
geo/                    Generic geometric primitives (Interval, Point, Rect)
  geo/r1/               1D intervals (float64)
  geo/r2/               2D rectangles (float64)
  geo/z1/               1D intervals (int)
  geo/z2/               2D points/rectangles (int)
nav/                    Keyboard navigation logic for treemap tiles
treemap/                Treemap generation from file trees
  treemap/tiling/       Tiling algorithms (SliceAndDice, splits, padding)
cancellable/            Context-aware channel send/receive helpers
logging/                Debug logging setup (disabled by default)
recovery/               Panic recovery for goroutines
testdata/               Test fixture directory tree
```

## Architecture

The application has three core goroutines coordinated via channels
(plus additional goroutines for signal handling and tcell's internal
screen event handler):

1. **File walker** (`files.WalkDir`) - Recursively walks the target
   directory, emitting `FileEvent`s
2. **Presenter** (`dux.Presenter.Loop`) - Processes file events and
   UI commands, rebuilds treemap state, sends `StateEvent`s
3. **App** (`app.App.Run`) - Renders the terminal UI, handles
   keyboard input, sends `Command`s

Channel flow:
`fileEvents` -> `presenter` -> `stateEvents` -> `app` -> `commands` -> `presenter`

### Coordinate Systems

The codebase uses two coordinate spaces for treemaps:

- **R2** (`geo.Rect[float64]`) - Continuous float coordinates used
  for layout computation
- **Z2** (`geo.Rect[int]`) - Integer coordinates snapped from R2
  for terminal cell rendering

### Tiling

Treemap layout is pluggable via the `tiling.Tiler` interface. The
default algorithm is `SliceAndDice` (alternating horizontal/vertical
splits) wrapped with `Padding`.

## Code Conventions

- **Formatting**: Standard `gofmt`
- **Generics**: Used extensively in `geo/`, `treemap/`, and `nav/`
  packages for type-safe geometry across float64 and int
- **Testing**: Tests use `github.com/stretchr/testify`
  (assert/require). Tests live alongside source in `*_test.go` files.
  Table-driven test style is preferred.
- **Concurrency**: All long-running operations accept
  `context.Context` for cancellation. Channel operations use helpers
  from `cancellable/` to prevent goroutine leaks.
- **Error handling**: Early returns, `PathError` unwrapping for
  user-facing messages
- **Dependencies**: Kept minimal - primary dependencies are `tcell/v2`
  for UI, `testify` for tests, and `golang.org/x/exp` for generic
  utilities

## CI

GitHub Actions (`.github/workflows/ci.yml`) runs on pushes and PRs
to `master`:
- `go build -v ./...`
- `go test -v ./...`

## Key Types

| Type | Package | Purpose |
|------|---------|---------|
| `App` | `app` | Main UI: screen management, event loop, rendering |
| `TreemapWidget` | `app` | Recursive treemap tile renderer |
| `Presenter` | `dux` | Event loop that processes commands and file events |
| `State` | `dux` | Immutable snapshot: treemap, selection, zoom |
| `Command` | `dux` | Interface for user actions (Navigate, etc.) |
| `FileTree` | `files` | Hierarchical file/directory representation |
| `FS` | `files` | Maintains file tree with O(1) path lookup |
| `Treemap[T]` | `treemap` | Generic treemap node with bounds/children |
| `Tiler` | `treemap/tiling` | Interface for layout algorithms |

## Navigation Directions

The `nav` package supports six directions: `Left`, `Right`, `Up`,
`Down`, `In` (descend into child), `Out` (ascend to parent).
Navigation is orientation-aware, meaning it understands whether a
tile group is split horizontally or vertically.

## Keyboard Controls

| Key | Action |
|-----|--------|
| `h`/`Left` | Navigate left |
| `j`/`Down` | Navigate down |
| `k`/`Up` | Navigate up |
| `l`/`Right` | Navigate right |
| `Enter` | Navigate into selected tile |
| `Backspace` | Navigate out of selected tile |
| `i` | Zoom in |
| `o` | Zoom out |
| `+`/`-` | Increase/decrease visible depth |
| `Space` | Pause/resume file scanning |
| `q`/`Esc`/`Ctrl-C` | Quit |
