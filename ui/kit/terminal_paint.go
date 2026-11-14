package kit

import (
	"image"
	"image/color"
	"strings"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/font"
	"github.com/dyike/keel/gio/io/event"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/io/semantic"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

func terminalColor(code uint32, fallback color.NRGBA) color.NRGBA {
	if code == 0 {
		return fallback
	}
	if code&(1<<24) != 0 {
		return color.NRGBA{R: uint8(code >> 16), G: uint8(code >> 8), B: uint8(code), A: 255}
	}
	n := int(code - 1)
	palette := [16]uint32{0x000000, 0xcd0000, 0x00cd00, 0xcdcd00, 0x0000ee, 0xcd00cd, 0x00cdcd, 0xe5e5e5, 0x7f7f7f, 0xff0000, 0x00ff00, 0xffff00, 0x5c5cff, 0xff00ff, 0x00ffff, 0xffffff}
	if n < 16 {
		return terminalColor(1<<24|palette[n], fallback)
	}
	if n < 232 {
		n -= 16
		levels := [6]uint8{0, 95, 135, 175, 215, 255}
		return color.NRGBA{R: levels[n/36], G: levels[n/6%6], B: levels[n%6], A: 255}
	}
	gray := uint8(8 + (min(n, 255)-232)*10)
	return color.NRGBA{R: gray, G: gray, B: gray, A: 255}
}
func (v *TerminalView) layout(gtx core.C) core.D {
	p := v.painter
	p.measure(gtx)
	cellW, cellH := max(1, p.metrics.space), max(1, p.metrics.lh)
	cols, rows := min(512, max(1, gtx.Constraints.Max.X/cellW)), min(256, max(1, gtx.Constraints.Max.Y/cellH))
	if cols != v.model.cols || rows != v.model.rows {
		v.model.resize(cols, rows)
		v.selected = false
	}
	cols, rows = v.Size()
	if cols != v.lastCols || rows != v.lastRows {
		v.lastCols, v.lastRows = cols, rows
		if fn := v.onResize; fn != nil {
			core.Call(gtx, func() { fn(cols, rows) })
		}
	}
	if v.wantFocus && gtx.Enabled() {
		gtx.Execute(key.FocusCmd{Tag: v})
		v.wantFocus = false
	}
	maxScroll := float32(max(0, v.lineCount()*cellH-gtx.Constraints.Max.Y))
	if v.follow {
		v.scrollY = maxScroll
	}
	v.scrollY = min(max(v.scrollY, 0), maxScroll)
	v.updateInput(gtx, maxScroll)
	v.focused = gtx.Focused(v) && gtx.Enabled()
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	defer area.Pop()
	semantic.Editor.Add(gtx.Ops)
	name := v.name
	if name == "" {
		name = locale.Current().Terminal
	}
	semantic.LabelOp(name).Add(gtx.Ops)
	first := max(0, int(v.scrollY)/cellH)
	last := min(v.lineCount(), first+rows+2)
	if !v.semanticKnown || v.semanticVersion != v.model.version || v.semanticFirst != first || v.semanticLast != last {
		var text strings.Builder
		for row := first; row < last; row++ {
			if row > first {
				text.WriteByte('\n')
			}
			text.WriteString(terminalLineText(v.line(row)))
		}
		v.semanticText = text.String()
		v.semanticVersion = v.model.version
		v.semanticFirst = first
		v.semanticLast = last
		v.semanticKnown = true
	}
	semantic.DescriptionOp(v.semanticText).Add(gtx.Ops)
	semantic.EnabledOp(gtx.Enabled()).Add(gtx.Ops)
	event.Op(gtx.Ops, v)
	key.InputHintOp{Tag: v, Hint: key.HintAny}.Add(gtx.Ops)
	pointer.CursorText.Add(gtx.Ops)
	v.scroll.Add(gtx.Ops)
	paint.FillShape(gtx.Ops, theme.CodeBg, clip.Rect{Max: gtx.Constraints.Max}.Op())
	a, z := v.anchor, v.end
	if z.less(a) {
		a, z = z, a
	}
	cursorRow := v.model.screen.y
	if v.model.normal == nil {
		cursorRow += len(v.model.history)
	}
	for row := first; row < last; row++ {
		y := row*cellH - int(v.scrollY)
		for col, cell := range v.line(row) {
			if cell.width == 0 {
				continue
			}
			x := col * cellW
			if x >= gtx.Constraints.Max.X {
				break
			}
			rect := image.Rect(x, y, x+max(1, int(cell.width))*cellW, y+cellH)
			fg, bg := terminalColor(cell.style.fg, theme.CodeText), terminalColor(cell.style.bg, theme.CodeBg)
			if cell.style.flags&terminalInverse != 0 {
				fg, bg = bg, fg
			}
			if cell.style.flags&terminalDim != 0 {
				fg.R = uint8((uint16(fg.R) + uint16(bg.R)) / 2)
				fg.G = uint8((uint16(fg.G) + uint16(bg.G)) / 2)
				fg.B = uint8((uint16(fg.B) + uint16(bg.B)) / 2)
			}
			if bg != theme.CodeBg {
				paint.FillShape(gtx.Ops, bg, clip.Rect(rect).Op())
			}
			at := codePos{row, col}
			if v.selected && !at.less(a) && at.less(z) {
				paint.FillShape(gtx.Ops, theme.Highlight, clip.Rect(rect).Op())
			}
			cursorX := v.model.screen.x
			if line := v.model.screen.lines[v.model.screen.y]; line[cursorX].width == 0 && cursorX > 0 {
				cursorX--
			}
			if v.focused && v.model.cursorVisible && row == cursorRow && col == cursorX {
				cursor := theme.Primary
				cursor.A = 100
				paint.FillShape(gtx.Ops, cursor, clip.Rect(rect).Op())
			}
			if cell.text == "" {
				continue
			}
			face := font.Font{Typeface: theme.MonoFace}
			if cell.style.flags&terminalBold != 0 {
				face.Weight = font.Bold
			}
			if cell.style.flags&terminalItalic != 0 {
				face.Style = font.Italic
			}
			bounds := clip.Rect(rect).Push(gtx.Ops)
			p.paintTextFace(gtx, face, cell.text, image.Pt(x, y+p.metrics.baseline), fg)
			if cell.style.flags&terminalUnderline != 0 {
				paint.FillShape(gtx.Ops, fg, clip.Rect(image.Rect(x, y+cellH-2, x+rect.Dx(), y+cellH-1)).Op())
			}
			if cell.style.flags&terminalStrike != 0 {
				paint.FillShape(gtx.Ops, fg, clip.Rect(image.Rect(x, y+cellH/2, x+rect.Dx(), y+cellH/2+1)).Op())
			}
			bounds.Pop()
		}
	}
	if v.focused {
		at := image.Pt(v.model.screen.x*cellW, cursorRow*cellH-int(v.scrollY)+p.metrics.baseline)
		if len(v.draft) > 0 {
			p.paintText(gtx, string(v.draft), at, theme.CodeText)
		}
		v.reportInput(gtx, at)
	}
	if maxScroll > 0 {
		h := min(gtx.Constraints.Max.Y, max(gtx.Dp(24), gtx.Constraints.Max.Y*gtx.Constraints.Max.Y/(v.lineCount()*cellH)))
		y := int(v.scrollY / maxScroll * float32(gtx.Constraints.Max.Y-h))
		c := theme.Muted
		c.A = 110
		paint.FillShape(gtx.Ops, c, clip.UniformRRect(image.Rect(gtx.Constraints.Max.X-gtx.Dp(6), y, gtx.Constraints.Max.X, y+h), gtx.Dp(3)).Op(gtx.Ops))
	}
	return core.D{Size: gtx.Constraints.Max}
}
func (v *TerminalView) reportInput(gtx core.C, at image.Point) {
	s := string(v.draft)
	if v.imeText != s {
		v.imeText = s
		gtx.Execute(key.SnippetCmd{Tag: v, Snippet: key.Snippet{Range: key.Range{End: len(v.draft)}, Text: s}})
	}
	caret := min(max(v.draftSelection.End, 0), len(v.draft))
	pos := codePos{at.Y, at.X + v.painter.textWidth(string(v.draft[:caret]))}
	if pos != v.imePos || v.imeRange != v.draftSelection {
		v.imeRange = v.draftSelection
		v.imePos = pos
		bounds := image.Rectangle{}
		if v.composition.Start >= 0 && len(v.draft) > 0 {
			bounds = image.Rect(at.X, at.Y-v.painter.metrics.ascent, at.X+v.painter.textWidth(s), at.Y+v.painter.metrics.descent)
		}
		gtx.Execute(key.SelectionCmd{Tag: v, Range: v.draftSelection, CompositionBounds: bounds, Caret: key.Caret{Pos: f32.Pt(float32(pos.col), float32(pos.line)), Ascent: float32(v.painter.metrics.ascent), Descent: float32(v.painter.metrics.descent)}})
	}
}
