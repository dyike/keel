package widget

import (
	"strings"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// Field is an editable text box with an optional label above it.
type Field struct {
	label, hint string
	name        string // for agents when there is no label; see SetName
	onChange    func(string)
	onSubmit    func(string)
	editor      widget.Editor
	last        string // content OnChange last saw, so SetValue stays silent
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

func (f *Field) Layout(gtx C) D {
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
		giolayout.Rigid(Muted(f.label).Layout),
		giolayout.Rigid(giolayout.Spacer{Height: 6}.Layout),
		giolayout.Rigid(f.box),
	)
}

func (f *Field) box(gtx C) D {
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
	return core.Semantic(gtx, f.frame, semantic.Editor, semantic.LabelOp(name), semantic.DescriptionOp(value))
}

func (f *Field) frame(gtx C) D {
	border := theme.Border
	if gtx.Focused(&f.editor) {
		border = theme.Primary
	}
	in := giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 10, Right: 10}
	return layout.Frame(gtx, theme.Surface, border, 6, in, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		if !f.editor.SingleLine {
			gtx.Constraints.Min.Y = gtx.Dp(72)
		}
		ed := material.Editor(theme.Material, &f.editor, f.hint)
		ed.TextSize = theme.BodySize
		ed.Color = theme.Text
		ed.HintColor = theme.Muted
		return ed.Layout(gtx)
	})
}
