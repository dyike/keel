package kit

import (
	"image"
	"math"
	"slices"
	"strings"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// PlotPoint is one measurement.
type PlotPoint struct{ X, Y float64 }

// PlotSeries is a named set of points, drawn as dots, or joined by a line
// with Lines.
type PlotSeries struct {
	Name   string
	Points []PlotPoint
}

// PlotView plots numeric x/y data and lets the user explore it: the wheel
// zooms around the pointer, dragging pans, hovering picks the nearest point
// and double-clicking (or the 复位 button, or 0) restores the full view.
// Focused, + and − zoom and the arrow keys pan.
type PlotView struct {
	title    string
	series   []PlotSeries
	lines    bool
	height   float32
	format   func(float64) string
	x0, x1   float64 // visible range
	y0, y1   float64
	fitted   bool
	pick     [2]int // series, point under the pointer; -1 none
	tag      int
	drag     gesture.Drag
	click    gesture.Click
	last     f32.Point
	hasData  bool
	disabled bool
	dragging bool
	dragView [4]float64
	plotW    float32
}

func Plot(series ...PlotSeries) *PlotView {
	v := &PlotView{height: 260, format: formatNumber, pick: [2]int{-1, -1}}
	v.SetSeries(series...)
	return v
}
func (v *PlotView) Title(s string) *PlotView { v.title = s; return v }
func (v *PlotView) Lines() *PlotView         { v.lines = true; return v }
func (v *PlotView) Height(dp float32) *PlotView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}
func (v *PlotView) Format(fn func(float64) string) *PlotView {
	if fn != nil {
		v.format = fn
	}
	return v
}

// SetSeries replaces the data and fits the view to it.
func (v *PlotView) SetSeries(series ...PlotSeries) {
	v.series = slices.Clone(series)
	v.hasData = false
	for i := range v.series {
		v.series[i].Points = slices.Clone(series[i].Points)
		for _, p := range series[i].Points {
			if validPlotPoint(p) {
				v.hasData = true
				break
			}
		}
	}
	v.fitted, v.pick = false, [2]int{-1, -1}
	v.dragging = false
}
func (v *PlotView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.cancelDrag()
		v.pick = [2]int{-1, -1}
	}
}
func (v *PlotView) cancelDrag() {
	if v.dragging {
		v.x0, v.x1, v.y0, v.y1 = v.dragView[0], v.dragView[1], v.dragView[2], v.dragView[3]
	}
	v.dragging = false
	v.drag = gesture.Drag{}
}
func validPlotPoint(p PlotPoint) bool    { return finiteNumber(p.X) && finiteNumber(p.Y) }
func validPlotRange(lo, hi float64) bool { return finiteNumber(lo) && finiteNumber(hi) && hi > lo }

// View returns the visible ranges.
func (v *PlotView) View() (x0, x1, y0, y1 float64) { return v.x0, v.x1, v.y0, v.y1 }

// SetView shows the given ranges.
func (v *PlotView) SetView(x0, x1, y0, y1 float64) {
	if validPlotRange(x0, x1) && validPlotRange(y0, y1) {
		v.x0, v.x1, v.y0, v.y1, v.fitted = x0, x1, y0, y1, true
	}
}

// Reset fits the view to all the data.
func (v *PlotView) Reset() {
	first := true
	for _, s := range v.series {
		for _, p := range s.Points {
			if !validPlotPoint(p) {
				continue
			}
			if first {
				v.x0, v.x1, v.y0, v.y1, first = p.X, p.X, p.Y, p.Y, false
			}
			v.x0, v.x1 = math.Min(v.x0, p.X), math.Max(v.x1, p.X)
			v.y0, v.y1 = math.Min(v.y0, p.Y), math.Max(v.y1, p.Y)
		}
	}
	if first {
		v.x0, v.x1, v.y0, v.y1 = 0, 1, 0, 1
	}
	v.x0, v.x1 = plotPaddedRange(v.x0, v.x1)
	v.y0, v.y1 = plotPaddedRange(v.y0, v.y1)
	v.fitted = true
}

// zoom scales the view by f around (fx, fy), fractions of the plot box.
func plotPaddedRange(lo, hi float64) (float64, float64) {
	scale := max(math.Abs(lo), math.Abs(hi))
	if scale == 0 {
		return -1, 1
	}
	a, b := lo/scale, hi/scale
	d := (b - a) * .05
	if d == 0 {
		d = .05
	}
	lower, upper := (a-d)*scale, (b+d)*scale
	lower = max(-math.MaxFloat64, lower)
	upper = min(math.MaxFloat64, upper)
	if lower == upper {
		lower = math.Nextafter(lower, math.Inf(-1))
		upper = math.Nextafter(upper, math.Inf(1))
	}
	return max(-math.MaxFloat64, lower), min(math.MaxFloat64, upper)
}
func plotZoomRange(lo, hi, f, at float64) (float64, float64) {
	scale := max(math.Abs(lo), math.Abs(hi))
	if scale == 0 {
		return lo, hi
	}
	a, b := lo/scale, hi/scale
	c := a*(1-at) + b*at
	low, high := (c+(a-c)*f)*scale, (c+(b-c)*f)*scale
	if !validPlotRange(low, high) || f < 1 && (high/scale-low/scale) < 1e-12 {
		return lo, hi
	}
	return low, high
}
func (v *PlotView) zoom(f, fx, fy float64) {
	if !finiteNumber(f) || f <= 0 || !finiteNumber(fx) || !finiteNumber(fy) {
		return
	}
	f = min(100, max(.01, f))
	fx = min(1, max(0, fx))
	fy = min(1, max(0, fy))
	v.x0, v.x1 = plotZoomRange(v.x0, v.x1, f, fx)
	v.y0, v.y1 = plotZoomRange(v.y0, v.y1, f, fy)
	v.pick = [2]int{-1, -1}
}
func plotPanRange(lo, hi, d float64) (float64, float64) {
	scale := max(math.Abs(lo), math.Abs(hi))
	if scale == 0 || !finiteNumber(d) {
		return lo, hi
	}
	a, b := lo/scale, hi/scale
	low, high := (a+(b-a)*d)*scale, (b+(b-a)*d)*scale
	if !validPlotRange(low, high) {
		return lo, hi
	}
	return low, high
}
func (v *PlotView) pan(dx, dy float64) {
	v.x0, v.x1 = plotPanRange(v.x0, v.x1, dx)
	v.y0, v.y1 = plotPanRange(v.y0, v.y1, dy)
	v.pick = [2]int{-1, -1}
}

func (v *PlotView) Render(cx *el.Context) el.Element {
	if v.dragging && !cx.Enabled(autoID("plot", v)) {
		v.cancelDrag()
	}
	if !v.fitted {
		v.Reset()
	}
	text := locale.Current()
	names := make([]string, len(v.series))
	for i, s := range v.series {
		names[i] = s.Name
	}
	name := v.title
	if name == "" {
		name = strings.Join(names, ", ")
	}
	head := el.Div().Row().Wrap().Items(el.Center).Gap(12)
	if v.title != "" {
		head.Child(el.Text(v.title).Bold())
	}
	head.Child(el.Div().Grow())
	if len(v.series) > 1 {
		for i, s := range v.series {
			head.Child(el.Div().Row().Items(el.Center).Gap(6).Child(
				el.Div().Size(el.Dp(10)).Rounded(theme.RadiusFull).Bg(theme.Chart[i%len(theme.Chart)]),
				el.Text(s.Name).TextSize(theme.TextSm).TextColor(theme.Muted)))
		}
	}
	head.Child(Button(text.ResetView, v.Reset).Variant(ButtonGhost).Size(24).Render(cx))

	yt, xt := niceTicks(v.y0, v.y1, 4), niceTicks(v.x0, v.x1, 5)
	axis := el.Div().W(el.Dp(axisWidth)).H(el.Dp(v.height)).NoShrink()
	for _, t := range yt {
		if t < v.y0 || t > v.y1 {
			continue
		}
		top := float32((1-axisFraction(t, v.y0, v.y1))*float64(v.height)) - 8
		axis.Child(el.Div().Absolute().Top(top).Right(8).Child(el.Text(v.format(t)).TextSize(theme.TextXs).TextColor(theme.Muted)))
	}
	box := el.Div().Grow().W(el.Dp(0)).H(el.Dp(v.height)).Items(el.Stretch).Child(
		el.Widget(core.Func(func(gtx core.C) core.D { return v.draw(gtx, xt, yt) })).H(el.Dp(v.height)))
	if !v.hasData {
		box.Child(el.Div().Absolute().Top(20).Left(12).Child(el.Text(text.NoData).TextColor(theme.Muted)))
	}
	if s, i := v.pick[0], v.pick[1]; s >= 0 && s < len(v.series) && i >= 0 && i < len(v.series[s].Points) && v.plotW > 0 {
		p := v.series[s].Points[i]
		left := float32(axisFraction(p.X, v.x0, v.x1)) * v.plotW
		top := float32((1 - axisFraction(p.Y, v.y0, v.y1)) * float64(v.height))
		if left+12+160 > v.plotW {
			left -= 12 + 160 + 12
		}
		box.Child(el.Div().Absolute().Left(max(left+12, 0)).Top(max(top-40, 0)).W(el.Dp(min(float32(160), v.plotW))).P(8).Gap(2).Rounded(theme.RadiusMd).
			Bg(theme.Surface).Border(1, theme.Border).Child(
			el.Text(v.series[s].Name).TextSize(theme.TextSm).Bold(),
			el.Text("x "+v.format(p.X)+"   y "+v.format(p.Y)).TextSize(theme.TextSm).TextColor(theme.Muted)))
	}
	xs := el.Div().Row().H(el.Dp(18))
	previous := 0.0
	for _, t := range xt {
		if t < v.x0 || t > v.x1 {
			continue
		}
		fraction := axisFraction(t, v.x0, v.x1)
		xs.Child(el.Div().W(el.Frac(float32(fraction-previous))).NoShrink(),
			el.Div().W(el.Dp(1)).H(el.Dp(18)).Mx(-.5).NoShrink().Child(el.Div().Absolute().Left(-12).W(el.Dp(56)).Child(el.Text(v.format(t)).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1))))
		previous = fraction
	}
	return el.Div().Disabled(v.disabled).ID(autoID("plot", v)).Role("figure").Name(name).Gap(10).P(12).Rounded(theme.RadiusLg).
		Bg(theme.Surface).Border(1, theme.Border).Items(el.Stretch).Focusable(true).
		FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
		OnKey(func(e el.KeyEvent) bool {
			ok := true
			if e.State == el.KeyPress {
				switch key.Name(e.Name) {
				case "+", "=":
					v.zoom(0.8, 0.5, 0.5)
				case "-":
					v.zoom(1.25, 0.5, 0.5)
				case "0":
					v.Reset()
				case key.NameLeftArrow:
					v.pan(-0.1, 0)
				case key.NameRightArrow:
					v.pan(0.1, 0)
				case key.NameUpArrow:
					v.pan(0, 0.1)
				case key.NameDownArrow:
					v.pan(0, -0.1)
				default:
					ok = false
				}
			}
			return ok
		}).
		Child(head, el.Div().Row().Child(axis, box), el.Div().Row().Child(el.Div().W(el.Dp(axisWidth)).NoShrink(), xs.Grow().W(el.Dp(0))))
}

func (v *PlotView) draw(gtx core.C, xt, yt []float64) core.D {
	size := gtx.Constraints.Max
	px := gtx.Metric.PxPerDp
	if px <= 0 {
		px = 1
	}
	if w := float32(size.X) / px; w != v.plotW && gtx.Enabled() {
		v.plotW = w
		gtx.Execute(op.InvalidateCmd{}) // the axis labels need this width: draw again
	}
	at := func(p PlotPoint) f32.Point {
		return f32.Pt(float32(min(1e6, max(-1e6, axisFraction(p.X, v.x0, v.x1))))*float32(size.X), float32(min(1e6, max(-1e6, 1-axisFraction(p.Y, v.y0, v.y1))))*float32(size.Y))
	}
	if gtx.Enabled() {
		v.events(gtx, size, at)
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	event.Op(gtx.Ops, &v.tag)
	v.drag.Add(gtx.Ops)
	v.click.Add(gtx.Ops)
	pointer.CursorCrosshair.Add(gtx.Ops)

	line := max(1, int(px))
	for _, t := range yt {
		if y := int(at(PlotPoint{v.x0, t}).Y); y >= 0 && y <= size.Y {
			fillRect(gtx, image.Rect(0, y, size.X, y+line), theme.Border)
		}
	}
	for _, t := range xt {
		if x := int(at(PlotPoint{t, v.y0}).X); x >= 0 && x <= size.X {
			fillRect(gtx, image.Rect(x, 0, x+line, size.Y), theme.Border)
		}
	}
	for i, s := range v.series {
		c := theme.Chart[i%len(theme.Chart)]
		if v.lines {
			var pts []f32.Point
			for _, p := range s.Points {
				if !validPlotPoint(p) {
					strokePath(gtx, pts, 2*px, c)
					pts = pts[:0]
					continue
				}
				pts = append(pts, at(p))
			}
			strokePath(gtx, pts, 2*px, c)
		}
		for j, p := range s.Points {
			if !validPlotPoint(p) {
				continue
			}
			q := at(p)
			if q.X < -8 || q.Y < -8 || q.X > float32(size.X)+8 || q.Y > float32(size.Y)+8 {
				continue
			}
			r := 4 * px
			if v.lines && (v.pick[0] != i || v.pick[1] != j) {
				continue // lines show the picked point only
			}
			if v.pick[0] == i && v.pick[1] == j {
				r = 6 * px
			}
			dot(gtx, q, r, 2*px, c, theme.Surface)
		}
	}
	area.Pop()
	return core.D{Size: size}
}

// events handles wheel zoom, drag pan, double-click reset and hover picking.
func (v *PlotView) events(gtx core.C, size image.Point, at func(PlotPoint) f32.Point) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	changed := false
	for {
		ev, ok := v.drag.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			break
		}
		switch ev.Kind {
		case pointer.Press:
			v.dragging = true
			v.dragView = [4]float64{v.x0, v.x1, v.y0, v.y1}
			v.last = ev.Position
		case pointer.Cancel:
			v.cancelDrag()
			changed = true
		case pointer.Release:
			v.dragging = false
		case pointer.Drag:
			if !v.dragging {
				continue
			}
			d := ev.Position.Sub(v.last)
			v.last = ev.Position
			v.pan(-float64(d.X)/float64(size.X), float64(d.Y)/float64(size.Y))
			changed = true
		}
	}
	for {
		ev, ok := v.click.Update(gtx.Source)
		if !ok {
			break
		}
		if ev.Kind == gesture.KindClick && ev.NumClicks >= 2 {
			v.Reset()
			changed = true
		}
	}
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &v.tag, Kinds: pointer.Move | pointer.Leave | pointer.Scroll,
			ScrollY: pointer.ScrollRange{Min: -1e6, Max: 1e6}})
		if !ok {
			break
		}
		e, ok := ev.(pointer.Event)
		if !ok {
			continue
		}
		switch e.Kind {
		case pointer.Scroll:
			f := math.Pow(1.0015, min(3000, max(-3000, float64(e.Scroll.Y))))
			v.zoom(f, float64(e.Position.X)/float64(size.X), 1-float64(e.Position.Y)/float64(size.Y))
			changed = true
		case pointer.Leave:
			if v.pick[0] >= 0 {
				v.pick, changed = [2]int{-1, -1}, true
			}
		case pointer.Move:
			best, pick := float32(12*gtx.Metric.PxPerDp)*float32(12*gtx.Metric.PxPerDp), [2]int{-1, -1}
			for i, s := range v.series {
				for j, p := range s.Points {
					if !validPlotPoint(p) {
						continue
					}
					d := at(p).Sub(e.Position)
					if dd := d.X*d.X + d.Y*d.Y; dd < best {
						best, pick = dd, [2]int{i, j}
					}
				}
			}
			if pick != v.pick {
				v.pick, changed = pick, true
			}
		}
	}
	if changed {
		gtx.Execute(op.InvalidateCmd{})
	}
}
