// Package editorstyle supplies the shared input appearance for widget and el.
package editorstyle

import (
	"image"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"
)

// Caret retains blink and font measurements for one editor.
type Caret struct {
	font               font.Font
	pixels             int
	shaper             *text.Shaper
	top, bottom        int
	focused            bool
	start, end, length int
	blink              time.Time
	bounds             image.Rectangle
}

// Layout draws an editor with an ink-height caret at Gio's actual caret
// coordinates. The caller MUST drain Editor.Update first, handling its change
// and submit events. Gio 0.10 has no separate caret paint hook: ReadOnly is set
// only during drawing to suppress its font-line-height caret, then restored.
// Input, selection, scrolling, undo and IME still belong to the original Editor.
func (c *Caret) Layout(gtx layout.Context, style material.EditorStyle, shaper *text.Shaper) layout.Dimensions {
	ed := style.Editor
	readOnly := ed.ReadOnly
	var dims layout.Dimensions
	func() {
		ed.ReadOnly = true
		defer func() { ed.ReadOnly = readOnly }()
		dims = style.Layout(gtx)
	}()
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
	pixels := gtx.Sp(style.TextSize)
	if c.shaper != shaper || c.font != style.Font || c.pixels != pixels {
		c.shaper, c.font, c.pixels = shaper, style.Font, pixels
		c.top, c.bottom = inkBounds(shaper, style.Font, pixels)
	}
	pos := ed.CaretCoords().Round()
	top, bottom := pos.Y+c.top, pos.Y+c.bottom
	if ed.Len() == 0 && style.Hint != "" {
		// Material paints the hint as a separate Label. Empty editor shaping
		// may use a different fallback font/baseline, so align with the hint's
		// first visible line rather than the empty buffer's caret baseline.
		if hintTop, hintBottom, ok := hintInk(gtx, style, shaper); ok {
			top, bottom = hintTop, hintBottom
		}
	}
	half := max(1, gtx.Dp(1)/2)
	c.bounds = image.Rect(pos.X-half, top, pos.X+half, bottom).Intersect(image.Rectangle{Max: dims.Size})
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

func hintInk(gtx layout.Context, style material.EditorStyle, shaper *text.Shaper) (top, bottom int, found bool) {
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
	for {
		g, ok := shaper.NextGlyph()
		if !ok {
			break
		}
		if g.Bounds.Max.Y > g.Bounds.Min.Y {
			a, b := int(g.Y)+g.Bounds.Min.Y.Floor(), int(g.Y)+g.Bounds.Max.Y.Ceil()
			if !found {
				top, bottom, found = a, b, true
			} else {
				top, bottom = min(top, a), max(bottom, b)
			}
		}
		if g.Flags&text.FlagLineBreak != 0 {
			break
		}
	}
	return
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
