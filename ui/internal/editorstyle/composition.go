package editorstyle

import (
	"image"
	"image/color"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

// InputMethodCaret uses the same ink metrics and baseline as the visible caret.
// Call after Layout, when the editor's scrolled geometry and font metrics exist.
func (c *Caret) InputMethodCaret(ed *widget.Editor) key.Caret {
	return key.Caret{Pos: ed.CaretCoords(), Ascent: float32(max(0, -c.top)), Descent: float32(max(0, c.bottom))}
}

// Composition paints the composing text's underline and returns its visible
// bounds in editor coordinates for the platform candidate window. Adapters that
// consume CompositionEvent themselves must call this after editor layout.
func Composition(g layout.Context, ed *widget.Editor, r key.Range, ink color.NRGBA) image.Rectangle {
	if r.Start < 0 || r.End < 0 || r.Start == r.End {
		return image.Rectangle{}
	}
	start, end := min(r.Start, r.End), max(r.Start, r.End)
	start, end = min(ed.Len(), start), min(ed.Len(), end)
	if start == end {
		return image.Rectangle{}
	}
	viewport := image.Rectangle{Max: g.Constraints.Max}
	thickness := max(1, g.Dp(1))
	var bounds image.Rectangle
	for _, region := range ed.Regions(start, end, nil) {
		rect := region.Bounds.Intersect(viewport)
		if rect.Empty() {
			continue
		}
		bounds = bounds.Union(rect)
		y := region.Bounds.Max.Y - max(region.Baseline/3, thickness)
		underline := image.Rect(region.Bounds.Min.X, y, region.Bounds.Max.X, y+thickness).Intersect(viewport)
		if !underline.Empty() {
			paint.FillShape(g.Ops, ink, clip.Rect(underline).Op())
		}
	}
	return bounds
}
