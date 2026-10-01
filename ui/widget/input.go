package widget

import (
	"image"
	"strings"

	"gioui.org/font"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/image/math/fixed"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/editorstyle"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// Field is an editable text box with an optional label above it.
type Field struct {
	label, hint            string
	name                   string // for agents when there is no label; see SetName
	onChange               func(string)
	onSubmit               func(string)
	editor                 widget.Editor
	caret                  editorstyle.Caret
	last                   string // content OnChange last saw, so SetValue stays silent
	disabled, clearable    bool
	prefix, suffix         core.Widget
	labelClick, clearClick widget.Clickable
	clearIcon              *IconView
	autoMin, autoMax       int
}

// Input creates a single-line field; Enter triggers OnSubmit. label may be empty.
func Input(label string) *Field {
	f := &Field{label: label}
	f.editor.SingleLine = true
	f.editor.Submit = true
	return f
}

// TextArea creates a multi-line field.
func TextArea(label string) *Field { return &Field{label: label} }

// SetName names the field for agents when it has no visible label; Form
// calls it with the field's label.
func (f *Field) SetName(name string) { f.name = name }

func (f *Field) Hint(s string) *Field            { f.hint = s; return f }
func (f *Field) Password() *Field                { f.editor.Mask = '•'; return f }
func (f *Field) MaxLength(n int) *Field          { f.editor.MaxLen = n; return f }
func (f *Field) OnChange(fn func(string)) *Field { f.onChange = fn; return f }
func (f *Field) OnSubmit(fn func(string)) *Field { f.onSubmit = fn; return f }
func (f *Field) Value() string                   { return f.editor.Text() }
func (f *Field) SetValue(s string)               { f.last = s; f.editor.SetText(s) }
func (f *Field) SetReadOnly(ro bool)             { f.editor.ReadOnly = ro }
func (f *Field) SetDisabled(v bool)              { f.disabled = v }
func (f *Field) Prefix(w core.Widget) *Field     { f.prefix = w; return f }
func (f *Field) Suffix(w core.Widget) *Field     { f.suffix = w; return f }
func (f *Field) Clearable() *Field               { f.clearable = true; return f }

// AutoHeight grows a TextArea with its content, scrolling after maxLines.
// It changes the viewport height, never truncates text. Single-line inputs
// ignore this setting. Bounds are normalized to 1 <= minLines <= maxLines.
func (f *Field) AutoHeight(minLines, maxLines int) *Field {
	f.autoMin = max(1, minLines)
	f.autoMax = max(f.autoMin, maxLines)
	return f
}

// Focus is also used by labels and custom application controls.
func (f *Field) Focus(gtx C) {
	if !f.disabled {
		gtx.Execute(key.FocusCmd{Tag: &f.editor})
	}
}

func (f *Field) Layout(gtx C) D {
	if f.disabled {
		gtx = gtx.Disabled()
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &f.labelClick, Kinds: pointer.Press})
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok && (e.Source != pointer.Mouse || e.Buttons.Contain(pointer.ButtonPrimary)) {
			f.Focus(gtx)
		}
	}
	for f.clearClick.Clicked(gtx) {
		if f.disabled || f.editor.ReadOnly || f.editor.Len() == 0 {
			continue
		}
		f.SetValue("")
		f.Focus(gtx)
		core.Call(gtx, func() {
			if f.onChange != nil {
				f.onChange("")
			}
		})
	}
	for {
		ev, ok := f.editor.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.ChangeEvent:
			// Gio reports SetText as a change too; only user edits reach OnChange.
			if text := f.editor.Text(); text != f.last {
				f.last = text
				if f.onChange != nil {
					core.Call(gtx, func() { f.onChange(text) })
				}
			}
		case widget.SubmitEvent:
			if f.onSubmit != nil {
				core.Call(gtx, func() { f.onSubmit(f.editor.Text()) })
			}
		}
	}
	if f.label == "" {
		return f.box(gtx)
	}
	return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
		giolayout.Rigid(func(gtx C) D {
			return core.Semantic(gtx, func(gtx C) D {
				// A label delegates focus to the editor; it has no separate Tab stop.
				event.Op(gtx.Ops, &f.labelClick)
				pointer.CursorPointer.Add(gtx.Ops)
				return Muted(f.label).Layout(gtx)
			})
		}),
		giolayout.Rigid(giolayout.Spacer{Height: 6}.Layout),
		giolayout.Rigid(f.box),
	)
}

func (f *Field) box(gtx C) D {
	return f.frame(gtx)
}
func (f *Field) editorBox(gtx C) D {
	name := f.label
	if name == "" {
		name = f.name
	}
	if name == "" {
		name = f.hint
	}
	value := f.editor.Text()
	if f.editor.Mask != 0 {
		value = strings.Repeat(string(f.editor.Mask), f.editor.Len())
	}
	// One textbox node carrying the label and current value; Gio's own editor
	// node inside it has neither.
	return core.Semantic(gtx, f.edit, semantic.Editor, semantic.LabelOp(name), semantic.DescriptionOp(value), semantic.EnabledOp(gtx.Enabled()))
}

func (f *Field) frame(gtx C) D {
	border := theme.Border
	if gtx.Focused(&f.editor) {
		border = theme.Primary
	}
	in := giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 10, Right: 10}
	bg := theme.Surface
	if f.disabled {
		bg = theme.Subtle
		border = theme.Border
	}
	return layout.Frame(gtx, bg, border, 6, in, func(gtx C) D {
		var children []giolayout.FlexChild
		if f.prefix != nil {
			children = append(children, giolayout.Rigid(f.prefix.Layout), giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout))
		}
		children = append(children, giolayout.Flexed(1, f.editorBox))
		if f.suffix != nil {
			children = append(children, giolayout.Rigid(giolayout.Spacer{Width: 8}.Layout), giolayout.Rigid(f.suffix.Layout))
		}
		if f.clearable && f.editor.Len() > 0 {
			children = append(children, giolayout.Rigid(giolayout.Spacer{Width: 6}.Layout), giolayout.Rigid(func(gtx C) D {
				g := gtx
				if f.editor.ReadOnly {
					g = g.Disabled()
				}
				name := f.label
				if name == "" {
					name = f.name
				}
				if name == "" {
					name = f.hint
				}
				return core.Semantic(g, func(gtx C) D {
					return f.clearClick.Layout(gtx, func(gtx C) D {
						gtx.Constraints.Min = image.Point{}
						if f.clearIcon == nil {
							f.clearIcon = Icon(IconClose).Size(16).Color(theme.Muted)
						}
						return giolayout.UniformInset(3).Layout(gtx, f.clearIcon.Layout)
					})
				}, semantic.Button, semantic.LabelOp("清空"+name), semantic.EnabledOp(g.Enabled()))
			}))
		}
		return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx, children...)
	})
}
func (f *Field) edit(gtx C) D {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	if !f.editor.SingleLine {
		gtx.Constraints.Min.Y = min(gtx.Constraints.Max.Y, gtx.Dp(72))
	}
	ed := material.Editor(theme.Material, &f.editor, f.hint)
	ed.TextSize = theme.BodySize
	ed.Color = theme.Text
	ed.HintColor = theme.Muted
	if !f.editor.SingleLine && f.autoMin > 0 {
		// Counting newlines misses soft wraps; native editor layout determines
		// content height, with measured font metrics defining the viewport bounds.
		shaper := theme.Material.Shaper
		shaper.LayoutString(text.Parameters{Font: font.Font{Typeface: theme.Material.Face}, PxPerEm: fixed.I(gtx.Sp(theme.BodySize)), MaxWidth: 1 << 20}, "国Ag")
		line := 1
		for {
			g, ok := shaper.NextGlyph()
			if !ok {
				break
			}
			line = max(line, g.Ascent.Ceil()+g.Descent.Ceil())
		}
		height := func(lines, limit int) int {
			if lines > limit/line {
				return limit
			}
			return lines * line
		}
		gtx.Constraints.Max.Y = height(f.autoMax, gtx.Constraints.Max.Y)
		gtx.Constraints.Min.Y = height(f.autoMin, gtx.Constraints.Max.Y)
		scale := gtx.Metric.PxPerSp
		if scale == 0 {
			scale = 1
		}
		ed.LineHeight = unit.Sp(float32(line) / scale)
		// This is the measured baseline spacing used by the viewport limits;
		// Gio's default 1.2 multiplier would clip the last requested line.
		ed.LineHeightScale = 1
	}
	if f.disabled {
		ed.Color = theme.Muted
	}
	return f.caret.Layout(gtx, ed, theme.Material.Shaper)
}
