package kit

import (
	"image"
	"math"
	"slices"
	"strconv"
	"strings"

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
)

// Series is one named row of values, one value per chart label.
type Series struct {
	Name   string
	Values []float64
}

// ChartKind selects how a Chart draws its series.
type ChartKind uint8

const (
	ChartLine ChartKind = iota
	ChartBar
	ChartArea
	ChartCandlestick
	ChartRadar
)

const axisWidth = 52 // dp for the y-axis labels

// ChartView draws categorical data as lines or bars over one y-axis. Series
// take the theme's Chart colors in order: series i is always slot i, so a
// series keeps its color when others are added or removed. Hovering shows a
// crosshair and a tooltip with every series' value; the data-table toggle
// shows the same numbers as a Table. Two or more series get a legend.
type ChartView struct {
	pointerX, pointerY float32
	options            chartOptions
	kind               ChartKind
	title              string
	labels             []string
	series             []Series
	height             float32
	stacked            bool
	format             func(float64) string
	table              bool
	tbl                *TableView
	hover              int
	tag                int // pointer handler tag
	candles            []Candle
	hidden             []bool
	disabled           bool
	plotW              float32 // painted plot width, dp
}

// LineChart plots each series as a line across the labels, e.g. months.
func LineChart(labels []string, series ...Series) *ChartView {
	v := &ChartView{kind: ChartLine, height: 220, format: formatNumber, hover: -1}
	v.SetData(labels, series...)
	return v
}

// BarChart draws a group of bars per label, one per series; Stacked piles them.
func BarChart(labels []string, series ...Series) *ChartView {
	v := LineChart(labels, series...)
	v.kind = ChartBar
	return v
}

// AreaChart fills each line to zero. Series overlap with translucent colors.
func AreaChart(labels []string, series ...Series) *ChartView {
	v := LineChart(labels, series...)
	v.kind = ChartArea
	return v
}

// Title names the chart; it also names it for agents.
func (v *ChartView) Title(s string) *ChartView { v.title = s; return v }

// Height sets the plot height in dp, 220 by default.
func (v *ChartView) Height(dp float32) *ChartView {
	if dp > 0 && finiteNumber(float64(dp)) {
		v.height = dp
	}
	return v
}

// Stacked piles a bar chart's series on each other instead of side by side.
func (v *ChartView) Stacked() *ChartView { v.stacked = true; return v }

// Format writes values for the axis, tooltip and table; by default with
// thousands separators and up to two decimals.
func (v *ChartView) Format(fn func(float64) string) *ChartView {
	if fn != nil {
		v.format = fn
	}
	return v
}

// SetData replaces the labels and series.
func (v *ChartView) SetData(labels []string, series ...Series) {
	v.labels, v.series = slices.Clone(labels), slices.Clone(series)
	for i := range v.series {
		v.series[i].Values = slices.Clone(series[i].Values)
	}
	v.hidden = make([]bool, len(series))
	v.tbl = nil
	if v.hover >= len(labels) {
		v.hover = -1
	}
}

func (v *ChartView) value(s Series, i int) float64 {
	if i >= 0 && i < len(s.Values) {
		return s.Values[i]
	}
	return math.NaN()
}

func (v *ChartView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.hover = -1
	}
}
func (v *ChartView) valueText(s Series, i int) string {
	x := v.value(s, i)
	if !finiteNumber(x) {
		return "—"
	}
	return v.format(x)
}

// span returns the y range the marks need; bars always include zero.
func (v *ChartView) span() (lo, hi float64) {
	first := true
	add := func(x float64) {
		if !finiteNumber(x) {
			return
		}
		if first || x < lo {
			lo = x
		}
		if first || x > hi {
			hi = x
		}
		first = false
	}
	for i := range v.labels {
		if v.kind == ChartBar && v.stacked {
			pos, neg := 0.0, 0.0
			for si, s := range v.series {
				if v.hidden[si] {
					continue
				}
				if x := v.value(s, i); !finiteNumber(x) {
					continue
				} else if x >= 0 {
					pos = min(math.MaxFloat64, pos+x)
				} else {
					neg = max(-math.MaxFloat64, neg+x)
				}
			}
			add(pos)
			add(neg)
			continue
		}
		for si, s := range v.series {
			if v.hidden[si] {
				continue
			}
			add(v.value(s, i))
		}
	}
	if v.kind == ChartBar || v.kind == ChartArea || first {
		add(0)
	}
	return lo, hi
}

func (v *ChartView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	names := make([]string, len(v.series))
	for i, s := range v.series {
		names[i] = s.Name
	}
	name := v.title
	if name == "" {
		name = strings.Join(names, ", ")
	}
	root := el.Div().Disabled(v.disabled).Role("figure").Name(name).Value(strconv.Itoa(len(v.labels))+"x"+strconv.Itoa(len(v.series))).
		Gap(10).P(theme.SpaceLg).Rounded(theme.RadiusLg).Bg(theme.Surface).Border(1, theme.Border).Items(el.Stretch)
	head := el.Div().Row().Wrap().Items(el.Center).Gap(theme.SpaceLg)
	if v.title != "" {
		head.Child(el.Text(v.title).Bold())
	}
	head.Child(el.Div().Grow())
	if len(v.series) > 1 && v.kind != ChartCandlestick {
		for i, s := range v.series {
			head.Child(legendItem(autoID("chart", v)+"/legend/"+strconv.Itoa(i), s.Name, v.seriesColor(i), 2, !v.hidden[i], func() { v.hidden[i] = !v.hidden[i] }))
		}
	}
	toggle := text.ShowTable
	if v.table {
		toggle = text.ShowChart
	}
	head.Child(Button(toggle, func() { v.table = !v.table }).Variant(ButtonGhost).Size(24).Render(cx))
	root.Child(head)
	if v.table {
		return root.Child(v.dataTable(cx))
	}

	if v.kind == ChartRadar {
		return root.Child(v.radar(cx))
	}
	ticks := v.axisTicks()
	lo, hi := ticks[0], ticks[len(ticks)-1]
	axis := el.Div().W(el.Dp(axisWidth)).H(el.Dp(v.height)).NoShrink()
	for _, t := range ticks {
		top := float32((1-axisFraction(t, lo, hi))*float64(v.height)) - 8
		axis.Child(el.Div().Absolute().Top(top).Right(8).Child(el.Text(v.format(t)).TextSize(theme.TextXs).TextColor(theme.Muted)))
	}
	plot := el.Div().Grow().W(el.Dp(0)).H(el.Dp(v.height)).Items(el.Stretch).Child(
		el.Widget(core.Func(func(gtx core.C) core.D { return v.draw(gtx, lo, hi, ticks) })).H(el.Dp(v.height)))
	for _, line := range v.options.references {
		if line.Label != "" && line.Value >= lo && line.Value <= hi {
			plot.Child(el.Div().Absolute().Left(4).Top(float32(1-axisFraction(line.Value, lo, hi))*v.height - 16).Child(el.Text(line.Label).TextSize(theme.TextXs).TextColor(line.Color)))
		}
	}
	if tip := v.tooltip(cx); tip != nil {
		plot.Child(tip)
	}
	root.Child(el.Div().Row().Items(el.Start).Child(axis, plot))

	// Build only the labels that can be read at this width, even on first mount.
	width := v.plotW
	if width <= 0 {
		width, _ = cx.ViewportSize()
		width = max(1, width-axisWidth-24)
	}
	indices := v.xTickIndices(width)
	// Bands are flex weights, so labels line up with the plot whatever its
	// painted width; runs of unlabeled bands collapse into one spacer. A label
	// is centered on its band and may overhang it.
	xs := el.Div().Row().H(el.Dp(18)).Grow().W(el.Dp(0))
	next := 0
	for _, i := range indices {
		if i > next {
			xs.Child(el.Div().W(el.Dp(0)).Flex(float32(i - next)))
		}
		xs.Child(el.Div().W(el.Dp(0)).Flex(1).Items(el.Center).Child(
			el.Div().W(el.Dp(56)).NoShrink().Items(el.Center).
				Child(el.Text(v.labels[i]).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1))))
		next = i + 1
	}
	if next < len(v.labels) {
		xs.Child(el.Div().W(el.Dp(0)).Flex(float32(len(v.labels) - next)))
	}
	return root.Child(el.Div().Row().Child(el.Div().W(el.Dp(axisWidth)).NoShrink(), xs))
}

func (v *ChartView) tooltip(cx *el.Context) el.Element {
	if v.hover < 0 || v.hover >= len(v.labels) || v.plotW <= 0 {
		return nil
	}
	band := v.plotW / float32(len(v.labels))
	center := (float32(v.hover) + 0.5) * band
	w := min(float32(168), v.plotW)
	left := center + 12
	if left+w > v.plotW {
		left = center - 12 - w
	}
	return v.tooltipPanel(cx).Absolute().Top(8).Left(max(left, 0)).W(el.Dp(w))
}

func (v *ChartView) tooltipPanel(cx *el.Context) *el.DivEl {
	tip := el.Div().P(theme.SpaceMd).Gap(theme.SpaceXs).Rounded(theme.RadiusMd).
		Bg(theme.Surface).Border(1, theme.Border).Items(el.Stretch).
		Child()
	if v.options.tooltip != nil {
		return tip.Child(v.options.tooltip(cx, v.tooltipData()))
	}
	tip.Child(el.Text(v.labels[v.hover]).TextSize(theme.TextSm).Bold())
	for i, s := range v.series {
		if v.hidden[i] {
			continue
		}
		tip.Child(el.Div().Row().Items(el.Center).Gap(theme.SpaceSm).Child(
			el.Div().Size(el.Dp(8)).Rounded(theme.RadiusSm).Bg(v.seriesColor(i)),
			el.Text(s.Name).TextSize(theme.TextSm).TextColor(theme.Muted).Grow().MaxLines(1),
			el.Text(v.valueText(s, v.hover)).TextSize(theme.TextSm)))
	}
	return tip
}

func (v *ChartView) dataTable(cx *el.Context) el.Element {
	if v.tbl == nil || len(v.tbl.cols) != len(v.series)+1 {
		cols := []*ColumnSpec{Col("")}
		for _, s := range v.series {
			cols = append(cols, Col(s.Name).Numeric())
		}
		v.tbl = Table(cols...).Height(v.height)
	}
	rows := make([][]string, len(v.labels))
	for i, l := range v.labels {
		rows[i] = []string{l}
		for _, s := range v.series {
			rows[i] = append(rows[i], v.valueText(s, i))
		}
	}
	v.tbl.SetRows(rows)
	return v.tbl.Render(cx)
}

// draw paints gridlines and marks into the plot box, and tracks the hovered label.
func (v *ChartView) draw(gtx core.C, lo, hi float64, ticks []float64) core.D {
	size := gtx.Constraints.Max
	px := gtx.Metric.PxPerDp
	if px <= 0 {
		px = 1
	}
	dp := func(x float32) float32 { return x * px }
	if w := float32(size.X) / px; w != v.plotW && gtx.Enabled() {
		v.plotW = w
		gtx.Execute(op.InvalidateCmd{}) // label thinning and the tooltip need this width
	}
	n := len(v.labels)
	band := float32(size.X) / float32(max(n, 1))
	if gtx.Enabled() {
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &v.tag, Kinds: pointer.Move | pointer.Enter | pointer.Leave})
			if !ok {
				break
			}
			e, ok := ev.(pointer.Event)
			if !ok {
				continue
			}
			h := -1
			if e.Kind != pointer.Leave && n > 0 {
				h = min(max(int(e.Position.X/band), 0), n-1)
			}
			if h != v.hover {
				v.hover = h
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	event.Op(gtx.Ops, &v.tag)
	defer area.Pop()

	y := func(x float64) float32 {
		return float32(size.Y) * float32(1-max(-1e6, min(1e6, axisFraction(x, lo, hi))))
	}
	for _, t := range ticks {
		yy := int(y(t))
		v.drawGuide(gtx, f32.Pt(0, float32(yy)), f32.Pt(float32(size.X), float32(yy)), px, theme.Border, v.options.gridDashed)
	}
	if v.hover >= 0 && v.kind == ChartBar {
		x0 := int(float32(v.hover) * band)
		fillRect(gtx, image.Rect(x0, 0, int(float32(v.hover+1)*band), size.Y), theme.Subtle)
		for _, t := range ticks { // keep the grid over the hover band
			yy := int(y(t))
			fillRect(gtx, image.Rect(x0, yy, int(float32(v.hover+1)*band), yy+max(1, int(px))), theme.Border)
		}
	}
	for i := 1; i <= v.options.gridColumns; i++ {
		x := float32(size.X) * float32(i) / float32(v.options.gridColumns+1)
		v.drawGuide(gtx, f32.Pt(x, 0), f32.Pt(x, float32(size.Y)), px, theme.Border, v.options.gridDashed)
	}
	for _, line := range v.options.references {
		if line.Value >= lo && line.Value <= hi {
			v.drawGuide(gtx, f32.Pt(0, y(line.Value)), f32.Pt(float32(size.X), y(line.Value)), px, line.Color, true)
		}
	}
	switch v.kind {
	case ChartCandlestick:
		v.drawCandles(gtx, y, dp)
	case ChartBar:
		v.drawBars(gtx, band, y, dp)
	default:
		// Fills first, lines second: a later series' fill must not tint an
		// earlier series' line.
		if v.kind == ChartArea {
			visible := 0
			for i := range v.series {
				if !v.hidden[i] {
					visible++
				}
			}
			for i, s := range v.series {
				if v.hidden[i] {
					continue
				}
				for _, points := range chartSegments(s.Values, n, band, y, max(size.X, 1)) {
					points = v.curvePoints(points)
					if len(points) < 2 {
						continue
					}
					var path clip.Path
					path.Begin(gtx.Ops)
					path.MoveTo(f32.Pt(points[0].X, y(0)))
					for _, p := range points {
						path.LineTo(p)
					}
					path.LineTo(f32.Pt(points[len(points)-1].X, y(0)))
					path.Close()
					fill := v.seriesColor(i)
					fill.A = areaAlpha(visible)
					if c := v.options.styles[i].Fill; c != nil {
						fill = *c
					}
					paint.FillShape(gtx.Ops, fill, clip.Outline{Path: path.End()}.Op())
				}
			}
		}
		for i, s := range v.series {
			if v.hidden[i] {
				continue
			}
			for _, points := range chartSegments(s.Values, n, band, y, max(size.X, 1)) {
				vertices := points
				points = v.curvePoints(points)
				color := v.seriesColor(i)
				if len(points) == 1 {
					dot(gtx, points[0], dp(2), 0, color, theme.Surface)
				} else {
					strokePath(gtx, points, dp(v.seriesWidth(i)), color)
					if v.options.styles[i].Dots {
						for _, p := range vertices {
							dot(gtx, p, dp(2), 0, color, theme.Surface)
						}
					}
				}
			}
		}
		if v.hover >= 0 && v.hover < n {
			x := (float32(v.hover) + 0.5) * band
			fillRect(gtx, image.Rect(int(x), 0, int(x)+max(1, int(px)), size.Y), theme.Muted)
			for i, s := range v.series {
				if v.hidden[i] || !finiteNumber(v.value(s, v.hover)) {
					continue
				}
				dot(gtx, f32.Pt(x, y(v.value(s, v.hover))), dp(4), dp(2), v.seriesColor(i), theme.Surface)
			}
		}
	}
	return core.D{Size: size}
}

// areaAlpha keeps stacked translucent fills from turning muddy where several
// series overlap: one series gets a clear tint, several get a light one.
func areaAlpha(visible int) uint8 {
	if visible <= 1 {
		return 64
	}
	return 28
}

// drawBars draws grouped or stacked bars: at most 24dp wide, 2dp of surface
// between neighbors, the data end rounded 4dp and the baseline end square.
func (v *ChartView) drawBars(gtx core.C, band float32, y func(float64) float32, dp func(float32) float32) {
	k := 0
	for i := range v.series {
		if !v.hidden[i] {
			k++
		}
	}
	k = max(k, 1)
	gap := min(dp(2), band*0.1/float32(k))
	bar := func(x0, x1, from, to float32, i int, round bool) {
		top, bottom := min(from, to), max(from, to)
		if bottom-top < 0.5 {
			return
		}
		r := min(dp(4), (x1-x0)/2, bottom-top)
		if !round {
			r = 0
		}
		rr := clip.RRect{Rect: image.Rect(int(x0), int(top), int(x1), int(bottom))}
		if to < from { // rising bar: round the top
			rr.NW, rr.NE = int(r), int(r)
		} else {
			rr.SW, rr.SE = int(r), int(r)
		}
		paint.FillShape(gtx.Ops, v.seriesColor(i), rr.Op(gtx.Ops))
	}
	zero := y(0)
	for j := range v.labels {
		center := (float32(j) + 0.5) * band
		if v.stacked {
			w := min(dp(24), band*0.6)
			x0, x1 := center-w/2, center+w/2
			// Only the outermost segment on each side gets the rounded end.
			lastPos, lastNeg := -1, -1
			for i, s := range v.series {
				if v.hidden[i] {
					continue
				}
				if x := v.value(s, j); x > 0 {
					lastPos = i
				} else if x < 0 {
					lastNeg = i
				}
			}
			pos, neg := 0.0, 0.0
			for i, s := range v.series {
				if v.hidden[i] {
					continue
				}
				x := v.value(s, j)
				if !finiteNumber(x) {
					continue
				}
				if x >= 0 {
					from, to := y(pos), y(min(math.MaxFloat64, pos+x))
					if pos > 0 {
						from -= gap // the surface gap between segments
					}
					bar(x0, x1, from, to, i, i == lastPos)
					pos = min(math.MaxFloat64, pos+x)
				} else {
					from, to := y(neg), y(max(-math.MaxFloat64, neg+x))
					if neg < 0 {
						from += gap
					}
					bar(x0, x1, from, to, i, i == lastNeg)
					neg = max(-math.MaxFloat64, neg+x)
				}
			}
			continue
		}
		w := min(dp(24), (band*0.7-gap*float32(k-1))/float32(k))
		x := center - (w*float32(k)+gap*float32(k-1))/2
		for i, s := range v.series {
			if v.hidden[i] {
				continue
			}
			if !finiteNumber(v.value(s, j)) {
				x += w + gap
				continue
			}
			bar(x, x+w, zero, y(v.value(s, j)), i, true)
			x += w + gap
		}
	}
}
