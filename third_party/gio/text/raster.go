// SPDX-License-Identifier: Unlicense OR MIT

package text

import (
	"image/color"

	"gioui.org/op"
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
	l.rasterHook = fn
}

// Raster returns the shaper's RasterHook, or nil. Keel patch.
func (l *Shaper) Raster() RasterHook { return l.rasterHook }
