// Package editorstyle supplies the shared input appearance for widget and el.
package editorstyle

import (
	"image"
	"image/color"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"
)

// Caret retains blink, caret and selection ink measurements for one editor.
type Caret struct {
	font                          font.Font
	pixels                        int
	shaper                        *text.Shaper
	top, bottom                   int
	caretTop, caretBottom         int // the body of 国 and A, which text centers
	focused                       bool
	start, end, length            int
	blink                         time.Time
	bounds                        image.Rectangle
	selectionText                 string
	selectionTop, selectionBottom int
	regions                       []widget.Region
}

// Layout draws an editor with ink-height selection and caret at Gio's actual
// coordinates. The caller MUST drain Editor.Update first, handling its change
// and submit events. Gio 0.10 has no separate caret paint hook: ReadOnly is set
// only during drawing to suppress its font-line-height caret, then restored.
// Input, selection, scrolling, undo and IME still belong to the original Editor.
func (c *Caret) Layout(gtx layout.Context, style material.EditorStyle, shaper *text.Shaper) layout.Dimensions {
	return c.LayoutDecorated(gtx, style, shaper, nil)
}

// LayoutDecorated adds editor-coordinate backgrounds after text layout and
// before selection and glyph painting. The decoration must not change layout.
func (c *Caret) LayoutDecorated(gtx layout.Context, style material.EditorStyle, shaper *text.Shaper, decorate func(layout.Context)) layout.Dimensions {
	ed := style.Editor
	readOnly := ed.ReadOnly
	selectionColor := style.SelectionColor
	style.SelectionColor = color.NRGBA{}
	rec := op.Record(gtx.Ops)
	var dims layout.Dimensions
	func() {
		ed.ReadOnly = true
		defer func() { ed.ReadOnly = readOnly }()
		dims = style.Layout(gtx)
	}()
	call := rec.Stop()
	pixels := gtx.Sp(style.TextSize)
	if c.shaper != shaper || c.font != style.Font || c.pixels != pixels {
		c.shaper, c.font, c.pixels = shaper, style.Font, pixels
		c.top, c.bottom = inkBounds(shaper, style.Font, pixels)
		c.caretTop, c.caretBottom = bodyBounds(shaper, style.Font, pixels)
		c.selectionText = ""
	}
	if decorate != nil {
		decorate(gtx)
	}
	// Layout first to obtain Gio's current (scrolled and wrapped) regions,
	// then replay its text above our ink-height selection backgrounds.
	if ed.SelectionLen() > 0 && (gtx.Focused(ed) || ed.PaintSelectionWhenUnfocused) {
		c.paintSelection(gtx, style, dims, selectionColor)
	} else {
		c.selectionText = ""
		c.regions = c.regions[:0]
	}
	call.Add(gtx.Ops)
	c.bounds = image.Rectangle{}
	focused := gtx.Focused(ed) && gtx.Enabled() && !readOnly
	if !focused {
		c.focused = false
		return dims
	}
	start, end := ed.Selection()
	if !c.focused || start != c.start || end != c.end || ed.Len() != c.length {
		c.blink = gtx.Now
	}
	c.focused, c.start, c.end, c.length = true, start, end, ed.Len()
	pos := ed.CaretCoords().Round()
	baseline := pos.Y
	if ed.Len() == 0 && style.Hint != "" {
		// Material paints the hint as a separate Label. Empty editor shaping
		// may use a different fallback font/baseline, so align with the hint's
		// first visible line rather than the empty buffer's caret baseline.
		if hint, ok := hintBaseline(gtx, style, shaper); ok {
			baseline = hint
		}
	}
	// The caret spans the body of 国 and A, 1dp beyond each end: centered
	// where the text's ink is centered, so it sits evenly on Chinese and on
	// Latin capitals rather than reaching down to descenders, as one spanning
	// 国 to g did (1.5dp low beside Chinese). 1.5dp wide: 1dp read as thin.
	pad := max(1, gtx.Dp(1))
	top, bottom := baseline+c.caretTop-pad, baseline+c.caretBottom+pad
	width := max(2, gtx.Dp(1.5))
	x0 := pos.X - width/2
	x1 := x0 + width
	// The line's box hugs the ink at its top, so the caret's 1dp above the
	// body would be cut off and the caret would sit low: let it reach a
	// little past the editor vertically. The field's own clip still applies.
	area := image.Rect(0, -2*pad, dims.Size.X, dims.Size.Y+2*pad)
	if ed.Len() == 0 && style.Hint != "" {
		// A caret centered on the hint's first glyph merges with its strokes
		// (a dark bar beside 搜 reads as 锼). Draw it just before the hint,
		// in the field's padding, the way native text fields do. Platform
		// text (CoreText) darkens stems past the outline by about a pixel.
		x1 = pos.X - gtx.Dp(1)
		x0 = x1 - width
		area.Min.X = x0
	}
	c.bounds = image.Rect(x0, top, x1, bottom).Intersect(area)
	// Match the usual editor cadence and settle to a visible caret after idle.
	const interval = 500 * time.Millisecond
	elapsed := max(time.Duration(0), gtx.Now.Sub(c.blink))
	if elapsed < 10*time.Second {
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(interval - elapsed%interval)})
		if elapsed/interval%2 != 0 {
			return dims
		}
	}
	if !c.bounds.Empty() {
		paint.FillShape(gtx.Ops, style.Color, clip.Rect(c.bounds).Op())
	}
	return dims
}

func (c *Caret) paintSelection(gtx layout.Context, style material.EditorStyle, dims layout.Dimensions, col color.NRGBA) {
	if !gtx.Enabled() {
		// Match material.Editor's disabled selection palette (Gio f32color.Disabled).
		lum := (13933*int(col.R) + 46871*int(col.G) + 4732*int(col.B)) / 65536
		mix := func(v uint8) uint8 { return uint8((int(v)*80 + lum*176) / 256) }
		col = color.NRGBA{R: mix(col.R), G: mix(col.G), B: mix(col.B), A: uint8(uint32(col.A) * 160 / 255)}
	}
	ed := style.Editor
	selected := ed.SelectedText()
	if ed.Mask != 0 {
		selected = strings.Map(func(r rune) rune {
			if r == '\n' {
				return r
			}
			return ed.Mask
		}, selected)
	}
	if selected != c.selectionText {
		c.selectionText = selected
		c.selectionTop, c.selectionBottom = c.top, c.bottom
		c.shaper.LayoutString(text.Parameters{Font: c.font, PxPerEm: fixed.I(c.pixels), MaxWidth: 1 << 20}, selected)
		found := false
		for g, ok := c.shaper.NextGlyph(); ok; g, ok = c.shaper.NextGlyph() {
			if g.Bounds.Min.Y == g.Bounds.Max.Y {
				continue
			}
			top, bottom := g.Bounds.Min.Y.Floor(), g.Bounds.Max.Y.Ceil()
			if !found {
				c.selectionTop, c.selectionBottom, found = top, bottom, true
			} else {
				c.selectionTop, c.selectionBottom = min(c.selectionTop, top), max(c.selectionBottom, bottom)
			}
		}
	}
	start, end := ed.Selection()
	c.regions = ed.Regions(start, end, c.regions[:0])
	pad := gtx.Dp(1)
	for _, region := range c.regions {
		r := region.Bounds
		baseline := r.Max.Y - region.Baseline
		r.Min.Y, r.Max.Y = baseline+c.selectionTop-pad, baseline+c.selectionBottom+pad
		r = r.Intersect(image.Rectangle{Max: dims.Size})
		if !r.Empty() {
			paint.FillShape(gtx.Ops, col, clip.Rect(r).Op())
		}
	}
}

// hintBaseline is the baseline of the hint's first line, where Material
// draws it.
func hintBaseline(gtx layout.Context, style material.EditorStyle, shaper *text.Shaper) (int, bool) {
	maxLines := 0
	if style.Editor.SingleLine {
		maxLines = 1
	}
	shaper.LayoutString(text.Parameters{
		Font: style.Font, PxPerEm: fixed.I(gtx.Sp(style.TextSize)),
		MaxWidth: gtx.Constraints.Max.X, MinWidth: gtx.Constraints.Min.X,
		MaxLines: maxLines, Alignment: style.Editor.Alignment, Locale: gtx.Locale,
		LineHeight: fixed.I(gtx.Sp(style.LineHeight)), LineHeightScale: style.LineHeightScale,
	}, style.Hint)
	g, ok := shaper.NextGlyph()
	if !ok {
		return 0, false
	}
	for _, more := shaper.NextGlyph(); more; _, more = shaper.NextGlyph() {
	}
	return int(g.Y), true
}

// bodyBounds is the ink of 国 and A around the baseline: the body that
// text centers (el's textShift), and the caret with it.
func bodyBounds(shaper *text.Shaper, face font.Font, pixels int) (top, bottom int) {
	shaper.LayoutString(text.Parameters{Font: face, PxPerEm: fixed.I(pixels), MaxWidth: 1 << 20}, "国Ag")
	found := false
	for i := 0; ; i++ {
		g, ok := shaper.NextGlyph()
		if !ok {
			break
		}
		if i >= 2 || g.Bounds.Min.Y == g.Bounds.Max.Y {
			continue
		}
		t, b := g.Bounds.Min.Y.Floor(), g.Bounds.Max.Y.Ceil()
		if !found {
			top, bottom, found = t, b, true
		} else {
			top, bottom = min(top, t), max(bottom, b)
		}
	}
	if !found {
		return -pixels * 7 / 10, 0
	}
	return top, bottom
}

// Use representative capitals, descenders and CJK ink so the height is stable
// across keystrokes, empty fields and password masks. Logical line descent can
// include substantial font leading; it must not extend the visual caret.
func inkBounds(shaper *text.Shaper, face font.Font, pixels int) (top, bottom int) {
	shaper.LayoutString(text.Parameters{Font: face, PxPerEm: fixed.I(pixels), MaxWidth: 1 << 20}, "国Ag")
	for {
		g, ok := shaper.NextGlyph()
		if !ok {
			break
		}
		if g.Bounds.Min.Y == g.Bounds.Max.Y {
			continue
		}
		top = min(top, g.Bounds.Min.Y.Floor())
		bottom = max(bottom, g.Bounds.Max.Y.Ceil())
	}
	if bottom <= top {
		return -pixels, 0
	}
	return
}
