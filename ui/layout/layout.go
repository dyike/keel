// Package layout arranges components: Column, Row, Card, spacing, and the
// Frame drawing helper that components use for bordered boxes.
package layout

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

type (
	C = core.C
	D = core.D
)

type flex struct {
	axis     layout.Axis
	gap      unit.Dp
	children []core.Widget
}

type grow struct{ core.Widget }

// Column stacks children vertically and stretches them to full width.
func Column(children ...core.Widget) core.Widget {
	return &flex{axis: layout.Vertical, gap: 12, children: children}
}

// Row places children side by side, vertically centered.
func Row(children ...core.Widget) core.Widget {
	return &flex{axis: layout.Horizontal, gap: 8, children: children}
}

// Grow lets a child of Row or Column take the remaining space.
func Grow(w core.Widget) core.Widget { return grow{w} }

func (f *flex) Layout(gtx C) D {
	items := make([]layout.FlexChild, 0, 2*len(f.children))
	for i, c := range f.children {
		if i > 0 {
			items = append(items, layout.Rigid(layout.Spacer{Width: f.gap, Height: f.gap}.Layout))
		}
		lay := func(gtx C) D {
			if f.axis == layout.Vertical {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
			}
			return c.Layout(gtx)
		}
		if _, ok := c.(grow); ok {
			items = append(items, layout.Flexed(1, lay))
		} else {
			items = append(items, layout.Rigid(lay))
		}
	}
	fl := layout.Flex{Axis: f.axis}
	if f.axis == layout.Horizontal {
		fl.Alignment = layout.Middle
	}
	return fl.Layout(gtx, items...)
}

type card struct{ content core.Widget }

// Card groups children in a rounded panel, laid out as a Column.
func Card(children ...core.Widget) core.Widget { return card{Column(children...)} }

func (c card) Layout(gtx C) D {
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	return Frame(gtx, theme.Surface, theme.Border, 10, layout.UniformInset(18), c.content.Layout)
}

type divider struct{}

// Divider is a full-width 1dp line.
func Divider() core.Widget { return divider{} }

func (divider) Layout(gtx C) D {
	size := image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
	paint.FillShape(gtx.Ops, theme.Border, clip.Rect{Max: size}.Op())
	return D{Size: size}
}

// Space is empty room of the given size in dp.
func Space(dp int) core.Widget {
	return core.Func(func(gtx C) D { return layout.Spacer{Width: unit.Dp(dp), Height: unit.Dp(dp)}.Layout(gtx) })
}

// Frame draws a rounded, bordered background behind inset content. Widgets use
// it for fields and panels.
func Frame(gtx C, bg, border color.NRGBA, radius unit.Dp, in layout.Inset, w layout.Widget) D {
	return layout.Background{}.Layout(gtx,
		func(gtx C) D {
			r := gtx.Dp(radius)
			size := gtx.Constraints.Min
			paint.FillShape(gtx.Ops, border, clip.UniformRRect(image.Rectangle{Max: size}, r).Op(gtx.Ops))
			b := gtx.Dp(1)
			inner := clip.UniformRRect(image.Rect(b, b, size.X-b, size.Y-b), max(r-b, 0))
			paint.FillShape(gtx.Ops, bg, inner.Op(gtx.Ops))
			return D{Size: size}
		},
		func(gtx C) D { return in.Layout(gtx, w) },
	)
}
