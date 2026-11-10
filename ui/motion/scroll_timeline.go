package motion

import "github.com/dyike/keel/ui/el"

// ScrollBinding drives one timeline cycle from a ScrollX/ScrollY container in
// the same root. ID names that container. Horizontal selects ScrollX. When
// Start and End are both zero, the complete scrollable extent is used;
// otherwise they are offsets in dp, with End greater than Start. Progress is
// clamped. A missing or non-scrollable container holds the start. It has no
// timer: native scroll, content updates and resizing drive the view as usual.
// Geometry is from the last painted frame, like Context.ScrollState.
type ScrollBinding struct {
	ID         string
	Horizontal bool
	Start, End float32
}

func (b ScrollBinding) progress(cx *el.Context) float32 {
	if b.ID == "" || !finiteTimeline(b.Start) || !finiteTimeline(b.End) || (b.Start != 0 || b.End != 0) && b.End <= b.Start {
		panic("motion: invalid scroll binding")
	}
	offset, viewport, content := cx.ScrollState(b.ID)
	if b.Horizontal {
		offset, viewport, content = cx.ScrollStateX(b.ID)
	}
	extent := content - viewport
	if viewport <= 0 || extent <= 0 {
		return 0
	}
	start, end := b.Start, b.End
	if start == 0 && end == 0 {
		end = extent
	}
	return min(max((offset-start)/(end-start), 0), 1)
}
