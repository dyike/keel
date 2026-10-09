package markdown

import (
	"strings"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
)

func encloseMath(gtx layout.Context, b mathBox, pair string, em int) mathBox {
	left, right, _ := strings.Cut(pair, "\x00")
	gap := max(2, em/8)
	side := max(7, em/2)
	lw, rw := side+gap, side+gap
	if left == "." {
		lw = 0
	}
	if right == "." {
		rw = 0
	}
	a, d := b.a+gap, b.d+gap
	return mathCompose(gtx, lw+b.w+rw, a, d, []mathPlacement{{b, lw, 0}}, func() {
		if lw > 0 {
			drawMathDelimiter(gtx, left, float32(gap), float32(-a+1), float32(d-1), float32(side-gap), float32(max(1, em/16)))
		}
		if rw > 0 {
			drawMathDelimiter(gtx, right, float32(lw+b.w+gap), float32(-a+1), float32(d-1), float32(side-gap), float32(max(1, em/16)))
		}
	})
}
func drawMathDelimiter(gtx layout.Context, kind string, x, top, bottom, w, stroke float32) {
	mid := (top + bottom) / 2
	switch kind {
	case "(", ")":
		edge, outer := x+w, x
		if kind == ")" {
			edge, outer = x, x+w
		}
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(edge, top))
		p.QuadTo(f32.Pt(outer, mid), f32.Pt(edge, bottom))
		s := clip.Stroke{Path: p.End(), Width: stroke}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	case "[", "]":
		edge, outer := x+w, x
		if kind == "]" {
			edge, outer = x, x+w
		}
		mathStroke(gtx, []f32.Point{f32.Pt(edge, top), f32.Pt(outer, top), f32.Pt(outer, bottom), f32.Pt(edge, bottom)}, stroke)
	case "{", "}":
		edge, outer := x+w, x
		if kind == "}" {
			edge, outer = x, x+w
		}
		inner := (edge + outer) / 2
		h := (bottom - top) / 8
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(f32.Pt(edge, top))
		p.QuadTo(f32.Pt(inner, top), f32.Pt(inner, top+h))
		p.LineTo(f32.Pt(inner, mid-h))
		p.QuadTo(f32.Pt(inner, mid), f32.Pt(outer, mid))
		p.QuadTo(f32.Pt(inner, mid), f32.Pt(inner, mid+h))
		p.LineTo(f32.Pt(inner, bottom-h))
		p.QuadTo(f32.Pt(inner, bottom), f32.Pt(edge, bottom))
		s := clip.Stroke{Path: p.End(), Width: stroke}.Op().Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		s.Pop()
	case "⟨", "⟩":
		edge, outer := x+w, x
		if kind == "⟩" {
			edge, outer = x, x+w
		}
		mathStroke(gtx, []f32.Point{f32.Pt(edge, top), f32.Pt(outer, mid), f32.Pt(edge, bottom)}, stroke)
	case "/":
		mathStroke(gtx, []f32.Point{f32.Pt(x+w, top), f32.Pt(x, bottom)}, stroke)
	default:
		mathStroke(gtx, []f32.Point{f32.Pt(x+w/2, top), f32.Pt(x+w/2, bottom)}, stroke)
		if kind == "‖" {
			mathStroke(gtx, []f32.Point{f32.Pt(x+w/2+2*stroke, top), f32.Pt(x+w/2+2*stroke, bottom)}, stroke)
		}
	}
}
