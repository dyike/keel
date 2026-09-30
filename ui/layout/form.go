package layout

import (
	"image"

	"gioui.org/io/semantic"
	giolayout "gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// FormRow is one labelled field of a Form.
type FormRow struct {
	label string
	field core.Widget
}

// Field pairs a label with the widget it describes.
func Field(label string, w core.Widget) FormRow { return FormRow{label, w} }

type form struct{ rows []FormRow }

// Form lays out labelled fields with the labels in one right-aligned column,
// which is faster to scan than labels above each field. Fields that accept a
// name (widget.Input(""), widget.Select("")) take their label as the name
// agents see.
func Form(rows ...FormRow) core.Widget {
	for _, r := range rows {
		if n, ok := r.field.(interface{ SetName(string) }); ok {
			n.SetName(r.label)
		}
	}
	return &form{rows: rows}
}

func (f *form) label(gtx C, text string) D {
	lb := material.Label(theme.Material, theme.BodySize, text)
	lb.Color, lb.Alignment = theme.Muted, 2 // text.End
	return core.Semantic(gtx, lb.Layout, semantic.LabelOp(text))
}

func (f *form) Layout(gtx C) D {
	// Measure the widest label so every field starts at the same x.
	width := 0
	for _, r := range f.rows {
		m := op.Record(gtx.Ops)
		g := gtx
		g.Constraints.Min = image.Point{}
		width = max(width, f.label(g, r.label).Size.X)
		m.Stop()
	}
	children := make([]core.Widget, len(f.rows))
	for i, r := range f.rows {
		children[i] = core.Func(func(gtx C) D {
			return giolayout.Flex{Alignment: giolayout.Middle}.Layout(gtx,
				giolayout.Rigid(func(gtx C) D {
					gtx.Constraints.Min.X, gtx.Constraints.Max.X = width, width
					return f.label(gtx, r.label)
				}),
				giolayout.Rigid(giolayout.Spacer{Width: 12}.Layout),
				giolayout.Flexed(1, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					return r.field.Layout(gtx)
				}),
			)
		})
	}
	return Column(children...).Layout(gtx)
}
