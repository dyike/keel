package markdown

import (
	"image"
	"image/color"
	"time"
	"unicode/utf8"

	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// TextRange is a half-open UTF-8 byte range in RenderedText, not Markdown source.
type TextRange struct{ Start, End int }

// RangeHighlight paints a background under text and beneath the selection.
// Later entries paint over earlier entries where their ranges overlap.
type RangeHighlight struct {
	Range      TextRange
	Background color.NRGBA
}

type runeHighlight struct {
	start, end int
	background color.NRGBA
}

type documentRanges struct {
	revision     uint64
	text, source string
	ready        bool
	reveal       *rangeReveal
}

type rangeReveal struct {
	offset  int
	source  string
	text    string
	expires time.Time
}

// RenderedText returns the selectable text from the most recently painted
// document. It is empty before the first frame. Block separators are included;
// decorations and custom code renderers are excluded. After changing source,
// wait for the next frame before using its offsets in the range APIs.
func (d *Doc) RenderedText() string { return d.ranges.text }

func validTextRange(s string, r TextRange) bool {
	boundary := func(i int) bool { return i == len(s) || utf8.RuneStart(s[i]) }
	return r.Start >= 0 && r.End >= r.Start && r.End <= len(s) && boundary(r.Start) && boundary(r.End)
}

// SetRangeHighlights replaces all highlights, returning false without changing
// them if any range is invalid or the source has not yet been painted. Empty
// input always clears them. Changes preserve highlights wholly within the
// unchanged prefix or suffix; highlights touching replaced text are dropped.
func (d *Doc) SetRangeHighlights(items []RangeHighlight) bool {
	if len(items) == 0 {
		d.selection.highlights = nil
		return true
	}
	if !d.ranges.ready || d.ranges.source != d.src {
		return false
	}
	next := make([]runeHighlight, 0, len(items))
	for _, h := range items {
		if !validTextRange(d.ranges.text, h.Range) {
			return false
		}
		if h.Range.Start == h.Range.End {
			continue
		}
		next = append(next, runeHighlight{
			start:      utf8.RuneCountInString(d.ranges.text[:h.Range.Start]),
			end:        utf8.RuneCountInString(d.ranges.text[:h.Range.End]),
			background: h.Background,
		})
	}
	d.selection.highlights = next
	return true
}

// RevealRange requests minimum vertical scrolling to expose the line containing
// the range start in the nearest enclosing ScrollY. Empty ranges are allowed.
// True means the request was accepted, not that an ancestor can scroll. Only
// the latest request survives; source changes cancel it and it expires in 1s.
func (d *Doc) RevealRange(r TextRange) bool {
	if !d.ranges.ready || d.ranges.source != d.src || !validTextRange(d.ranges.text, r) {
		return false
	}
	d.ranges.reveal = &rangeReveal{
		offset: utf8.RuneCountInString(d.ranges.text[:r.Start]),
		text:   d.ranges.text,
		source: d.src, expires: time.Now().Add(time.Second),
	}
	return true
}

func (d *Doc) syncRanges() {
	if d.ranges.ready && d.ranges.revision == d.selection.revision {
		d.ranges.source = d.src
		return
	}
	d.ranges.revision = d.selection.revision
	next := string(d.selection.text)
	if next != d.ranges.text {
		before, after := []rune(d.ranges.text), d.selection.text
		prefix := 0
		for prefix < min(len(before), len(after)) && before[prefix] == after[prefix] {
			prefix++
		}
		suffix := 0
		for suffix < min(len(before), len(after))-prefix && before[len(before)-1-suffix] == after[len(after)-1-suffix] {
			suffix++
		}
		kept := d.selection.highlights[:0]
		for _, h := range d.selection.highlights {
			switch {
			case h.end <= prefix:
				kept = append(kept, h)
			case h.start >= len(before)-suffix:
				delta := len(after) - len(before)
				h.start += delta
				h.end += delta
				kept = append(kept, h)
			}
		}
		d.selection.highlights = kept
		d.ranges.text = next
	}
	d.ranges.source, d.ranges.ready = d.src, true
}

func (r *richText) paintRange(gtx core.C, lo, hi int, bg color.NRGBA) {
	if lo >= hi || bg.A == 0 {
		return
	}
	for _, p := range r.pieces {
		a, b := max(lo, p.start), min(hi, p.start+p.runes)
		if a >= b {
			continue
		}
		x0, x1 := p.rect.Min.X+p.xAt(a-p.start), p.rect.Min.X+p.xAt(b-p.start)
		paint.FillShape(gtx.Ops, bg, clip.Rect(image.Rect(x0, p.rect.Min.Y, x1, p.rect.Max.Y)).Op())
	}
}

func (d *Doc) revealRange(cx *el.Context, gtx core.C) {
	req := d.ranges.reveal
	if req == nil {
		return
	}
	if req.source != d.src || req.text != d.ranges.text || !gtx.Now.Before(req.expires) {
		d.ranges.reveal = nil
		return
	}
	origin, viewport := cx.PaintGeometry()
	for i, part := range d.selection.parts {
		// Separators reveal the following block. End-of-document uses the last line.
		if req.offset >= part.end && i < len(d.selection.parts)-1 {
			continue
		}
		local := max(0, req.offset-part.start)
		for j, p := range part.r.rt.pieces {
			if local >= p.start+p.runes && j < len(part.r.rt.pieces)-1 {
				continue
			}
			rect := p.rect.Add(part.rect.Min).Add(origin)
			dy := 0
			if rect.Min.Y < viewport.Min.Y || rect.Dy() > viewport.Dy() {
				dy = rect.Min.Y - viewport.Min.Y
			} else if rect.Max.Y > viewport.Max.Y {
				dy = rect.Max.Y - viewport.Max.Y
			}
			cx.ScrollBy(dy)
			d.ranges.reveal = nil
			return
		}
	}
	gtx.Execute(op.InvalidateCmd{At: req.expires})
}
