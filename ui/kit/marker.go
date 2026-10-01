package kit

import (
	"gioui.org/f32"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"image/color"
)

type MarkerShape uint8

const (
	MarkerDot MarkerShape = iota
	MarkerSquare
	MarkerDiamond
)

// MarkerView is decoration only; callers supply their own adjacent Text.
type MarkerView struct {
	shape MarkerShape
	size  float32
	color *color.NRGBA
}

func Marker(shape MarkerShape) *MarkerView            { return &MarkerView{shape: shape, size: 8} }
func (v *MarkerView) Color(c color.NRGBA) *MarkerView { v.color = &c; return v }
func (v *MarkerView) Size(dp float32) *MarkerView {
	if dp > 0 {
		v.size = dp
	}
	return v
}
func (v *MarkerView) Render(*el.Context) el.Element {
	c := theme.Text
	if v.color != nil {
		c = *v.color
	}
	return el.Widget(core.Func(func(gtx core.C) core.D {
		s := gtx.Constraints.Min
		var shape clip.Op
		switch v.shape {
		case MarkerSquare:
			shape = clip.Rect{Max: s}.Op()
		case MarkerDiamond:
			var p clip.Path
			p.Begin(gtx.Ops)
			w, h := float32(s.X), float32(s.Y)
			p.MoveTo(f32.Pt(w/2, 0))
			p.LineTo(f32.Pt(w, h/2))
			p.LineTo(f32.Pt(w/2, h))
			p.LineTo(f32.Pt(0, h/2))
			p.Close()
			shape = clip.Outline{Path: p.End()}.Op()
		default:
			shape = clip.Ellipse{Max: s}.Op(gtx.Ops)
		}
		paint.FillShape(gtx.Ops, c, shape)
		return core.D{Size: s}
	})).Size(el.Dp(v.size))
}
