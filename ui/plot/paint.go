package plot

import (
	"image"
	"image/color"
	"math"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/layout"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/third_party/gio/widget/material"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/theme"
)

// Canvas clips every primitive to Bounds and translates local pixels to its
// origin. Construct it during Layout; do not retain it across frames.
type Canvas struct {
	Context core.C
	Bounds  image.Rectangle
}

func (c Canvas) scope() func() {
	transform := op.Offset(c.Bounds.Min).Push(c.Context.Ops)
	cut := clip.Rect{Max: c.Bounds.Size()}.Push(c.Context.Ops)
	return func() { cut.Pop(); transform.Pop() }
}

// Geometry outside +/- 10 million pixels is rejected before reaching Gio.
func coordinate(x float32) bool { return finite(float64(x)) && math.Abs(float64(x)) <= 1e7 }
func point(p f32.Point) bool    { return coordinate(p.X) && coordinate(p.Y) }
func path(gtx core.C, points []f32.Point, closed bool) clip.PathSpec {
	var p clip.Path
	p.Begin(gtx.Ops)
	p.MoveTo(points[0])
	for _, q := range points[1:] {
		p.LineTo(q)
	}
	if closed {
		p.Close()
	}
	return p.End()
}
func stroke(gtx core.C, points []f32.Point, width float32, c color.NRGBA) {
	if len(points) > 1 && width > 0 && coordinate(width) {
		paint.FillShape(gtx.Ops, c, clip.Stroke{Path: path(gtx, points, false), Width: width}.Op())
	}
}

// Bar paints a rectangle, including horizontal or negative bars supplied by
// the caller. Reversed rectangle corners are normalized. Radius is in pixels.
type Rectangle struct{ Min, Max f32.Point }

type Bar struct {
	Rect   Rectangle
	Fill   color.NRGBA
	Radius float32
}

func (b Bar) Paint(c Canvas) {
	if !point(b.Rect.Min) || !point(b.Rect.Max) || !coordinate(b.Radius) {
		return
	}
	r := b.Rect
	lo := f32.Pt(min(r.Min.X, r.Max.X), min(r.Min.Y, r.Max.Y))
	hi := f32.Pt(max(r.Min.X, r.Max.X), max(r.Min.Y, r.Max.Y))
	if lo == hi {
		return
	}
	defer c.scope()()
	rpx := image.Rect(int(math.Floor(float64(lo.X))), int(math.Floor(float64(lo.Y))), int(math.Ceil(float64(hi.X))), int(math.Ceil(float64(hi.Y))))
	radius := int(max(0, min(b.Radius, min(hi.X-lo.X, hi.Y-lo.Y)/2)))
	paint.FillShape(c.Context.Ops, b.Fill, clip.UniformRRect(rpx, radius).Op(c.Context.Ops))
}

type Curve uint8

const (
	CurveLinear Curve = iota
	CurveStepAfter
	CurveSmooth
)

// Line splits at invalid points; dots use the original samples. Smooth uses a
// per-segment smoothstep interpolation and preserves both sample endpoints.
type Line struct {
	Points           []f32.Point
	Stroke           color.NRGBA
	Width, DotRadius float32
	Curve            Curve
}

func curvePoints(in []f32.Point, curve Curve) []f32.Point {
	if len(in) < 2 || curve == CurveLinear {
		return in
	}
	out := []f32.Point{in[0]}
	for i := 1; i < len(in); i++ {
		a, b := in[i-1], in[i]
		if curve == CurveStepAfter {
			out = append(out, f32.Pt(b.X, a.Y), b)
		} else if curve == CurveSmooth {
			for j := 1; j <= 12; j++ {
				t := float32(j) / 12
				out = append(out, f32.Pt(a.X+(b.X-a.X)*t, a.Y+(b.Y-a.Y)*t*t*(3-2*t)))
			}
		} else {
			out = append(out, b)
		}
	}
	return out
}
func (l Line) Paint(c Canvas) {
	defer c.scope()()
	var run []f32.Point
	flush := func() { stroke(c.Context, curvePoints(run, l.Curve), l.Width, l.Stroke); run = run[:0] }
	for _, p := range l.Points {
		if !point(p) {
			flush()
			continue
		}
		run = append(run, p)
	}
	flush()
	if l.DotRadius > 0 && coordinate(l.DotRadius) {
		for _, p := range l.Points {
			if point(p) {
				drawDot(c.Context, p, l.DotRadius, l.Stroke)
			}
		}
	}
}

// Area fills between matching upper and lower boundaries. Invalid pairs split
// the polygon. Both slices must have equal length; neither is retained.
type Area struct {
	Upper, Lower []f32.Point
	Fill, Stroke color.NRGBA
	Width        float32
	Curve        Curve
}

func (a Area) Paint(c Canvas) {
	if len(a.Upper) != len(a.Lower) {
		return
	}
	defer c.scope()()
	var upper, lower []f32.Point
	flush := func() {
		if len(upper) >= 2 {
			u := curvePoints(upper, a.Curve)
			l := curvePoints(lower, a.Curve)
			poly := append([]f32.Point(nil), u...)
			for i := len(l) - 1; i >= 0; i-- {
				poly = append(poly, l[i])
			}
			paint.FillShape(c.Context.Ops, a.Fill, clip.Outline{Path: path(c.Context, poly, true)}.Op())
			stroke(c.Context, u, a.Width, a.Stroke)
		}
		upper, lower = upper[:0], lower[:0]
	}
	for i, p := range a.Upper {
		if !point(p) || !point(a.Lower[i]) {
			flush()
			continue
		}
		upper = append(upper, p)
		lower = append(lower, a.Lower[i])
	}
	flush()
}

// Arc paints a pie or donut sector. Radii are nonnegative pixels and inner must
// not exceed outer. The sampled outline includes exact angular endpoints.
type Arc struct {
	Center       f32.Point
	Inner, Outer float32
	Start, End   float64
	Fill         color.NRGBA
}

func (a Arc) Paint(c Canvas) {
	if !point(a.Center) || !coordinate(a.Inner) || !coordinate(a.Outer) || a.Inner < 0 || a.Outer <= a.Inner || !finite(a.Start) || !finite(a.End) {
		return
	}
	sweep := a.End - a.Start
	if !finite(sweep) || sweep == 0 || math.Abs(sweep) > 2*math.Pi+1e-12 {
		return
	}
	n := min(2048, max(2, int(math.Ceil(math.Abs(sweep)*math.Sqrt(float64(a.Outer))*2))))
	points := make([]f32.Point, 0, 2*(n+1))
	at := func(radius float32, angle float64) f32.Point {
		return a.Center.Add(f32.Pt(radius*float32(math.Cos(angle)), radius*float32(math.Sin(angle))))
	}
	start := math.Mod(a.Start, 2*math.Pi)
	for i := 0; i <= n; i++ {
		points = append(points, at(a.Outer, start+sweep*float64(i)/float64(n)))
	}
	if a.Inner == 0 {
		points = append(points, a.Center)
	} else {
		for i := n; i >= 0; i-- {
			points = append(points, at(a.Inner, start+sweep*float64(i)/float64(n)))
		}
	}
	defer c.scope()()
	paint.FillShape(c.Context.Ops, a.Fill, clip.Outline{Path: path(c.Context, points, true)}.Op())
}
func drawDot(gtx core.C, p f32.Point, r float32, c color.NRGBA) {
	rect := image.Rect(int(p.X-r), int(p.Y-r), int(p.X+r+.5), int(p.Y+r+.5))
	paint.FillShape(gtx.Ops, c, clip.Ellipse(rect).Op(gtx.Ops))
}

type Dot struct {
	At     f32.Point
	Radius float32
	Fill   color.NRGBA
}

func (d Dot) Paint(c Canvas) {
	if point(d.At) && d.Radius > 0 && coordinate(d.Radius) {
		defer c.scope()()
		drawDot(c.Context, d.At, d.Radius, d.Fill)
	}
}

// CrossLine draws optional horizontal/vertical guides through At.
type CrossLine struct {
	At                   f32.Point
	Horizontal, Vertical bool
	Stroke               color.NRGBA
	Width                float32
}

func (x CrossLine) Paint(c Canvas) {
	if !point(x.At) {
		return
	}
	defer c.scope()()
	size := c.Bounds.Size()
	if x.Horizontal {
		stroke(c.Context, []f32.Point{f32.Pt(0, x.At.Y), f32.Pt(float32(size.X), x.At.Y)}, x.Width, x.Stroke)
	}
	if x.Vertical {
		stroke(c.Context, []f32.Point{f32.Pt(x.At.X, 0), f32.Pt(x.At.X, float32(size.Y))}, x.Width, x.Stroke)
	}
}

type AxisSide uint8

const (
	AxisBottom AxisSide = iota
	AxisTop
	AxisLeft
	AxisRight
)

type AxisText struct {
	Position float32
	Text     string
}

// Axis paints a baseline, ticks and labels in a local canvas. At is Y for
// horizontal axes and X for vertical axes. Labels are clipped to Canvas bounds;
// reserve margins there. TextSize is sp; geometry and stroke width are pixels.
type Axis struct {
	Side                          AxisSide
	At, From, To, TickSize, Width float32
	Ticks                         []AxisText
	Stroke, TextColor             color.NRGBA
	TextSize                      unit.Sp
}

func (a Axis) Paint(c Canvas) {
	if a.Side > AxisRight || !coordinate(a.At) || !coordinate(a.From) || !coordinate(a.To) || !coordinate(a.TickSize) {
		return
	}
	defer c.scope()()
	vertical := a.Side == AxisLeft || a.Side == AxisRight
	sign := float32(1)
	if a.Side == AxisLeft || a.Side == AxisTop {
		sign = -1
	}
	pos := func(t, offset float32) f32.Point {
		if vertical {
			return f32.Pt(a.At+offset, t)
		}
		return f32.Pt(t, a.At+offset)
	}
	stroke(c.Context, []f32.Point{pos(a.From, 0), pos(a.To, 0)}, a.Width, a.Stroke)
	for _, tick := range a.Ticks {
		if !coordinate(tick.Position) {
			continue
		}
		p := pos(tick.Position, sign*a.TickSize)
		stroke(c.Context, []f32.Point{pos(tick.Position, 0), p}, a.Width, a.Stroke)
		if tick.Text == "" {
			continue
		}
		size := a.TextSize
		if size <= 0 || !finite(float64(size)) {
			size = 12
		}
		label := material.Label(theme.Material, size, tick.Text)
		label.Color = a.TextColor
		label.MaxLines = 1
		g := c.Context
		g.Constraints = layout.Constraints{Max: c.Bounds.Size()}
		record := op.Record(g.Ops)
		dim := label.Layout(g)
		call := record.Stop()
		x, y := p.X-float32(dim.Size.X)/2, p.Y+3
		if vertical {
			y = p.Y - float32(dim.Size.Y)/2
			x = p.X + 3
			if sign < 0 {
				x = p.X - 3 - float32(dim.Size.X)
			}
		} else if sign < 0 {
			y = p.Y - 3 - float32(dim.Size.Y)
		}
		trans := op.Offset(image.Pt(int(x), int(y))).Push(g.Ops)
		call.Add(g.Ops)
		trans.Pop()
	}
}
