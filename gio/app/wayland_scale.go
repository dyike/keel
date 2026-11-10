// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import (
	"image"
	"math"
)

func (w *window) renderScale() float32 {
	if w.fractionalScale != 0 {
		return float32(w.fractionalScale) / 120
	}
	return float32(max(1, w.scale))
}
func scalePoint(p image.Point, scale float32) image.Point {
	return image.Pt(int(math.Round(float64(p.X)*float64(scale))), int(math.Round(float64(p.Y)*float64(scale))))
}
func (w *window) bufferSize(p image.Point) image.Point  { return scalePoint(p, w.renderScale()) }
func (w *window) surfaceSize(p image.Point) image.Point { return scalePoint(p, 1/w.renderScale()) }

func (w *window) preferredScale(scale uint32) {
	if scale == 0 || scale == w.fractionalScale || w.viewport == nil {
		return
	}
	old := w.renderScale()
	minimum, maximum, restore := w.surfaceSize(w.config.MinSize), w.surfaceSize(w.config.MaxSize), w.surfaceSize(w.wsize)
	w.fractionalScale = scale
	ratio := w.renderScale() / old
	// Constraints and the restore size are stored in buffer pixels. Keep their
	// logical dimensions stable when moving between differently scaled outputs.
	w.config.MinSize = w.bufferSize(minimum)
	w.config.MaxSize = w.bufferSize(maximum)
	w.wsize = w.bufferSize(restore)
	w.lastPos = w.lastPos.Mul(ratio)
	w.lastTouch = w.lastTouch.Mul(ratio)
	w.scroll.dist = w.scroll.dist.Mul(ratio)
	w.resetFling()
	if w.topLvl != nil {
		w.setWindowConstraints()
	}
	w.suggestBounds()
	w.draw(true)
}

// Older compositors may omit configure_bounds. Use a conservative fraction of
// the entered output's logical size; this is not an exact work-area query.
func (w *window) suggestBounds() {
	if w.suggestedBounds != (image.Point{}) || w.w == nil || (w.viewport != nil && w.fractionalScale == 0) {
		return
	}
	for _, output := range w.disp.outputConfig {
		entered := false
		for _, other := range output.windows {
			entered = entered || other == w
		}
		if !entered || output.width <= 0 || output.height <= 0 {
			continue
		}
		size := image.Pt(output.width, output.height)
		if output.transform%2 != 0 {
			size.X, size.Y = size.Y, size.X
		}
		size = w.surfaceSize(size)
		w.suggestedBounds = image.Pt(size.X*9/10, size.Y*9/10)
		return
	}
}
