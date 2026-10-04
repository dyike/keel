package kit

import (
	"gioui.org/f32"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"image/color"
	"math"
)

// RadarChart compares nonnegative series on three or more radial dimensions.
// Invalid/missing values leave gaps; values over RadarMax are clipped to the ring.
func RadarChart(labels []string, series ...Series) *ChartView {
	v := LineChart(labels, series...)
	v.kind = ChartRadar
	v.height = 300
	return v
}
func (v *ChartView) radarScale() float64 {
	if v.options.radarMax > 0 {
		return v.options.radarMax
	}
	top := 0.0
	for i, s := range v.series {
		if v.hidden[i] {
			continue
		}
		for j := range v.labels {
			x := v.value(s, j)
			if finiteNumber(x) {
				top = max(top, x)
			}
		}
	}
	if top <= 0 {
		return 1
	}
	return top
}
func (v *ChartView) radarRadius(w, h float32) float32 {
	r := max(0, min(w/2-48, h/2-32))
	if v.options.radarRadius > 0 {
		r = min(r, v.options.radarRadius)
	}
	return r
}
func radarPoint(i, n int, r float32) f32.Point {
	a := float64(i)*2*math.Pi/float64(max(n, 1)) - math.Pi/2
	return f32.Pt(float32(math.Cos(a))*r, float32(math.Sin(a))*r)
}
func (v *ChartView) radar(cx *el.Context) el.Element {
	if len(v.labels) < 3 {
		return el.Div().H(el.Dp(v.height)).Center().Child(el.Text(locale.Current().NoData))
	}
	w := v.plotW
	if w <= 0 {
		w, _ = cx.ViewportSize()
		w = max(1, w-2*theme.SpaceLg)
	}
	r := v.radarRadius(w, v.height)
	id := autoID("radar", v)
	plot := el.Div().ID(id).H(el.Dp(v.height)).WFull().Child(el.Widget(core.Func(v.drawRadar)).H(el.Dp(v.height)).WFull())
	for i, label := range v.labels {
		p := radarPoint(i, len(v.labels), r+14)
		var content el.Element = el.Text(label).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(2)
		if v.options.radarLabel != nil {
			if c := v.options.radarLabel(cx, i, label); c != nil {
				content = c
			}
		}
		plot.Child(el.Div().Absolute().Left(max(0, min(w-72, w/2+p.X-36))).Top(max(0, min(v.height-20, v.height/2+p.Y-8))).W(el.Dp(72)).Items(el.Center).Child(content))
	}
	if v.hover >= 0 && v.hover < len(v.labels) && !v.disabled {
		plot.Child(el.Div().ID(id + "/point").Absolute().Left(v.pointerX).Top(v.pointerY).Size(el.Dp(1)))
		panel := v.tooltipPanel(cx).W(el.Dp(168)).Role("tooltip").Disabled(true).Shadow(theme.ElevationSm)
		cx.Overlay(id+"/tooltip", el.Anchored(id+"/point", panel).Owner(id).Placement(el.Bottom, el.Start).Offset(12))
	}
	return plot
}
func (v *ChartView) drawRadar(gtx core.C) core.D {
	size := gtx.Constraints.Max
	px := gtx.Metric.PxPerDp
	if px <= 0 {
		px = 1
	}
	w, h := float32(size.X)/px, float32(size.Y)/px
	if gtx.Enabled() && w != v.plotW {
		v.plotW = w
		gtx.Execute(op.InvalidateCmd{})
	}
	c := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	r := v.radarRadius(w, h) * px
	n := len(v.labels)
	if gtx.Enabled() {
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &v.tag, Kinds: pointer.Move | pointer.Enter | pointer.Leave})
			if !ok {
				break
			}
			e := ev.(pointer.Event)
			next := -1
			x, y := e.Position.X-c.X, e.Position.Y-c.Y
			if e.Kind != pointer.Leave && r > 0 && x*x+y*y <= r*r {
				a := math.Atan2(float64(y), float64(x)) + math.Pi/2
				if a < 0 {
					a += 2 * math.Pi
				}
				next = int(math.Round(a*float64(n)/(2*math.Pi))) % n
			}
			if next != v.hover || (next >= 0 && (v.pointerX != e.Position.X/px || v.pointerY != e.Position.Y/px)) {
				v.pointerX, v.pointerY = e.Position.X/px, e.Position.Y/px
				v.hover = next
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
	v.hoverMotion.update(gtx, n, v.hover, !v.noHoverAnimation && !v.disabled)
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	defer area.Pop()
	event.Op(gtx.Ops, &v.tag)
	levels := v.options.radarLevels
	if levels == 0 {
		levels = 4
	}
	for level := 1; level <= levels; level++ {
		pts := make([]f32.Point, n+1)
		for i := 0; i < n; i++ {
			pts[i] = radarPoint(i, n, r*float32(level)/float32(levels)).Add(c)
		}
		pts[n] = pts[0]
		strokePath(gtx, pts, px, theme.Border)
	}
	for i := 0; i < n; i++ {
		strokePath(gtx, []f32.Point{c, radarPoint(i, n, r).Add(c)}, px, theme.Border)
	}
	top := v.radarScale()
	for si, s := range v.series {
		if v.hidden[si] {
			continue
		}
		pts := make([]f32.Point, n)
		valid := make([]bool, n)
		complete := true
		for i := range pts {
			x := v.value(s, i)
			valid[i] = finiteNumber(x) && x >= 0
			complete = complete && valid[i]
			if valid[i] {
				pts[i] = radarPoint(i, n, r*float32(min(1, x/top))).Add(c)
			}
		}
		col := v.seriesColor(si)
		if complete {
			var p clip.Path
			p.Begin(gtx.Ops)
			p.MoveTo(pts[0])
			for _, pt := range pts[1:] {
				p.LineTo(pt)
			}
			p.Close()
			fill := col
			fill.A = 40
			if s := v.options.styles[si]; s.Fill != nil {
				fill = *s.Fill
			}
			paint.FillShape(gtx.Ops, fill, clip.Outline{Path: p.End()}.Op())
		}
		for i := range pts {
			j := (i + 1) % n
			if valid[i] && valid[j] {
				strokePath(gtx, []f32.Point{pts[i], pts[j]}, px*v.seriesWidth(si), col)
			}
			weight := v.hoverMotion.weight(i)
			if v.options.styles[si].Dots {
				weight = 1
			}
			if valid[i] && weight > 0 {
				dot(gtx, pts[i], 3*px, px, chartEmphasis(col, weight), chartEmphasis(theme.Surface, weight))
			}
		}
	}
	return core.D{Size: size}
}
func (v *ChartView) drawGuide(gtx core.C, a, b f32.Point, width float32, c color.NRGBA, dashed bool) {
	if !dashed {
		strokePath(gtx, []f32.Point{a, b}, max(1, width), c)
		return
	}
	dx, dy := b.X-a.X, b.Y-a.Y
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length <= 0 {
		return
	}
	for start := float32(0); start < length; start += 8 * max(1, width) {
		end := min(length, start+4*max(1, width))
		strokePath(gtx, []f32.Point{f32.Pt(a.X+dx*start/length, a.Y+dy*start/length), f32.Pt(a.X+dx*end/length, a.Y+dy*end/length)}, max(1, width), c)
	}
}
