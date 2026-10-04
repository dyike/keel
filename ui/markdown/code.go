package markdown

import (
	"fmt"
	"github.com/dyike/keel/ui/locale"
	"image"
	"image/color"
	"strings"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
)

var (
	codeIcon  = mustIcon(icons.ActionCode)
	copyIcon  = mustIcon(icons.ContentContentCopy)
	wrapIcon  = mustIcon(icons.EditorWrapText)
	checkIcon = mustIcon(icons.NavigationCheck)
)

func mustIcon(data []byte) *widget.Icon {
	i, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return i
}

// codeView keeps highlighted text, wrapping, horizontal scrolling and copy
// feedback. Only the body scrolls; the language and actions stay in place.
type codeView struct {
	rich   *richBlock
	wrap   bool
	copied time.Time
	body   codeBody
}

func (d *Doc) code(cx *el.Context, b *block) el.Element {
	cv, _ := b.view.(*codeView)
	if cv == nil {
		cv = &codeView{rich: highlight(b.lang, b.code)}
		b.view = cv
	}
	context := CodeBlockContext{ID: fmt.Sprintf("markdown-code-%p", cv), Language: b.lang, Text: b.code, Wrapped: cv.wrap}
	if render := d.codeRenderers[strings.ToLower(b.lang)]; render != nil {
		if custom := render(cx, context); custom != nil {
			return el.Div().ID(context.ID).Items(el.Stretch).Child(custom)
		}
	}
	cv.body.view = cv
	cv.rich.anchor = b.anchor
	label := b.lang
	if label == "" || label == "text" || label == "txt" || label == "plaintext" {
		label = locale.Current().PlainText
	}
	text := locale.Current()
	copyLabel, copyGlyph := text.Copy, copyIcon
	if time.Since(cv.copied) < 2*time.Second {
		copyLabel, copyGlyph = text.Copied, checkIcon
	}
	wrapLabel := text.WrapLines
	if cv.wrap {
		wrapLabel = text.NoWrapLines
	}
	header := el.Div().Row().Items(el.Center).Gap(6).Px(14).Pt(8).Pb(4).Child(
		iconElement(codeIcon, 18, theme.Text),
		el.Widget(core.Func(func(gtx core.C) core.D {
			return core.Semantic(gtx, func(gtx core.C) core.D { return codeLabel(gtx, label, 13, theme.Text) }, semantic.LabelOp(label))
		})).Grow(),
		codeAction("wrap", wrapLabel, wrapIcon, cv.wrap, func() {
			cv.wrap = !cv.wrap
			cv.body.scroll.Stop()
			cv.body.barWheel.Stop()
			cv.body.scrollX = 0
		}),
		codeAction("copy", copyLabel, copyGlyph, false, func() {
			el.WriteClipboard(b.code)
			cv.copied = time.Now()
			time.AfterFunc(2*time.Second, func() { core.Update(func() {}) })
		}),
	)
	if d.codeActions != nil {
		if actions := d.codeActions(cx, context); actions != nil {
			header.Child(el.Div().ID("custom-actions").NoShrink().Child(actions))
		}
	}
	return el.Div().ID(context.ID).Role("code").Name(label).Bg(followColor(CodeBg, theme.CodeBg)).Border(1, followColor(CodeBorder, theme.Border)).Rounded(14).Child(header,
		el.Div().Px(16).Pt(6).Pb(16).Child(el.Widget(&cv.body)),
	)
}

// Center the visible ink, not the font's line box: CJK fallback fonts
// reserve different descent space than Latin fonts and square icons.
func codeLabel(gtx core.C, label string, size unit.Sp, color color.NRGBA) core.D {
	rn := run{size: size, font: font.Font{Typeface: theme.Face}}
	res := shapeLine(gtx, theme.Material.Shaper, rn, label, gtx.Constraints.Max.X, false)
	height := res.inkAscent + res.inkDescent
	dims := gtx.Constraints.Constrain(image.Pt(res.width, height))
	off := op.Offset(image.Pt(0, (dims.Y-height)/2+res.inkAscent-res.ascent)).Push(gtx.Ops)
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	res.call.Add(gtx.Ops)
	off.Pop()
	return core.D{Size: dims}
}

func iconElement(icon *widget.Icon, size float32, color color.NRGBA) el.Element {
	return el.Widget(core.Func(func(gtx core.C) core.D {
		return icon.Layout(gtx, color)
	})).Size(el.Dp(size)).NoShrink()
}

func codeAction(id, label string, icon *widget.Icon, selected bool, click func()) el.Element {
	hovered := false
	color := theme.Text
	if id == "wrap" && !selected {
		color = theme.Muted
	}
	button := el.Div().ID(id).Name(label).Size(el.Dp(32)).Rounded(8).Center().NoShrink().
		Selected(selected).CursorPointer().OnClick(click).
		Hover(func(s *el.Style) { hovered = true; s.Bg(followColor(CodeHover, theme.SubtleHover)) }).Child(iconElement(icon, 18, color))
	if selected {
		button.Bg(followColor(CodeHover, theme.SubtleHover))
	}
	return button.Decorate(func(gtx core.C, draw func()) {
		hovered = false
		draw()
		if hovered {
			codeTooltip(gtx, label)
		}
	})
}

func codeTooltip(gtx core.C, label string) {
	m := op.Record(gtx.Ops)
	tm := op.Record(gtx.Ops)
	g := gtx
	g.Constraints = layout.Constraints{Max: image.Pt(gtx.Dp(160), gtx.Dp(40))}
	dims := codeLabel(g, label, 12, theme.OnColor)
	call := tm.Stop()
	padX, padY := gtx.Dp(10), gtx.Dp(5)
	w, h := dims.Size.X+2*padX, dims.Size.Y+2*padY
	pos := image.Pt(gtx.Constraints.Max.X-w, -h-gtx.Dp(7))
	stack := op.Offset(pos).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, theme.RGB(0x20201e), clip.UniformRRect(image.Rect(0, 0, w, h), gtx.Dp(6)).Op(gtx.Ops))
	var arrow clip.Path
	arrow.Begin(gtx.Ops)
	x := float32(w - gtx.Constraints.Max.X/2)
	y, half := float32(h), float32(gtx.Dp(4))
	arrow.MoveTo(f32.Pt(x-half, y))
	arrow.LineTo(f32.Pt(x+half, y))
	arrow.LineTo(f32.Pt(x, y+half))
	arrow.Close()
	paint.FillShape(gtx.Ops, theme.RGB(0x20201e), clip.Outline{Path: arrow.End()}.Op())
	inner := op.Offset(image.Pt(padX, padY)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	inner.Pop()
	stack.Pop()
	// Reset clipping so the tooltip can sit above the rounded code card.
	op.Defer(gtx.Ops, m.Stop())
}

type codeBody struct {
	view                *codeView
	scroll              gesture.Scroll
	barWheel            gesture.Scroll
	bar                 gesture.Drag
	scrollX             int
	full, viewport      image.Point
	barStart, barScroll int
}

const codeWidth = 1 << 24

func (c *codeBody) Layout(gtx core.C) core.D {
	g := gtx
	if !c.view.wrap {
		g.Constraints = layout.Constraints{Max: image.Pt(codeWidth, codeWidth)}
	}
	m := op.Record(gtx.Ops)
	full := c.view.rich.Layout(g)
	call := m.Stop()
	width := gtx.Constraints.Constrain(image.Pt(min(full.Size.X, gtx.Constraints.Max.X), full.Size.Y)).X
	maxScroll := max(full.Size.X-width, 0)
	barH := 0
	if maxScroll > 0 && !c.view.wrap {
		barH = gtx.Dp(16)
	}
	size := gtx.Constraints.Constrain(image.Pt(width, full.Size.Y+barH))
	// Layout metadata is needed by the document's selection controller before
	// painting. Measuring must not consume input or change scroll position.
	c.full, c.viewport = full.Size, size
	if gtx.Enabled() {
		c.scrollX = min(c.scrollX, maxScroll)
		c.scrollX += c.scroll.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Horizontal,
			pointer.ScrollRange{Min: -c.scrollX, Max: maxScroll - c.scrollX}, pointer.ScrollRange{})
		c.scrollX = min(max(c.scrollX, 0), maxScroll)
		c.scrollX += c.barWheel.Update(gtx.Metric, gtx.Source, gtx.Now, gesture.Both,
			pointer.ScrollRange{Min: -c.scrollX, Max: maxScroll - c.scrollX}, pointer.ScrollRange{Min: -c.scrollX, Max: maxScroll - c.scrollX})
		c.scrollX = min(max(c.scrollX, 0), maxScroll)
		c.updateBar(gtx, maxScroll)
	}
	area := clip.Rect{Max: image.Pt(size.X, size.Y-barH)}.Push(gtx.Ops)
	if maxScroll > 0 {
		c.scroll.Add(gtx.Ops)
	}
	offset := op.Offset(image.Pt(-c.scrollX, 0)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	offset.Pop()
	area.Pop()
	if barH > 0 {
		c.paintBar(gtx, size, maxScroll, barH)
	}
	return core.D{Size: size}
}

func (c *codeBody) updateBar(gtx core.C, maxScroll int) {
	for {
		e, ok := c.bar.Update(gtx.Metric, gtx.Source, gesture.Horizontal)
		if !ok {
			break
		}
		switch e.Kind {
		case pointer.Press:
			thumb := c.thumbWidth(gtx, c.viewport.X)
			travel := c.viewport.X - thumb
			if travel > 0 {
				left := travel * c.scrollX / max(maxScroll, 1)
				x := int(e.Position.X)
				if x < left || x > left+thumb {
					c.scrollX = min(max((x-thumb/2)*maxScroll/travel, 0), maxScroll)
				}
			}
			c.barStart, c.barScroll = int(e.Position.X), c.scrollX
		case pointer.Drag:
			thumb := c.thumbWidth(gtx, c.viewport.X)
			travel := c.viewport.X - thumb
			if travel > 0 {
				c.scrollX = min(max(c.barScroll+(int(e.Position.X)-c.barStart)*maxScroll/travel, 0), maxScroll)
			}
		}
	}
}

func (c *codeBody) thumbWidth(gtx core.C, width int) int {
	return min(width, max(gtx.Dp(32), width*width/max(c.full.X, 1)))
}

func (c *codeBody) paintBar(gtx core.C, size image.Point, maxScroll, barH int) {
	thumb := c.thumbWidth(gtx, size.X)
	x := (size.X - thumb) * c.scrollX / maxScroll
	y := size.Y - barH + gtx.Dp(6)
	r := image.Rect(x, y, x+thumb, y+gtx.Dp(4))
	paint.FillShape(gtx.Ops, theme.RGB(0xa6abb3), clip.UniformRRect(r, gtx.Dp(2)).Op(gtx.Ops))
	area := clip.Rect(image.Rect(0, size.Y-barH, size.X, size.Y)).Push(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)
	c.barWheel.Add(gtx.Ops)
	c.bar.Add(gtx.Ops)
	area.Pop()
}

// selectionBounds maps the code's glyphs to document coordinates while keeping
// hit testing inside the visible text viewport, excluding the scrollbar.
func (c *codeBody) selectionBounds(bounds image.Rectangle) (image.Rectangle, image.Rectangle) {
	visible := bounds
	visible.Max.Y = min(visible.Max.Y, visible.Min.Y+c.full.Y)
	x := min(c.scrollX, max(c.full.X-c.viewport.X, 0))
	text := image.Rectangle{Min: bounds.Min.Sub(image.Pt(x, 0))}
	text.Max = text.Min.Add(image.Pt(max(c.full.X, bounds.Dx()), c.full.Y))
	return text, visible
}

var _ core.Widget = (*codeBody)(nil)

// A streaming append reparses the tail. Preserve code controls and scroll
// position while refreshing highlighted text, including code nested in lists
// or quotations.
func preserveCodeViews(old, next []block) {
	for i := range min(len(old), len(next)) {
		a, b := &old[i], &next[i]
		if a.kind != b.kind {
			continue
		}
		if a.kind == codeBlock && a.lang == b.lang {
			if cv, ok := a.view.(*codeView); ok {
				if a.code != b.code {
					cv.rich = highlight(b.lang, b.code)
					cv.copied = time.Time{}
				}
				b.view = cv
			}
		}
		preserveCodeViews(a.children, b.children)
		for j := range min(len(a.items), len(b.items)) {
			preserveCodeViews(a.items[j].blocks, b.items[j].blocks)
		}
	}
}
