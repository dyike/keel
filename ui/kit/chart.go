package kit

import (
	"image"
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
)

const axisWidth = 52 // dp for the y-axis labels

// ChartView draws categorical data as lines or bars over one y-axis. Series
// take the theme's Chart colors in order: series i is always slot i, so a
// series keeps its color when others are added or removed. Hovering shows a
// crosshair and a tooltip with every series' value; the data-table toggle
// shows the same numbers as a Table. Two or more series get a legend.
type ChartView struct {
	kind    ChartKind
	title   string
	labels  []string
	series  []Series
	height  float32
	stacked bool
	format  func(float64) string
	table   bool
	tbl     *TableView
	hover   int
	tag     int     // pointer handler tag
	plotW   float32 // painted plot width, dp
}

// LineChart plots each series as a line across the labels, e.g. months.
func LineChart(labels []string, series ...Series) *ChartView {
	return &ChartView{kind: ChartLine, labels: labels, series: series, height: 220, format: formatNumber, hover: -1}
}

// BarChart draws a group of bars per label, one per series; Stacked piles them.
func BarChart(labels []string, series ...Series) *ChartView {
	v := LineChart(labels, series...)
	v.kind = ChartBar
	return v
}

// Title names the chart; it also names it for agents.
func (v *ChartView) Title(s string) *ChartView { v.title = s; return v }

// Height sets the plot height in dp, 220 by default.
func (v *ChartView) Height(dp float32) *ChartView {
	if dp > 0 {
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
	v.labels, v.series = labels, series
	if v.hover >= len(labels) {
		v.hover = -1
	}
}

func (v *ChartView) value(s Series, i int) float64 {
	if i < len(s.Values) {
		return s.Values[i]
	}
	return 0
}

// span returns the y range the marks need; bars always include zero.
func (v *ChartView) span() (lo, hi float64) {
	first := true
	add := func(x float64) {
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
			for _, s := range v.series {
				if x := v.value(s, i); x >= 0 {
					pos += x
				} else {
					neg += x
				}
			}
			add(pos)
			add(neg)
			continue
		}
		for _, s := range v.series {
			add(v.value(s, i))
		}
	}
	if v.kind == ChartBar || first {
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
	root := el.Div().Role("figure").Name(name).Value(strconv.Itoa(len(v.labels))+"x"+strconv.Itoa(len(v.series))).
		Gap(10).P(12).Rounded(8).Bg(theme.Surface).Border(1, theme.Border).Items(el.Stretch)
	head := el.Div().Row().Items(el.Center).Gap(12)
	if v.title != "" {
		head.Child(el.Text(v.title).Bold())
	}
	head.Child(el.Div().Grow())
	if len(v.series) > 1 {
		for i, s := range v.series {
			head.Child(el.Div().Row().Items(el.Center).Gap(6).Child(
				el.Div().Size(el.Dp(10)).Rounded(2).Bg(theme.Chart[i%len(theme.Chart)]),
				el.Text(s.Name).TextSize(12).TextColor(theme.Muted)))
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

	lo, hi := v.span()
	ticks := niceTicks(lo, hi, 4)
	lo, hi = ticks[0], ticks[len(ticks)-1]
	axis := el.Div().W(el.Dp(axisWidth)).H(el.Dp(v.height)).NoShrink()
	for _, t := range ticks {
		top := float32((1-(t-lo)/(hi-lo))*float64(v.height)) - 8
		axis.Child(el.Div().Absolute().Top(top).Right(8).Child(el.Text(v.format(t)).TextSize(11).TextColor(theme.Muted)))
	}
	plot := el.Div().Grow().W(el.Dp(0)).H(el.Dp(v.height)).Items(el.Stretch).Child(
		el.Widget(core.Func(func(gtx core.C) core.D { return v.draw(gtx, lo, hi, ticks) })).H(el.Dp(v.height)))
	if tip := v.tooltip(); tip != nil {
		plot.Child(tip)
	}
	root.Child(el.Div().Row().Items(el.Start).Child(axis, plot))

	// Category labels under the plot, thinned so they never collide.
	xs := el.Div().Row().Pl(axisWidth)
	every := 1
	if v.plotW > 0 && len(v.labels) > 0 {
		every = max(1, int(56*float32(len(v.labels))/v.plotW+0.999))
	}
	for i, l := range v.labels {
		cell := el.Div().Flex(1).W(el.Dp(0)).Items(el.Center)
		if i%every == 0 {
			cell.Child(el.Text(l).TextSize(11).TextColor(theme.Muted).MaxLines(1))
		}
		xs.Child(cell)
	}
	return root.Child(xs)
}

func (v *ChartView) tooltip() el.Element {
	if v.hover < 0 || v.hover >= len(v.labels) || v.plotW <= 0 {
		return nil
	}
	band := v.plotW / float32(len(v.labels))
	center := (float32(v.hover) + 0.5) * band
	const w = 168
	left := center + 12
	if left+w > v.plotW {
		left = center - 12 - w
	}
	tip := el.Div().Absolute().Top(8).Left(max(left, 0)).W(el.Dp(w)).P(8).Gap(4).Rounded(6).
		Bg(theme.Surface).Border(1, theme.Border).Items(el.Stretch).
		Child(el.Text(v.labels[v.hover]).TextSize(12).Bold())
	for i, s := range v.series {
		tip.Child(el.Div().Row().Items(el.Center).Gap(6).Child(
			el.Div().Size(el.Dp(8)).Rounded(4).Bg(theme.Chart[i%len(theme.Chart)]),
			el.Text(s.Name).TextSize(12).TextColor(theme.Muted).Grow().MaxLines(1),
			el.Text(v.format(v.value(s, v.hover))).TextSize(12)))
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
			rows[i] = append(rows[i], v.format(v.value(s, i)))
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
	area.Pop()

	y := func(x float64) float32 { return float32(size.Y) * float32(1-(x-lo)/(hi-lo)) }
	for _, t := range ticks {
		yy := int(y(t))
		fillRect(gtx, image.Rect(0, yy, size.X, yy+max(1, int(px))), theme.Border)
	}
	if v.hover >= 0 && v.kind == ChartBar {
		x0 := int(float32(v.hover) * band)
		fillRect(gtx, image.Rect(x0, 0, int(float32(v.hover+1)*band), size.Y), theme.Subtle)
		for _, t := range ticks { // keep the grid over the hover band
			yy := int(y(t))
			fillRect(gtx, image.Rect(x0, yy, int(float32(v.hover+1)*band), yy+max(1, int(px))), theme.Border)
		}
	}
	switch v.kind {
	case ChartBar:
		v.drawBars(gtx, band, y, dp)
	default:
		for i, s := range v.series {
			pts := make([]f32.Point, 0, n)
			for j := 0; j < n; j++ {
				pts = append(pts, f32.Pt((float32(j)+0.5)*band, y(v.value(s, j))))
			}
			strokePath(gtx, pts, dp(2), theme.Chart[i%len(theme.Chart)])
		}
		if v.hover >= 0 && v.hover < n {
			x := (float32(v.hover) + 0.5) * band
			fillRect(gtx, image.Rect(int(x), 0, int(x)+max(1, int(px)), size.Y), theme.Muted)
			for i, s := range v.series {
				dot(gtx, f32.Pt(x, y(v.value(s, v.hover))), dp(4), dp(2), theme.Chart[i%len(theme.Chart)], theme.Surface)
			}
		}
	}
	return core.D{Size: size}
}

// drawBars draws grouped or stacked bars: at most 24dp wide, 2dp of surface
// between neighbors, the data end rounded 4dp and the baseline end square.
func (v *ChartView) drawBars(gtx core.C, band float32, y func(float64) float32, dp func(float32) float32) {
	k := max(len(v.series), 1)
	gap := dp(2)
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
		paint.FillShape(gtx.Ops, theme.Chart[i%len(theme.Chart)], rr.Op(gtx.Ops))
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
				if x := v.value(s, j); x > 0 {
					lastPos = i
				} else if x < 0 {
					lastNeg = i
				}
			}
			pos, neg := 0.0, 0.0
			for i, s := range v.series {
				x := v.value(s, j)
				if x >= 0 {
					from, to := y(pos), y(pos+x)
					if pos > 0 {
						from -= gap // the surface gap between segments
					}
					bar(x0, x1, from, to, i, i == lastPos)
					pos += x
				} else {
					from, to := y(neg), y(neg+x)
					if neg < 0 {
						from += gap
					}
					bar(x0, x1, from, to, i, i == lastNeg)
					neg += x
				}
			}
			continue
		}
		w := min(dp(24), (band*0.7-gap*float32(k-1))/float32(k))
		x := center - (w*float32(k)+gap*float32(k-1))/2
		for i, s := range v.series {
			bar(x, x+w, zero, y(v.value(s, j)), i, true)
			x += w + gap
		}
	}
}
