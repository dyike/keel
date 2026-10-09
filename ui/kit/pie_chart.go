package kit

import (
	"image"
	"math"
	"slices"
	"strconv"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/event"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
)

// PieSlice is one labeled part. Only finite positive values make sectors.
type PieSlice struct {
	Name  string
	Value float64
}

type PieChartView struct {
	hoverMotion        chartHoverMotion
	noHoverAnimation   bool
	tooltipContent     func(*el.Context, PieChartTooltip) el.Element
	pointerX, pointerY float32
	title              string
	data               []PieSlice
	hidden             []bool
	height, hole       float32
	disabled, table    bool
	format             func(float64) string
	hover, tag         int
	tbl                *TableView
}

func PieChart(data ...PieSlice) *PieChartView {
	v := &PieChartView{height: 240, format: formatNumber, hover: -1}
	v.SetData(data...)
	return v
}
func (v *PieChartView) Title(s string) *PieChartView { v.title = s; return v }
func (v *PieChartView) Height(dp float32) *PieChartView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Donut cuts a central hole, clamped to 0..0.9 of the radius.
func (v *PieChartView) Donut(fraction float32) *PieChartView {
	if finiteNumber(float64(fraction)) {
		v.hole = min(.9, max(0, fraction))
	}
	return v
}
func (v *PieChartView) Format(fn func(float64) string) *PieChartView {
	if fn != nil {
		v.format = fn
	}
	return v
}
func (v *PieChartView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.hoverMotion = chartHoverMotion{}
		v.hover = -1
	}
}

// SetData copies all slices and restores their visibility.
func (v *PieChartView) SetData(data ...PieSlice) {
	v.hoverMotion = chartHoverMotion{}
	v.data = slices.Clone(data)
	v.hidden = make([]bool, len(data))
	v.hover = -1
	v.tbl = nil
}
func (v *PieChartView) fractions() []float64 {
	out := make([]float64, len(v.data))
	scale := 0.0
	for i, s := range v.data {
		if !v.hidden[i] && finiteNumber(s.Value) && s.Value > scale {
			scale = s.Value
		}
	}
	if scale == 0 {
		return out
	}
	total := 0.0
	for i, s := range v.data {
		if !v.hidden[i] && finiteNumber(s.Value) && s.Value > 0 {
			out[i] = s.Value / scale
			total += out[i]
		}
	}
	for i := range out {
		out[i] /= total
	}
	return out
}
func (v *PieChartView) valueText(i int) string {
	x := v.data[i].Value
	if !finiteNumber(x) || x < 0 {
		return "—"
	}
	return v.format(x)
}
func pieShare(x float64) string { return strconv.FormatFloat(x*100, 'f', 1, 64) + "%" }
func (v *PieChartView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	parts := v.fractions()
	box := surface().Role("figure").Name(v.title).Value(strconv.Itoa(len(v.data))).Disabled(v.disabled).P(theme.SpaceLg).Gap(10).Items(el.Stretch)
	heading := el.Div().Row().Wrap().Gap(theme.SpaceMd).Items(el.Center).Child(el.Text(v.title).Bold(), el.Div().Grow())
	toggle := text.ShowTable
	if v.table {
		toggle = text.ShowChart
	}
	heading.Child(Button(toggle, func() { v.table = !v.table }).Variant(ButtonGhost).Size(24).Render(cx))
	box.Child(heading)
	if v.table {
		if v.tbl == nil {
			v.tbl = Table(Col(""), Col(text.ChartValue).Numeric(), Col(text.ChartShare).Numeric()).Height(v.height)
		}
		v.tbl.cols[1].title = text.ChartValue
		v.tbl.cols[2].title = text.ChartShare
		rows := make([][]string, len(v.data))
		for i, s := range v.data {
			rows[i] = []string{s.Name, v.valueText(i), pieShare(parts[i])}
		}
		v.tbl.SetRows(rows)
		return box.Child(v.tbl.Render(cx))
	}
	any := false
	for _, f := range parts {
		any = any || f > 0
	}
	if !any {
		box.Child(el.Div().H(el.Dp(v.height)).Center().Child(el.Text(text.NoData).TextColor(theme.Muted)))
	} else {
		id := autoID("pie-plot", v)
		plot := el.Div().ID(id).H(el.Dp(v.height)).WFull().Child(el.Widget(core.Func(func(gtx core.C) core.D { return v.draw(gtx, parts) })).HFull().WFull())
		v.renderTooltip(cx, id, parts, plot)
		box.Child(plot)
	}
	if v.hover >= 0 && v.hover < len(v.data) && parts[v.hover] > 0 {
		box.Child(el.Text(v.data[v.hover].Name + "  " + v.valueText(v.hover) + "  " + pieShare(parts[v.hover])).TextSize(theme.TextSm))
	} else {
		box.Child(el.Div().H(el.Dp(16)))
	}
	legend := el.Div().Row().Wrap().Gap(theme.SpaceSm)
	for i, s := range v.data {
		legend.Child(legendItem(autoID("pie", v)+"/legend/"+strconv.Itoa(i), s.Name, theme.Chart[i%len(theme.Chart)], 5, !v.hidden[i], func() { v.hidden[i] = !v.hidden[i]; v.hover = -1 }))
	}
	return box.Child(legend)
}
func pieHit(x, y, r, hole float64, parts []float64) int {
	d := math.Hypot(x, y)
	if d > r || d < r*hole {
		return -1
	}
	angle := math.Atan2(y, x) + math.Pi/2
	if angle < 0 {
		angle += 2 * math.Pi
	}
	part := angle / (2 * math.Pi)
	total := 0.0
	for i, p := range parts {
		total += p
		if p > 0 && part < total {
			return i
		}
	}
	return -1
}
func (v *PieChartView) draw(gtx core.C, parts []float64) core.D {
	size := gtx.Constraints.Max
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	r := float64(min(size.X, size.Y))/2 - 2
	if r <= 0 {
		return core.D{Size: size}
	}
	if gtx.Enabled() {
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &v.tag, Kinds: pointer.Move | pointer.Enter | pointer.Leave})
			if !ok {
				break
			}
			e := ev.(pointer.Event)
			hit := -1
			if e.Kind != pointer.Leave {
				hit = pieHit(float64(e.Position.X-center.X), float64(e.Position.Y-center.Y), r, float64(v.hole), parts)
			}
			scale := gtx.Metric.PxPerDp
			if scale <= 0 {
				scale = 1
			}
			x, y := e.Position.X/scale, e.Position.Y/scale
			if hit >= 0 && (v.pointerX != x || v.pointerY != y) {
				v.pointerX, v.pointerY = x, y
				if v.tooltipContent != nil {
					gtx.Execute(op.InvalidateCmd{})
				}
			}
			if hit != v.hover {
				v.hover = hit
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
	v.hoverMotion.update(gtx, len(v.data), v.hover, !v.noHoverAnimation && !v.disabled)
	area := clip.Rect(image.Rectangle{Max: size}).Push(gtx.Ops)
	defer area.Pop()
	event.Op(gtx.Ops, &v.tag)
	angle := -math.Pi / 2
	for i, f := range parts {
		if f <= 0 {
			continue
		}
		end := angle + f*2*math.Pi
		steps := max(2, int(math.Ceil(f*2*math.Pi*r/2)))
		point := func(a, rad float64) f32.Point {
			return f32.Pt(center.X+float32(math.Cos(a)*rad), center.Y+float32(math.Sin(a)*rad))
		}
		var p clip.Path
		p.Begin(gtx.Ops)
		p.MoveTo(point(angle, r))
		for j := 1; j <= steps; j++ {
			p.LineTo(point(angle+(end-angle)*float64(j)/float64(steps), r))
		}
		if v.hole > 0 {
			for j := steps; j >= 0; j-- {
				p.LineTo(point(angle+(end-angle)*float64(j)/float64(steps), r*float64(v.hole)))
			}
		} else {
			p.LineTo(center)
		}
		p.Close()
		path := p.End()
		paint.FillShape(gtx.Ops, theme.Chart[i%len(theme.Chart)], clip.Outline{Path: path}.Op())
		if emphasis := v.hoverMotion.weight(i); emphasis > 0 {
			color := theme.Text
			color.A = uint8(float32(color.A) * emphasis)
			paint.FillShape(gtx.Ops, color, clip.Stroke{Path: path, Width: float32(gtx.Dp(2))}.Op())
		}
		angle = end
	}
	return core.D{Size: size}
}
