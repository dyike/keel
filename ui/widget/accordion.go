package widget

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/theme"
)

type accordionItem struct {
	title          string
	body           core.Widget
	open, disabled bool
	click          widget.Clickable
}

// AccordionView keeps collapsible sections. Closed bodies keep their widget
// state but are not laid out, painted, or included in keyboard focus traversal.
type AccordionView struct {
	items              []*accordionItem
	multiple, disabled bool
	onChange           func([]int)
}

func Accordion() *AccordionView { return &AccordionView{} }
func (a *AccordionView) Add(title string, body core.Widget) *AccordionView {
	a.items = append(a.items, &accordionItem{title: title, body: body})
	return a
}
func (a *AccordionView) Multiple() *AccordionView { a.multiple = true; return a }
func (a *AccordionView) OnChange(fn func(openIndices []int)) *AccordionView {
	a.onChange = fn
	return a
}
func (a *AccordionView) SetDisabled(v bool) { a.disabled = v }
func (a *AccordionView) SetItemDisabled(i int, v bool) {
	if i >= 0 && i < len(a.items) {
		a.items[i].disabled = v
	}
}
func (a *AccordionView) IsOpen(i int) bool { return i >= 0 && i < len(a.items) && a.items[i].open }

// OpenIndices returns a fresh slice in display order.
func (a *AccordionView) OpenIndices() []int {
	var open []int
	for i, v := range a.items {
		if v.open {
			open = append(open, i)
		}
	}
	return open
}

// SetOpen does not notify OnChange. In single mode opening one closes others.
func (a *AccordionView) SetOpen(i int, open bool) {
	if i < 0 || i >= len(a.items) {
		return
	}
	if open && !a.multiple {
		for _, v := range a.items {
			v.open = false
		}
	}
	a.items[i].open = open
}
func (a *AccordionView) Layout(gtx C) D {
	for i, v := range a.items {
		g := gtx
		if a.disabled || v.disabled {
			g = g.Disabled()
		}
		for v.click.Clicked(g) {
			if a.disabled || v.disabled {
				continue
			}
			a.SetOpen(i, !v.open)
			gtx.Execute(key.FocusCmd{Tag: &v.click})
			open := a.OpenIndices()
			core.Call(gtx, func() {
				if a.onChange != nil {
					a.onChange(open)
				}
			})
		}
		for {
			ev, ok := g.Event(key.Filter{Focus: &v.click, Name: key.NameDownArrow}, key.Filter{Focus: &v.click, Name: key.NameUpArrow}, key.Filter{Focus: &v.click, Name: key.NameHome}, key.Filter{Focus: &v.click, Name: key.NameEnd})
			if !ok {
				break
			}
			e, ok := ev.(key.Event)
			if !ok || e.State != key.Press || a.disabled || v.disabled {
				continue
			}
			target, step := i, 1
			switch e.Name {
			case key.NameUpArrow:
				step = -1
			case key.NameHome:
				target = -1
			case key.NameEnd:
				target = len(a.items)
				step = -1
			}
			for n := 0; n < len(a.items); n++ {
				target = (target + step + len(a.items)) % len(a.items)
				if !a.items[target].disabled {
					gtx.Execute(key.FocusCmd{Tag: &a.items[target].click})
					break
				}
			}
		}
	}
	return core.Semantic(gtx, func(gtx C) D {
		children := make([]giolayout.FlexChild, 0, len(a.items)*3)
		for _, v := range a.items {
			children = append(children, giolayout.Rigid(func(gtx C) D { return a.header(gtx, v) }))
			if v.open && v.body != nil {
				children = append(children, giolayout.Rigid(func(gtx C) D {
					return giolayout.Inset{Top: 8, Bottom: 8, Left: 12, Right: 12}.Layout(gtx, v.body.Layout)
				}))
			}
			children = append(children, giolayout.Rigid(layout.Divider().Layout))
		}
		return giolayout.Flex{Axis: giolayout.Vertical}.Layout(gtx, children...)
	}, core.Role("accordion"))
}
func (a *AccordionView) header(gtx C, v *accordionItem) D {
	disabled := a.disabled || v.disabled
	if disabled {
		gtx = gtx.Disabled()
	}
	state := "collapsed"
	if v.open {
		state = "expanded"
	}
	return core.Semantic(gtx, func(gtx C) D {
		return v.click.Layout(gtx, func(gtx C) D {
			if !disabled {
				pointer.CursorPointer.Add(gtx.Ops)
			}
			bg := theme.Surface
			var border color.NRGBA
			if v.click.Hovered() && !disabled {
				bg = theme.SubtleHover
			}
			if gtx.Focused(&v.click) && !disabled {
				border = theme.Primary
			}
			return layout.Frame(gtx, bg, border, 4, giolayout.Inset{Top: 10, Bottom: 10, Left: 12, Right: 12}, func(gtx C) D {
				return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
					giolayout.Flexed(1, func(gtx C) D {
						lb := material.Label(theme.Material, theme.BodySize, v.title)
						lb.Color = theme.Text
						if disabled {
							lb.Color = theme.Muted
						}
						return layoutLabel(gtx, lb)
					}),
					giolayout.Rigid(func(gtx C) D {
						size := image.Pt(gtx.Dp(24), gtx.Dp(18))
						x, y := float32(size.X)/2, float32(size.Y)/2
						d := float32(gtx.Dp(3))
						var p clip.Path
						p.Begin(gtx.Ops)
						if v.open {
							p.MoveTo(f32.Pt(x-d, y-d/2))
							p.LineTo(f32.Pt(x, y+d/2))
							p.LineTo(f32.Pt(x+d, y-d/2))
						} else {
							p.MoveTo(f32.Pt(x-d/2, y-d))
							p.LineTo(f32.Pt(x+d/2, y))
							p.LineTo(f32.Pt(x-d/2, y+d))
						}
						paint.FillShape(gtx.Ops, theme.Muted, clip.Stroke{Path: p.End(), Width: float32(gtx.Dp(1))}.Op())
						return D{Size: size}
					}))
			})
		})
	}, semantic.Button, core.Role("disclosure", state), semantic.LabelOp(v.title), semantic.EnabledOp(!disabled), semantic.SelectedOp(v.open))
}
