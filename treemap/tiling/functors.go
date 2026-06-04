package tiling

import (
	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r2"
)

// The types in this file describe the five independent dimensions that span
// the design space of sequential, rectangular, space-filling layouts, as
// characterized by Baudel & Broeksema, "Capturing the Design Space of
// Sequential Space-Filling Layouts" (IEEE TVCG 2012):
//
//	Order   T -> T                    order in which items are laid out
//	Size    T -> R                    how individual items are sized
//	Score   C x R -> R                local maxima delimit chunks
//	Recurse C -> B                    whether to re-layout inside a chunk
//	Phrase  C -> (Side, Direction)    how chunks are placed in the free space
//
// A Sequential tiler (see sequential.go) is parametrized by one functor for
// each dimension. Well-known layouts (slice & dice, strip, squarified, pivot)
// are particular points in this space; see layouts.go.

// Side is the edge of the remaining available space that a finished chunk is
// placed against. A chunk on the West or East edge spans the full height and
// stacks its items vertically; a chunk on the North or South edge spans the
// full width and stacks its items horizontally.
type Side int

const (
	North Side = iota
	South
	East
	West
)

// vertical reports whether items in a chunk on this side stack along the Y
// axis (true) or the X axis (false).
func (s Side) vertical() bool {
	return s == East || s == West
}

// Direction is the order in which items are stacked within a chunk.
type Direction int

const (
	Down  Direction = iota // vertical stack, top to bottom (increasing Y)
	Up                     // vertical stack, bottom to top (decreasing Y)
	Right                  // horizontal stack, left to right (increasing X)
	Left                   // horizontal stack, right to left (decreasing X)
)

// Config is the placement of a chunk: which side it is laid against and in
// which direction its items are stacked.
type Config struct {
	Side Side
	Dir  Direction
}

// Item is a single thing to lay out at one level of the hierarchy.
type Item struct {
	File *files.FileTree
	Size float64
}

// Chunk accumulates a run of items that will be placed together against one
// side of the available space.
type Chunk struct {
	Items  []Item
	Config Config
	Sum    float64
	Rect   r2.Rect
}

// State is threaded through a single Tile() call. The paper's functors are
// stateful (X x State -> Y x State); State holds the bookkeeping the score and
// phrase functors need.
type State struct {
	Depth        int     // current hierarchy depth
	Available    r2.Rect // free space, shrinks as chunks are placed
	OverallSum   float64 // total size of all items at this level
	RemainingSum float64 // size of items not yet placed
	ItemCount    int     // number of items at this level
	Index        int     // index of the item currently being scored
}

// The five functor signatures.
type (
	OrderFunc   func(items []Item, st *State) []Item
	SizeFunc    func(ft *files.FileTree) float64
	ScoreFunc   func(c *Chunk, itemSize float64, st *State) float64
	RecurseFunc func(c *Chunk, st *State) bool
	PhraseFunc  func(prev *Chunk, st *State) Config
)

// FileSize is the default Size functor: an item's weight is its file size.
func FileSize(ft *files.FileTree) float64 {
	return float64(ft.File().Size)
}

// --- Score functors -------------------------------------------------------

// oneChunk never starts a new chunk: it returns a constant, so the running
// score never decreases and every item lands in a single chunk. Combined with
// byDepthParity this reproduces the classic slice & dice layout.
func oneChunk(c *Chunk, itemSize float64, st *State) float64 {
	return 0
}

// chunkAspect returns the expanding and fixed dimensions of the available
// space relative to a chunk's stacking axis.
func chunkAspect(c *Chunk, st *State) (expanding, fixed float64) {
	if c.Config.Side.vertical() {
		return st.Available.Y.Length(), st.Available.X.Length()
	}
	return st.Available.X.Length(), st.Available.Y.Length()
}

// bestMinAspectRatio scores a chunk by the aspect ratio of its smallest item
// (closest to 1 is best). As items are added the worst item's aspect ratio
// improves then degrades; the local maximum delimits the chunk. This is the
// squarified heuristic (paper, Listing 6).
func bestMinAspectRatio(c *Chunk, itemSize float64, st *State) float64 {
	expanding, fixed := chunkAspect(c, st)
	newSum := c.Sum + itemSize
	if newSum == 0 || st.OverallSum == 0 || fixed == 0 {
		return 0
	}
	minItem := itemSize
	for _, it := range c.Items {
		if it.Size < minItem {
			minItem = it.Size
		}
	}
	ar := expanding * (minItem / newSum) / (fixed * newSum / st.OverallSum)
	if ar > 1 {
		return 1 / ar
	}
	return ar
}

// bestAverageAspectRatio scores a chunk by the average aspect ratio of its
// items (paper, Listing 5). Combined with stripPhrase it produces the strip
// layout.
func bestAverageAspectRatio(c *Chunk, itemSize float64, st *State) float64 {
	expanding, fixed := chunkAspect(c, st)
	newSum := c.Sum + itemSize
	if newSum == 0 || st.OverallSum == 0 || fixed == 0 {
		return 0
	}
	count := float64(len(c.Items) + 1)
	ar := expanding / count / (fixed * newSum / st.OverallSum)
	if ar > 1 {
		return 1 / ar
	}
	return ar
}

// pivotByMiddle is a parabola peaking when the current chunk holds half of the
// items, so the chunk closes around the middle of the input. With a recurse
// functor this splits the items into nested halves, a pivot layout (paper,
// Listing 7, PivotByMiddle).
func pivotByMiddle(c *Chunk, itemSize float64, st *State) float64 {
	half := float64(st.ItemCount) / 2
	d := float64(len(c.Items)+1) - half
	return -(d * d)
}

// --- Recurse functors -----------------------------------------------------

// recurseManyItems re-enters the layout inside a chunk whenever it holds more
// than two items, the stop criterion used for pivot layouts.
func recurseManyItems(c *Chunk, st *State) bool {
	return len(c.Items) > 2
}

// --- Phrase functors ------------------------------------------------------

// byDepthParity alternates the stacking axis with depth: vertical stacks at
// even depth (full width rows split horizontally), horizontal stacks at odd
// depth (full height columns split vertically). This is slice & dice.
func byDepthParity(prev *Chunk, st *State) Config {
	if st.Depth%2 == 0 {
		return Config{Side: West, Dir: Down}
	}
	return Config{Side: North, Dir: Right}
}

// stripPhrase places every chunk as a full-width row along the top, so chunks
// stack downward, the strip layout.
func stripPhrase(prev *Chunk, st *State) Config {
	return Config{Side: North, Dir: Right}
}

// keepSquare places each chunk against the longer side of the remaining space,
// keeping the leftover area as square as possible.
func keepSquare(prev *Chunk, st *State) Config {
	if st.Available.X.Length() >= st.Available.Y.Length() {
		return Config{Side: West, Dir: Down}
	}
	return Config{Side: North, Dir: Right}
}
