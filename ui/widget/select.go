package widget

import (
	"image"

	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

// SelectBox picks one of a list of options from a dropdown. Clicking outside
// the dropdown or pressing Esc closes it.
type SelectBox struct {
	label, hint, name string
	options           []string
	index             int
	open              bool
	onChange          func(string)

	box     widget.Clickable
	outside widget.Clickable
	opts    []widget.Clickable
}

// Select creates a dropdown; label may be empty. Nothing is chosen at first.
func Select(label string, options ...string) *SelectBox {
	s := &SelectBox{label: label, hint: "请选择", index: -1}
	s.SetOptions(options...)
	return s
}

func (s *SelectBox) Hint(h string) *SelectBox                  { s.hint = h; return s }
func (s *SelectBox) OnChange(fn func(value string)) *SelectBox { s.onChange = fn; return s }

// SetName names the dropdown for agents when it has no visible label; Form
// calls it with the field's label.
func (s *SelectBox) SetName(name string) { s.name = name }

// SetOptions replaces the options, keeping the choice if it is still present.
func (s *SelectBox) SetOptions(options ...string) {
	cur := s.Value()
	s.options = options
	s.opts = make([]widget.Clickable, len(options))
	s.index = -1
	s.SetValue(cur)
}

// Value returns the chosen option, or "".
func (s *SelectBox) Value() string {
	if s.index < 0 || s.index >= len(s.options) {
		return ""
	}
	return s.options[s.index]
}

// Index returns the chosen option's index, or -1.
func (s *SelectBox) Index() int { return s.index }

// SetValue chooses the option equal to v ("" clears) without calling OnChange.
func (s *SelectBox) SetValue(v string) {
	s.index = -1
	for i, o := range s.options {
		if o == v {
			s.index = i
		}
	}
}

func (s *SelectBox) Layout(gtx C) D {
	for s.box.Clicked(gtx) {
		s.open = !s.open
	}
	for s.outside.Clicked(gtx) {
		s.open = false
	}
	for i := range s.opts {
		for s.opts[i].Clicked(gtx) {
			s.open = false
			if i != s.index {
				s.index = i
				v := s.options[i]
				core.Call(gtx, func() {
					if s.onChange != nil {
						s.onChange(v)
					}
				})
			}
		}
	}
	if s.open {
		for {
			ev, ok := gtx.Event(key.Filter{Name: key.NameEscape})
			if !ok {
				break
			}
			if e, ok := ev.(key.Event); ok && e.State == key.Press {
				s.open = false
			}
		}
	}
	if s.label == "" {
		return s.field(gtx)
	}
	return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx,
		giolayout.Rigid(Muted(s.label).Layout),
		giolayout.Rigid(giolayout.Spacer{Height: 6}.Layout),
		giolayout.Rigid(s.field),
	)
}

func (s *SelectBox) field(gtx C) D {
	name := s.label
	if name == "" {
		name = s.name
	}
	if name == "" {
		name = s.hint
	}
	d := s.box.Layout(gtx, func(gtx C) D {
		pointer.CursorPointer.Add(gtx.Ops)
		return core.Semantic(gtx, s.closed, semantic.Button, semantic.LabelOp(name), core.Role("select", s.Value()))
	})
	if s.open {
		m := op.Record(gtx.Ops)
		s.dropdown(gtx, d.Size)
		op.Defer(gtx.Ops, m.Stop()) // drawn after, so on top of, the rest of the window
	}
	return d
}

// closed draws the field as it looks while the dropdown is shut.
func (s *SelectBox) closed(gtx C) D {
	border := theme.Border
	if s.open {
		border = theme.Primary
	}
	in := giolayout.Inset{Top: 8 + theme.CJKNudge, Bottom: 8 - theme.CJKNudge, Left: 10, Right: 10}
	return layout.Frame(gtx, theme.Surface, border, 6, in, func(gtx C) D {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		text, col := s.Value(), theme.Text
		if text == "" {
			text, col = s.hint, theme.Muted
		}
		return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
			giolayout.Flexed(1, func(gtx C) D {
				lb := material.Label(theme.Material, theme.BodySize, text)
				lb.Color, lb.MaxLines = col, 1
				return lb.Layout(gtx)
			}),
			giolayout.Rigid(func(gtx C) D {
				lb := material.Label(theme.Material, theme.BodySize, "▾")
				lb.Color = theme.Muted
				return lb.Layout(gtx)
			}),
		)
	})
}

// dropdown draws the options under a field of the given size, over a
// window-sized transparent area that closes the dropdown when clicked.
func (s *SelectBox) dropdown(gtx C, field image.Point) {
	const far = 100000
	st := op.Offset(image.Pt(-far, -far)).Push(gtx.Ops)
	catch := gtx
	catch.Constraints = giolayout.Exact(image.Pt(2*far, 2*far))
	s.outside.Layout(catch, func(gtx C) D { return D{Size: gtx.Constraints.Min} })
	st.Pop()

	defer op.Offset(image.Pt(0, field.Y+4)).Push(gtx.Ops).Pop()
	gtx.Constraints = giolayout.Constraints{Min: image.Pt(field.X, 0), Max: image.Pt(field.X, gtx.Dp(280))}
	layout.Frame(gtx, theme.Surface, theme.Border, 6, giolayout.UniformInset(4), func(gtx C) D {
		children := make([]giolayout.FlexChild, len(s.options))
		for i, o := range s.options {
			children[i] = giolayout.Rigid(func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				return s.opts[i].Layout(gtx, func(gtx C) D {
					return core.Semantic(gtx, func(gtx C) D {
						if i == s.index || s.opts[i].Hovered() {
							fillRounded(gtx, theme.Highlight, gtx.Constraints.Min.X, gtx.Dp(34), 4)
						}
						gtx.Constraints.Min.Y = gtx.Dp(34)
						return giolayout.W.Layout(gtx, func(gtx C) D {
							return giolayout.Inset{Left: 8, Right: 8, Top: theme.CJKNudge}.Layout(gtx, material.Label(theme.Material, theme.BodySize, o).Layout)
						})
					}, core.Role("option"), semantic.LabelOp(o), semantic.SelectedOp(i == s.index))
				})
			})
		}
		return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx, children...)
	})
}
