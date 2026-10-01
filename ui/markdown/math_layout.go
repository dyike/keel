package markdown

import (
	"image"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"

	"github.com/dyike/keel/ui/theme"
)

// Every box has a baseline, with positive extents above and below it.
// Children share it unless explicitly shifted for fractions, scripts or limits.
type mathBox struct {
	w, a, d int
	call    op.CallOp
}
type mathPlacement struct {
	box         mathBox
	x, baseline int
}

func mathCompose(gtx layout.Context, w, a, d int, parts []mathPlacement, strokes func()) mathBox {
	m := op.Record(gtx.Ops)
	for _, p := range parts {
		s := op.Offset(image.Pt(p.x, a+p.baseline-p.box.a)).Push(gtx.Ops)
		p.box.call.Add(gtx.Ops)
		s.Pop()
	}
	if strokes != nil {
		s := op.Offset(image.Pt(0, a)).Push(gtx.Ops)
		strokes()
		s.Pop()
	}
	return mathBox{w: w, a: a, d: d, call: m.Stop()}
}
func mathRule(gtx layout.Context, r image.Rectangle) {
	s := clip.Rect(r).Push(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	s.Pop()
}
func mathStroke(gtx layout.Context, points []f32.Point, width float32) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(points[0])
	for _, p := range points[1:] {
		path.LineTo(p)
	}
	s := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	s.Pop()
}
func layoutMath(gtx layout.Context, shaper *text.Shaper, n *mathExpr, rn run, display bool) mathBox {
	em := gtx.Sp(rn.size)
	gap := max(2, em/8)
	axis := em / 4
	small := rn
	small.size *= 0.72
	child := func(n *mathExpr) mathBox { return layoutMath(gtx, shaper, n, rn, display) }
	script := func(n *mathExpr) mathBox { return layoutMath(gtx, shaper, n, small, false) }
	switch n.kind {
	case "row":
		var parts []mathPlacement
		w, a, d := 0, 0, 0
		for _, c := range n.children {
			b := child(c)
			parts = append(parts, mathPlacement{box: b, x: w})
			w += b.w
			a = max(a, b.a)
			d = max(d, b.d)
		}
		return mathCompose(gtx, w, a, d, parts, nil)
	case "space":
		w := em / 4
		if n.value == "quad" {
			w = em
		}
		if n.value == "qquad" {
			w = 2 * em
		}
		return mathCompose(gtx, w, 0, 0, nil, nil)
	case "fraction":
		fractionSize := rn
		if !display {
			fractionSize.size *= 0.85
		}
		num := layoutMath(gtx, shaper, n.children[0], fractionSize, false)
		den := layoutMath(gtx, shaper, n.children[1], fractionSize, false)
		w := max(num.w, den.w) + 2*gap
		thickness := max(1, em/18)
		ny := -axis - gap - num.d
		dy := -axis + thickness + gap + den.a
		return mathCompose(gtx, w, num.a-ny, dy+den.d, []mathPlacement{{num, (w - num.w) / 2, ny}, {den, (w - den.w) / 2, dy}}, func() { mathRule(gtx, image.Rect(0, -axis, w, -axis+thickness)) })
	case "sqrt":
		b := child(n.children[0])
		lead := max(8, em*2/3)
		top := -b.a - gap
		parts := []mathPlacement{{b, lead + gap, 0}}
		a := b.a + gap + max(1, em/18)
		if n.sup != nil {
			index := script(n.sup)
			parts = append(parts, mathPlacement{index, 0, -em / 2})
			a = max(a, index.a+em/2)
		}
		w := lead + gap + b.w + gap
		return mathCompose(gtx, w, a, b.d, parts, func() {
			mathStroke(gtx, []f32.Point{f32.Pt(1, float32(-axis)), f32.Pt(float32(lead)/3, float32(-axis-gap)), f32.Pt(float32(lead)/2, float32(b.d)), f32.Pt(float32(lead), float32(top)), f32.Pt(float32(w), float32(top))}, float32(max(1, em/18)))
		})
	case "scripts":
		base := child(n.children[0])
		sup, sub := mathBox{}, mathBox{}
		if n.sup != nil {
			sup = script(n.sup)
		}
		if n.sub != nil {
			sub = script(n.sub)
		}
		// Display sums and products place limits above and below the operator.
		if display && n.children[0].kind == "large" && n.children[0].value != "∫" && n.children[0].value != "∮" {
			w := max(base.w, sup.w, sub.w)
			parts := []mathPlacement{{base, (w - base.w) / 2, 0}}
			a, d := base.a, base.d
			if n.sup != nil {
				y := -base.a - gap - sup.d
				parts = append(parts, mathPlacement{sup, (w - sup.w) / 2, y})
				a = sup.a - y
			}
			if n.sub != nil {
				y := base.d + gap + sub.a
				parts = append(parts, mathPlacement{sub, (w - sub.w) / 2, y})
				d = y + sub.d
			}
			return mathCompose(gtx, w+gap, a, d, parts, nil)
		}
		x := base.w + max(1, gap/2)
		sy := -max(em*2/5, base.a-em/4)
		dy := max(em/5, base.d-em/4)
		if n.sup != nil && n.sub != nil {
			dy = max(dy, sy+sup.d+gap+sub.a)
		}
		parts := []mathPlacement{{base, 0, 0}}
		a, d := base.a, base.d
		if n.sup != nil {
			parts = append(parts, mathPlacement{sup, x, sy})
			a = max(a, sup.a-sy)
		}
		if n.sub != nil {
			parts = append(parts, mathPlacement{sub, x, dy})
			d = max(d, dy+sub.d)
		}
		return mathCompose(gtx, x+max(sup.w, sub.w), a, d, parts, nil)
	case "delimited":
		return encloseMath(gtx, child(n.children[0]), n.value, em)
	case "matrix":
		return layoutMatrix(gtx, shaper, n, rn, display)
	case "mathbf", "mathit":
		if n.kind == "mathbf" {
			rn.font.Weight = font.Bold
		} else {
			rn.font.Style = font.Italic
		}
		return layoutMath(gtx, shaper, n.children[0], rn, display)
	default:
		rn.font.Typeface = font.Typeface("STIX Two Math, STIXGeneral, Cambria Math, Times New Roman, Go, " + string(theme.Face))
		if n.kind == "variable" {
			rn.font.Style = font.Italic
		}
		if n.kind == "large" && display {
			rn.size *= 1.5
		}
		res := shapeLine(gtx, shaper, rn, n.value, codeWidth, false)
		pad := 0
		if n.kind == "operator" {
			pad = gap
		}
		// Font ascent/descent include leading and fallback-font padding.
		// Math geometry uses the actual ink so scripts stay next to the base.
		m := op.Record(gtx.Ops)
		off := op.Offset(image.Pt(0, res.inkAscent-res.ascent)).Push(gtx.Ops)
		res.call.Add(gtx.Ops)
		off.Pop()
		b := mathBox{res.width, res.inkAscent, res.inkDescent, m.Stop()}
		if pad > 0 {
			return mathCompose(gtx, b.w+2*pad, b.a, b.d, []mathPlacement{{b, pad, 0}}, nil)
		}
		return b
	}
}

func layoutMatrix(gtx layout.Context, shaper *text.Shaper, n *mathExpr, rn run, display bool) mathBox {
	em := gtx.Sp(rn.size)
	gap := max(3, em/4)
	cols := 0
	for _, r := range n.children {
		cols = max(cols, len(r.children))
	}
	widths := make([]int, cols)
	asc := make([]int, len(n.children))
	desc := make([]int, len(n.children))
	rows := make([][]mathBox, len(n.children))
	for i, row := range n.children {
		for j, c := range row.children {
			b := layoutMath(gtx, shaper, c, rn, display)
			rows[i] = append(rows[i], b)
			widths[j] = max(widths[j], b.w)
			asc[i] = max(asc[i], b.a)
			desc[i] = max(desc[i], b.d)
		}
	}
	h := 0
	for i := range rows {
		h += asc[i] + desc[i]
		if i > 0 {
			h += gap
		}
	}
	a := h/2 + em/4
	d := h - a
	lead := em/2 + gap
	if n.value == "matrix" || n.value == "aligned" {
		lead = 0
	}
	w := lead
	for j, cw := range widths {
		w += cw
		if j > 0 {
			w += em
		}
	}
	w += lead
	parts := []mathPlacement{}
	y := -a
	for i, row := range rows {
		x := lead
		y += asc[i]
		for j, b := range row {
			pad := (widths[j] - b.w) / 2
			if n.value == "aligned" {
				pad = 0
				if j%2 == 0 {
					pad = widths[j] - b.w
				}
			}
			if n.value == "cases" {
				pad = 0
			}
			parts = append(parts, mathPlacement{b, x + pad, y})
			x += widths[j] + em
		}
		y += desc[i] + gap
	}
	return mathCompose(gtx, w, a+gap, d+gap, parts, func() {
		if lead == 0 {
			return
		}
		top, bottom := float32(-a-gap/2), float32(d+gap/2)
		for side := 0; side < 2; side++ {
			if n.value == "cases" && side == 1 {
				continue
			}
			x := float32(lead - gap)
			sign := float32(-1)
			if side == 1 {
				x = float32(w - lead + gap)
				sign = 1
			}
			outer := x + sign*float32(em/3)
			mid := (top + bottom) / 2
			switch n.value {
			case "pmatrix":
				var p clip.Path
				p.Begin(gtx.Ops)
				p.MoveTo(f32.Pt(x, top))
				p.QuadTo(f32.Pt(outer, mid), f32.Pt(x, bottom))
				s := clip.Stroke{Path: p.End(), Width: float32(max(1, em/16))}.Op().Push(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				s.Pop()
			case "bmatrix":
				mathStroke(gtx, []f32.Point{f32.Pt(x, top), f32.Pt(outer, top), f32.Pt(outer, bottom), f32.Pt(x, bottom)}, float32(max(1, em/16)))
			case "Bmatrix", "cases":
				mathStroke(gtx, []f32.Point{f32.Pt(x, top), f32.Pt((x+outer)/2, top+float32(gap)), f32.Pt((x+outer)/2, mid-float32(gap)), f32.Pt(outer, mid), f32.Pt((x+outer)/2, mid+float32(gap)), f32.Pt((x+outer)/2, bottom-float32(gap)), f32.Pt(x, bottom)}, float32(max(1, em/16)))
			default:
				mathStroke(gtx, []f32.Point{f32.Pt(x, top), f32.Pt(x, bottom)}, float32(max(1, em/16)))
				if n.value == "Vmatrix" {
					mathStroke(gtx, []f32.Point{f32.Pt(x+sign*3, top), f32.Pt(x+sign*3, bottom)}, 1)
				}
			}
		}
	})
}
