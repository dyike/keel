// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"image/color"

	"github.com/dyike/keel/gio/op"
)

// RasterHook draws a run of glyphs in a solid color with a platform
// rasterizer instead of filling vector outlines. Keel patch: macOS draws
// text with CoreText this way.
//
// gs are glyphs of one line, as passed to Shape. The run's origin is the
// current transform, which callers keep an integral translation, plus fracX
// pixels to the right. It reports false when it drew nothing; the caller then
// draws the run as vector outlines.
type RasterHook = func(ops *op.Ops, gs []Glyph, col color.NRGBA, fracX float32) bool

// SetRasterHook sets the shaper's RasterHook; nil restores vector outlines.
// Keel patch.
func (l *Shaper) SetRasterHook(fn func(ops *op.Ops, gs []Glyph, col color.NRGBA, fracX float32) bool) {
	l.setRasterHook(fn, false)
}

// SetBorrowedRasterHook is SetRasterHook for a synchronous renderer that does
// not retain or modify gs. Keel's CoreText renderer uses this opt-in contract
// so labels and editors can share a bounded shaper-owned glyph buffer.
// Existing SetRasterHook callers keep their independent per-layout buffers.
func (l *Shaper) SetBorrowedRasterHook(fn RasterHook) {
	l.setRasterHook(fn, true)
}

func (l *Shaper) setRasterHook(fn RasterHook, borrows bool) {
	if l.glyphBufferEpoch == ^uint64(0) {
		panic("text: glyph buffer generation exhausted")
	}
	l.rasterHook = fn
	l.rasterBorrows = borrows
	l.glyphBuffer = nil
	l.glyphBufferEpoch++
}

// Raster returns the shaper's RasterHook, or nil. Keel patch.
func (l *Shaper) Raster() RasterHook { return l.rasterHook }

// BorrowGlyphBuffer lends widget painting scratch, not retained text layout.
// Return it with ReleaseGlyphBuffer after painting. A buffer is never shared
// by overlapping borrowers, and its capacity matches the existing 32-glyph
// paint batches. The token prevents reuse across a raster-hook change.
// Keel patch; applications normally use Label or Editor instead.
func (l *Shaper) BorrowGlyphBuffer() ([]Glyph, uint64) {
	if l.rasterHook != nil && !l.rasterBorrows {
		return make([]Glyph, 0, 32), 0
	}
	if l.glyphBufferEpoch == 0 {
		l.glyphBufferEpoch = 1
	}
	buf := l.glyphBuffer
	l.glyphBuffer = nil
	if buf == nil {
		buf = make([]Glyph, 0, 32)
	}
	return buf[:0:32], l.glyphBufferEpoch
}

func (l *Shaper) ReleaseGlyphBuffer(buf []Glyph, token uint64) {
	if token != 0 && token == l.glyphBufferEpoch && l.glyphBuffer == nil && cap(buf) == 32 && (l.rasterHook == nil || l.rasterBorrows) {
		l.glyphBuffer = buf[:0]
	}
}
