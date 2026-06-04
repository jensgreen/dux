package tiling

// Layout pairs a human-readable name with a tiler. DefaultLayouts returns the
// set the UI cycles through.
type Layout struct {
	Name  string
	Tiler Tiler
}

// DefaultLayouts returns the layouts available at runtime, in cycle order,
// each wrapped with the given padding.
func DefaultLayouts(p Padding) []Layout {
	return []Layout{
		{Name: "slice & dice", Tiler: WithPadding(SliceAndDice(), p)},
		{Name: "strip", Tiler: WithPadding(Strip(), p)},
		{Name: "squarified", Tiler: WithPadding(Squarified(), p)},
		{Name: "pivot", Tiler: WithPadding(Pivot(), p)},
	}
}

// VerticalSplit lays all items in one row, dividing the space along X so each
// item spans the full height. Independent of depth.
func VerticalSplit() Sequential {
	return Sequential{
		Size:   FileSize,
		Score:  oneChunk,
		Phrase: func(*Chunk, *State) Config { return Config{Side: North, Dir: Right} },
	}
}

// HorizontalSplit lays all items in one column, dividing the space along Y so
// each item spans the full width. Independent of depth.
func HorizontalSplit() Sequential {
	return Sequential{
		Size:   FileSize,
		Score:  oneChunk,
		Phrase: func(*Chunk, *State) Config { return Config{Side: West, Dir: Down} },
	}
}

// SliceAndDice lays out all items in a single chunk whose split axis
// alternates with depth. This is the classic Johnson-Shneiderman layout.
func SliceAndDice() Sequential {
	return Sequential{
		Size:   FileSize,
		Score:  oneChunk,
		Phrase: byDepthParity,
	}
}

// Strip breaks items into full-width rows, breaking a row when adding the next
// item would worsen its average aspect ratio.
func Strip() Sequential {
	return Sequential{
		Size:   FileSize,
		Score:  bestAverageAspectRatio,
		Phrase: stripPhrase,
	}
}

// Squarified greedily groups items into chunks that keep the worst item's
// aspect ratio close to 1, placing each chunk against the longer free edge.
func Squarified() Sequential {
	return Sequential{
		Size:   FileSize,
		Score:  bestMinAspectRatio,
		Phrase: keepSquare,
	}
}

// Pivot recursively splits items into nested halves around a pivot.
func Pivot() Sequential {
	return Sequential{
		Size:    FileSize,
		Score:   pivotByMiddle,
		Recurse: recurseManyItems,
		Phrase:  keepSquare,
	}
}
