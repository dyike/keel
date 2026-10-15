package el

import (
	"hash/maphash"
	"image"

	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/text"
	"github.com/dyike/keel/gio/widget/material"
	"golang.org/x/image/math/fixed"
)

type textMeasureKey struct {
	shaper *text.Shaper
	params text.Parameters
	value  string
}

// measureLabel follows Gio's Label logical bounds without building glyph
// paths, paint operations or bitmap brushes. Results live only for this
// engine's frame: text, font, DPI, locale and constraints cannot become stale.
// The bounds/visibility rules are adapted from Gio v0.10.3 widget/label.go
// (SPDX-License-Identifier: Unlicense OR MIT).
func (e *engine) measureLabel(gtx layout.Context, lb material.LabelStyle) layout.Dimensions {
	cs := gtx.Constraints
	params := text.Parameters{Font: lb.Font, PxPerEm: fixed.I(gtx.Sp(lb.TextSize)), MaxLines: lb.MaxLines, Truncator: lb.Truncator, Alignment: lb.Alignment, WrapPolicy: lb.WrapPolicy, MinWidth: cs.Min.X, MaxWidth: cs.Max.X, Locale: gtx.Locale, LineHeight: fixed.I(gtx.Sp(lb.LineHeight)), LineHeightScale: lb.LineHeightScale}
	key := textMeasureKey{lb.Shaper, params, lb.Text}
	// Height affects visibility and constrained dimensions; measurement contexts
	// in Keel are unbounded vertically, but preserve it in the cache as well.
	cacheKey := textMeasureCacheKey{key, cs.Min.Y, cs.Max.Y}
	if e.textMeasurementSeed == (maphash.Seed{}) {
		e.textMeasurementSeed = maphash.MakeSeed()
	}
	hash := maphash.Comparable(e.textMeasurementSeed, cacheKey)
	if dims, ok := e.cachedTextMeasurement(hash, cacheKey); ok {
		return dims
	}
	lb.Shaper.LayoutString(params, lb.Text)
	var bounds image.Rectangle
	first := true
	baseline, lines := 0, 0
	for g, ok := lb.Shaper.NextGlyph(); ok; g, ok = lb.Shaper.NextGlyph() {
		if lb.MaxLines > 0 {
			if g.Flags&text.FlagLineBreak != 0 {
				lines++
			}
			if lines == lb.MaxLines && g.Flags&text.FlagParagraphBreak != 0 {
				break
			}
		}
		logical := image.Rect(g.X.Floor(), int(g.Y)-g.Ascent.Ceil(), (g.X + g.Advance).Ceil(), int(g.Y)+g.Descent.Ceil())
		if first {
			first = false
			baseline = int(g.Y)
			bounds = logical
		}
		above, below := logical.Max.Y < 0, logical.Min.Y > cs.Max.Y
		left, right := logical.Max.X < 0, logical.Min.X > cs.Max.X
		if !above && !below && !left && !right {
			bounds.Min.X = min(bounds.Min.X, logical.Min.X)
			bounds.Min.Y = min(bounds.Min.Y, logical.Min.Y)
			bounds.Max.X = max(bounds.Max.X, logical.Max.X)
			bounds.Max.Y = max(bounds.Max.Y, logical.Max.Y)
		}
		if below {
			break
		}
	}
	size := cs.Constrain(bounds.Size())
	dims := layout.Dimensions{Size: size, Baseline: size.Y - baseline}
	e.cacheTextMeasurement(hash, cacheKey, dims)
	return dims
}

type textMeasureCacheKey struct {
	textMeasureKey
	minHeight, maxHeight int
}

type textMeasurement struct {
	key  textMeasureCacheKey
	dims layout.Dimensions
	next int // one-based collision chain; zero is absent
}

func (e *engine) cachedTextMeasurement(hash uint64, key textMeasureCacheKey) (layout.Dimensions, bool) {
	for i := e.textMeasurements[hash]; i != 0; {
		entry := &e.textMeasurementEntries[i-1]
		if entry.key == key {
			return entry.dims, true
		}
		i = entry.next
	}
	return layout.Dimensions{}, false
}

func (e *engine) cacheTextMeasurement(hash uint64, key textMeasureCacheKey, dims layout.Dimensions) {
	if e.textMeasurements == nil {
		e.textMeasurements = make(map[uint64]int)
	}
	e.textMeasurementEntries = append(e.textMeasurementEntries, textMeasurement{key, dims, e.textMeasurements[hash]})
	e.textMeasurements[hash] = len(e.textMeasurementEntries)
}

// The engine belongs to RootWidget and survives frames. Release text keys at
// every root layout, retaining only a modest reusable map allocation.
func (e *engine) beginTextMeasurements() {
	if len(e.textMeasurementEntries) > 4096 {
		e.textMeasurements = nil
		e.textMeasurementEntries = nil
	} else {
		clear(e.textMeasurements)
		clear(e.textMeasurementEntries)
		e.textMeasurementEntries = e.textMeasurementEntries[:0]
	}
}
