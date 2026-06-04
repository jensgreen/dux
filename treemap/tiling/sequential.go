package tiling

import (
	"math"

	"github.com/jensgreen/dux/files"
	"github.com/jensgreen/dux/geo/r1"
	"github.com/jensgreen/dux/geo/r2"
)

// Sequential is a universal tiler parametrized by the five layout functors
// (see functors.go). Concrete layouts are values of this type with particular
// functors plugged in; see layouts.go. A nil functor falls back to a sensible
// default, so the zero value lays out a single slice & dice level.
//
// Sequential implements the Tiler interface and so drops directly into the
// existing treemap construction in treemap.newR2Treemap, which handles
// descent between hierarchy levels. The Recurse functor is a separate,
// intra-level recursion used by pivot layouts and is resolved entirely within
// a single Tile() call.
type Sequential struct {
	Order   OrderFunc
	Size    SizeFunc
	Score   ScoreFunc
	Recurse RecurseFunc
	Phrase  PhraseFunc
}

func (s Sequential) Tile(rect r2.Rect, fileTree files.FileTree, depth int) (tiles []Tile, spillage r2.Rect) {
	sizeFn := s.Size
	if sizeFn == nil {
		sizeFn = FileSize
	}

	children := fileTree.Children()
	items := make([]Item, 0, len(children))
	var sum float64
	for _, child := range children {
		size := sizeFn(child)
		items = append(items, Item{File: child, Size: size})
		sum += size
	}

	st := &State{
		Depth:        depth,
		Available:    rect,
		OverallSum:   sum,
		RemainingSum: sum,
		ItemCount:    len(items),
	}
	if s.Order != nil {
		items = s.Order(items, st)
	}

	tiles = s.layout(items, st, &spillage)
	return tiles, spillage
}

// layout runs the chunking loop over items within st.Available, appending the
// resulting tiles and accumulating hidden items into spillage. It is also the
// recursion target for the Recurse functor.
func (s Sequential) layout(items []Item, st *State, spillage *r2.Rect) []Tile {
	scoreFn := s.Score
	if scoreFn == nil {
		scoreFn = oneChunk
	}
	phraseFn := s.Phrase
	if phraseFn == nil {
		phraseFn = byDepthParity
	}

	total := len(items)
	var tiles []Tile
	var prev *Chunk
	chunk := &Chunk{Config: phraseFn(nil, st)}
	prevScore := math.Inf(-1)

	flush := func() {
		if len(chunk.Items) == 0 {
			return
		}
		placeChunk(chunk, st)
		// Recurse only into a proper subset of the items; recursing into a
		// chunk that captured everything would not terminate.
		if s.Recurse != nil && len(chunk.Items) < total && s.Recurse(chunk, st) {
			sub := &State{
				Depth:        st.Depth,
				Available:    chunk.Rect,
				OverallSum:   chunk.Sum,
				RemainingSum: chunk.Sum,
				ItemCount:    len(chunk.Items),
			}
			subItems := chunk.Items
			if s.Order != nil {
				subItems = s.Order(subItems, sub)
			}
			tiles = append(tiles, s.layout(subItems, sub, spillage)...)
		} else {
			tiles = append(tiles, stackItems(chunk, spillage)...)
		}
		st.RemainingSum -= chunk.Sum
		prev = chunk
	}

	for i := range items {
		it := items[i]
		st.Index = i
		cur := scoreFn(chunk, it.Size, st)
		if cur < prevScore {
			flush()
			chunk = &Chunk{Config: phraseFn(prev, st)}
			prevScore = scoreFn(chunk, it.Size, st)
		} else {
			prevScore = cur
		}
		chunk.Items = append(chunk.Items, it)
		chunk.Sum += it.Size
	}
	flush()

	return tiles
}

// placeChunk gives the chunk a rectangle along its configured side of the
// available space, proportional to the chunk's share of the remaining size,
// and shrinks the available space accordingly (the paper's reduce step).
func placeChunk(c *Chunk, st *State) {
	f := c.Sum / st.RemainingSum
	av := st.Available
	switch c.Config.Side {
	case West:
		w := f * av.X.Length()
		c.Rect = r2.Rect{X: r1.Interval{Lo: av.X.Lo, Hi: av.X.Lo + w}, Y: av.Y}
		st.Available.X.Lo += w
	case East:
		w := f * av.X.Length()
		c.Rect = r2.Rect{X: r1.Interval{Lo: av.X.Hi - w, Hi: av.X.Hi}, Y: av.Y}
		st.Available.X.Hi -= w
	case North:
		h := f * av.Y.Length()
		c.Rect = r2.Rect{X: av.X, Y: r1.Interval{Lo: av.Y.Lo, Hi: av.Y.Lo + h}}
		st.Available.Y.Lo += h
	case South:
		h := f * av.Y.Length()
		c.Rect = r2.Rect{X: av.X, Y: r1.Interval{Lo: av.Y.Hi - h, Hi: av.Y.Hi}}
		st.Available.Y.Hi -= h
	}
}

// stackItems lays each item of a finished chunk inside the chunk's rectangle,
// stacked along the chunk's axis and sized in proportion to the item. Items
// that fall below the minimum displayable size are dropped and accumulated
// into spillage, preserving dux's small-file hiding behaviour.
func stackItems(c *Chunk, spillage *r2.Rect) []Tile {
	var tiles []Tile
	if c.Config.Side.vertical() {
		cross := c.Rect.Y.Length()
		reverse := c.Config.Dir == Up
		pos := c.Rect.Y.Lo
		if reverse {
			pos = c.Rect.Y.Hi
		}
		for _, it := range c.Items {
			dy := it.Size / c.Sum * cross
			lo, hi := pos, pos+dy
			if reverse {
				lo, hi = pos-dy, pos
			}
			if dy < MINIMUM_HEIGHT {
				// Too short to show; grow spillage from the bottom.
				spillage.Y.Lo -= dy
				continue
			}
			tiles = append(tiles, Tile{File: *it.File, Rect: r2.Rect{X: c.Rect.X, Y: r1.Interval{Lo: lo, Hi: hi}}})
			if reverse {
				pos -= dy
			} else {
				pos += dy
			}
		}
		return tiles
	}

	cross := c.Rect.X.Length()
	reverse := c.Config.Dir == Left
	pos := c.Rect.X.Lo
	if reverse {
		pos = c.Rect.X.Hi
	}
	for _, it := range c.Items {
		dx := it.Size / c.Sum * cross
		lo, hi := pos, pos+dx
		if reverse {
			lo, hi = pos-dx, pos
		}
		if dx < MINIMUM_WIDTH {
			// Too narrow to show; grow spillage from the right.
			spillage.X.Lo -= dx
			continue
		}
		tiles = append(tiles, Tile{File: *it.File, Rect: r2.Rect{X: r1.Interval{Lo: lo, Hi: hi}, Y: c.Rect.Y}})
		if reverse {
			pos -= dx
		} else {
			pos += dx
		}
	}
	return tiles
}
