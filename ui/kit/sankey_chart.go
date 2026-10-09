package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/event"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/third_party/gio/io/semantic"
	"github.com/dyike/keel/third_party/gio/op"
	"github.com/dyike/keel/third_party/gio/op/clip"
	"github.com/dyike/keel/third_party/gio/op/paint"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"image"
	"image/color"
	"math"
	"slices"
	"strconv"
)

// SankeyNode is a named flow endpoint. Nil Color uses the theme palette.
type SankeyNode struct {
	Name  string
	Color *color.NRGBA
}

// SankeyLink references node indices; Value must be finite and nonnegative.
type SankeyLink struct {
	Source, Target int
	Value          float64
}
type SankeyAlign uint8

const (
	SankeyAlignJustify SankeyAlign = iota
	SankeyAlignLeft
	SankeyAlignRight
	SankeyAlignCenter
)

type SankeyValueScale uint8

const (
	SankeyValueScaleLinear SankeyValueScale = iota
	SankeyValueScaleSqrt
)

type SankeyTooltip struct {
	Index int
	Node  SankeyNode
	Value float64
}
type sankeyRect struct{ x, y, w, h float32 }
type sankeyGeometry struct {
	nodes []sankeyRect
	links [][4]float32
}

// SankeyChartView lays out a directed acyclic flow graph, with raw throughput
// labels, hover highlighting and a data table. Node throughput is max(in,out).
type SankeyChartView struct {
	hoverMotion                                        chartHoverMotion
	noHoverAnimation                                   bool
	minLinkWidth                                       float32
	nodes                                              []SankeyNode
	links                                              []SankeyLink
	order                                              []int
	values                                             []float64
	title                                              string
	height, width, nodeWidth, padding, radius, opacity float32
	align                                              SankeyAlign
	scale                                              SankeyValueScale
	iterations                                         int
	hover, tag                                         int
	disabled, table                                    bool
	err                                                error
	tbl                                                *TableView
	format                                             func(float64) string
	labels                                             func(*el.Context, SankeyTooltip) el.Element
	tooltip                                            func(*el.Context, SankeyTooltip) el.Element
}

func SankeyChart(nodes []SankeyNode, links []SankeyLink) *SankeyChartView {
	v := &SankeyChartView{height: 300, nodeWidth: 10, padding: 16, opacity: .3, iterations: 6, hover: -1, format: formatNumber}
	v.err = v.SetData(nodes, links)
	return v
}

// SetData validates all references, totals and topology before replacing data.
// Cycles, self-links, negative/non-finite values and overflowing totals return
// errors and preserve the previous graph. Inputs and optional colors are copied.
func (v *SankeyChartView) SetData(nodes []SankeyNode, links []SankeyLink) error {
	incoming, outgoing := make([]float64, len(nodes)), make([]float64, len(nodes))
	degree := make([]int, len(nodes))
	adj := make([][]int, len(nodes))
	for i, l := range links {
		if l.Source < 0 || l.Source >= len(nodes) || l.Target < 0 || l.Target >= len(nodes) || l.Source == l.Target || !finiteNumber(l.Value) || l.Value < 0 {
			return fmt.Errorf("sankey: invalid link %d", i)
		}
		incoming[l.Target] += l.Value
		outgoing[l.Source] += l.Value
		if !finiteNumber(incoming[l.Target]) || !finiteNumber(outgoing[l.Source]) {
			return fmt.Errorf("sankey: flow total overflow")
		}
		degree[l.Target]++
		adj[l.Source] = append(adj[l.Source], l.Target)
	}
	order := []int{}
	for i, d := range degree {
		if d == 0 {
			order = append(order, i)
		}
	}
	for head := 0; head < len(order); head++ {
		for _, j := range adj[order[head]] {
			degree[j]--
			if degree[j] == 0 {
				order = append(order, j)
			}
		}
	}
	if len(order) != len(nodes) {
		return fmt.Errorf("sankey: cyclic flow graph")
	}
	owned := slices.Clone(nodes)
	for i := range owned {
		if owned[i].Color != nil {
			c := *owned[i].Color
			owned[i].Color = &c
		}
	}
	values := make([]float64, len(nodes))
	for i := range values {
		values[i] = max(incoming[i], outgoing[i])
	}
	v.hoverMotion = chartHoverMotion{}
	v.nodes, v.links, v.order, v.values = owned, slices.Clone(links), order, values
	v.hover = -1
	v.tbl = nil
	v.err = nil
	return nil
}
func (v *SankeyChartView) Error() error                    { return v.err }
func (v *SankeyChartView) Title(s string) *SankeyChartView { v.title = s; return v }
func (v *SankeyChartView) Height(dp float32) *SankeyChartView {
	if finiteNumber(float64(dp)) && dp > 0 {
		v.height = dp
	}
	return v
}
func (v *SankeyChartView) NodeWidth(dp float32) *SankeyChartView {
	if finiteNumber(float64(dp)) && dp > 0 && dp <= 128 {
		v.nodeWidth = dp
	}
	return v
}
func (v *SankeyChartView) NodePadding(dp float32) *SankeyChartView {
	if finiteNumber(float64(dp)) && dp >= 0 && dp <= 128 {
		v.padding = dp
	}
	return v
}
func (v *SankeyChartView) NodeRadius(dp float32) *SankeyChartView {
	if finiteNumber(float64(dp)) && dp >= 0 && dp <= 64 {
		v.radius = dp
	}
	return v
}
func (v *SankeyChartView) LinkOpacity(a float32) *SankeyChartView {
	if finiteNumber(float64(a)) && a >= 0 && a <= 1 {
		v.opacity = a
	}
	return v
}
func (v *SankeyChartView) NodeAlign(a SankeyAlign) *SankeyChartView {
	if a <= SankeyAlignCenter {
		v.align = a
	}
	return v
}
func (v *SankeyChartView) ValueScale(s SankeyValueScale) *SankeyChartView {
	if s <= SankeyValueScaleSqrt {
		v.scale = s
	}
	return v
}
func (v *SankeyChartView) Iterations(n int) *SankeyChartView {
	if n >= 0 && n <= 64 {
		v.iterations = n
	}
	return v
}
func (v *SankeyChartView) Format(fn func(float64) string) *SankeyChartView {
	if fn != nil {
		v.format = fn
	}
	return v
}

// Labels replaces the text beside each node; nil restores name and raw throughput.
func (v *SankeyChartView) Labels(fn func(*el.Context, SankeyTooltip) el.Element) *SankeyChartView {
	v.labels = fn
	return v
}
func (v *SankeyChartView) TooltipContent(fn func(*el.Context, SankeyTooltip) el.Element) *SankeyChartView {
	v.tooltip = fn
	return v
}
func (v *SankeyChartView) SetDisabled(on bool) {
	v.disabled = on
	if on {
		v.hoverMotion = chartHoverMotion{}
		v.hover = -1
	}
}
func (v *SankeyChartView) nodeColor(i int) color.NRGBA {
	if v.nodes[i].Color != nil {
		return *v.nodes[i].Color
	}
	return theme.Chart[i%len(theme.Chart)]
}
func (v *SankeyChartView) datum(i int) SankeyTooltip {
	node := v.nodes[i]
	if node.Color != nil {
		c := *node.Color
		node.Color = &c
	}
	return SankeyTooltip{i, node, v.values[i]}
}
func (v *SankeyChartView) layout(w, h float32) sankeyGeometry {
	g := sankeyGeometry{nodes: make([]sankeyRect, len(v.nodes)), links: make([][4]float32, len(v.links))}
	if len(v.nodes) == 0 {
		return g
	}
	depth, tail := make([]int, len(v.nodes)), make([]int, len(v.nodes))
	out, in := make([][]int, len(v.nodes)), make([][]int, len(v.nodes))
	for i, l := range v.links {
		out[l.Source] = append(out[l.Source], i)
		in[l.Target] = append(in[l.Target], i)
	}
	last := 0
	for _, i := range v.order {
		for _, li := range out[i] {
			j := v.links[li].Target
			depth[j] = max(depth[j], depth[i]+1)
			last = max(last, depth[j])
		}
	}
	for k := len(v.order) - 1; k >= 0; k-- {
		i := v.order[k]
		for _, li := range out[i] {
			tail[i] = max(tail[i], tail[v.links[li].Target]+1)
		}
	}
	columns := make([][]int, last+1)
	for i := range v.nodes {
		d := depth[i]
		switch v.align {
		case SankeyAlignJustify:
			if len(out[i]) == 0 {
				d = last
			}
		case SankeyAlignRight:
			d = last - tail[i]
		case SankeyAlignCenter:
			if len(in[i]) == 0 && len(out[i]) > 0 {
				d = last
				for _, li := range out[i] {
					d = min(d, max(0, depth[v.links[li].Target]-1))
				}
			}
		}
		columns[d] = append(columns[d], i)
	}
	maxValue := 0.0
	for _, value := range v.values {
		maxValue = max(maxValue, value)
	}
	weights := make([]float64, len(v.nodes))
	for i, value := range v.values {
		if maxValue > 0 {
			weights[i] = value / maxValue
			if v.scale == SankeyValueScaleSqrt {
				weights[i] = math.Sqrt(weights[i])
			}
		}
	}
	gap := v.padding
	maxCount := 1
	for _, col := range columns {
		maxCount = max(maxCount, len(col))
	}
	gap = min(gap, max(0, h/float32(maxCount)*.25))
	zeroH := min(float32(2), max(0, (h-gap*float32(maxCount-1))/float32(maxCount)))
	unit := float64(h)
	for _, col := range columns {
		sum := 0.0
		zeros := 0
		for _, i := range col {
			sum += weights[i]
			if weights[i] == 0 {
				zeros++
			}
		}
		if sum > 0 {
			unit = min(unit, float64(max(0, h-gap*float32(max(0, len(col)-1))-float32(zeros)*zeroH))/sum)
		}
	}
	var minimumHeights []float64
	minimum := float64(0)
	if v.minLinkWidth > 0 {
		minimumHeights, unit, minimum = v.minimumFlowLayout(weights, unit, columns, in, out, h, gap, zeroH)
	}
	nw := min(v.nodeWidth, max(0, w))
	for d, col := range columns {
		total := gap * float32(max(0, len(col)-1))
		for _, i := range col {
			height := float32(weights[i] * unit)
			if weights[i] == 0 {
				height = zeroH
			}
			if minimumHeights != nil {
				height = float32(minimumHeights[i])
			}
			x := float32(0)
			if last > 0 {
				x = float32(d) * (w - nw) / float32(last)
			}
			g.nodes[i] = sankeyRect{x, 0, nw, height}
			total += height
		}
		y := max(0, (h-total)/2)
		for _, i := range col {
			r := &g.nodes[i]
			r.y = y
			y += r.h + gap
		}
	}
	// Alternate weighted neighbor relaxation with collision resolution in input order.
	for pass := 0; pass < v.iterations; pass++ {
		for step := 0; step < len(columns); step++ {
			d := step
			if pass%2 == 1 {
				d = len(columns) - 1 - step
			}
			col := columns[d]
			for _, i := range col {
				edges := in[i]
				if pass%2 == 1 {
					edges = out[i]
				}
				sum, weighted := 0.0, 0.0
				for _, li := range edges {
					l := v.links[li]
					j := l.Source
					if pass%2 == 1 {
						j = l.Target
					}
					weight := 0.0
					if maxValue > 0 {
						weight = l.Value / maxValue
					}
					sum += weight
					weighted += weight * float64(g.nodes[j].y+g.nodes[j].h/2)
				}
				if sum > 0 {
					r := &g.nodes[i]
					r.y += (float32(weighted/sum) - r.h/2 - r.y) * .5
				}
			}
			bottom := float32(0)
			for _, i := range col {
				r := &g.nodes[i]
				r.y = max(r.y, bottom)
				bottom = r.y + r.h + gap
			}
			if len(col) > 0 && bottom-gap > h {
				bottom = h
				for k := len(col) - 1; k >= 0; k-- {
					r := &g.nodes[col[k]]
					r.y = min(r.y, bottom-r.h)
					bottom = r.y - gap
				}
			}
		}
	}
	sourceY, targetY := make([]float32, len(v.nodes)), make([]float32, len(v.nodes))
	for i, l := range v.links {
		a, b := g.nodes[l.Source], g.nodes[l.Target]
		ah, bh := float32(0), float32(0)
		if v.values[l.Source] > 0 {
			ah = a.h * float32(l.Value/v.values[l.Source])
		}
		if v.values[l.Target] > 0 {
			bh = b.h * float32(l.Value/v.values[l.Target])
		}
		if minimumHeights != nil && l.Value > 0 {
			ah = float32(max(minimum, weights[l.Source]*unit*(l.Value/v.values[l.Source])))
			bh = float32(max(minimum, weights[l.Target]*unit*(l.Value/v.values[l.Target])))
		}
		g.links[i] = [4]float32{a.y + sourceY[l.Source], ah, b.y + targetY[l.Target], bh}
		sourceY[l.Source] += ah
		targetY[l.Target] += bh
	}
	return g
}
func (v *SankeyChartView) Render(cx *el.Context) el.Element {
	text := locale.Current()
	box := surface().Role("figure").Name(v.title).Value(strconv.Itoa(len(v.nodes)) + " nodes, " + strconv.Itoa(len(v.links)) + " links").Disabled(v.disabled).P(theme.SpaceLg).Gap(theme.SpaceMd).Items(el.Stretch)
	toggle := text.ShowTable
	if v.table {
		toggle = text.ShowChart
	}
	box.Child(el.Div().Row().Items(el.Center).Child(el.Text(v.title).Bold(), el.Div().Grow(), Button(toggle, func() { v.table = !v.table }).Variant(ButtonGhost).Size(24).Render(cx)))
	if v.err != nil {
		return box.Child(el.Text(v.err.Error()).TextColor(theme.Danger))
	}
	if v.table {
		if v.tbl == nil {
			v.tbl = Table(Col("Source"), Col("Target"), Col(text.ChartValue).Numeric()).Height(v.height)
		}
		rows := make([][]string, len(v.links))
		for i, l := range v.links {
			rows[i] = []string{v.nodes[l.Source].Name, v.nodes[l.Target].Name, v.format(l.Value)}
		}
		v.tbl.SetRows(rows)
		return box.Child(v.tbl.Render(cx))
	}
	if len(v.nodes) == 0 {
		return box.Child(el.Div().H(el.Dp(v.height)).Center().Child(el.Text(text.NoData)))
	}
	w := v.width
	if w <= 0 {
		w, _ = cx.ViewportSize()
		w = max(1, w-2*theme.SpaceLg)
	}
	g := v.layout(w, v.height)
	plot := el.Div().WFull().H(el.Dp(v.height)).Child(el.Widget(core.Func(v.draw)).WFull().H(el.Dp(v.height)))
	var labels sankeyLabelPlacement
	plot.Decorate(func(gtx core.C, draw func()) {
		origin, _ := cx.PaintGeometry()
		labels.reset(image.Rectangle{Min: origin, Max: origin.Add(gtx.Constraints.Max)}, gtx.Dp(2))
		draw()
	})
	order := make([]int, len(g.nodes))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		if a == b {
			return 0
		}
		if a == v.hover {
			return -1
		}
		if b == v.hover {
			return 1
		}
		if v.values[a] > v.values[b] {
			return -1
		}
		if v.values[a] < v.values[b] {
			return 1
		}
		return 0
	})
	for _, i := range order {
		r := g.nodes[i]
		data := v.datum(i)
		var label el.Element = el.Div().Child(el.Text(v.format(data.Value)).TextSize(theme.TextXs).MaxLines(1), el.Text(data.Node.Name).TextSize(theme.TextXs).TextColor(theme.Muted).MaxLines(1))
		if v.labels != nil {
			if c := v.labels(cx, data); c != nil {
				label = c
			}
		}
		left := r.x + r.w + 4
		labelW := min(float32(100), max(0, w-left))
		if r.x > w/2 {
			left = max(0, r.x-104)
			labelW = max(0, r.x-left-4)
		}
		box := el.Div().Absolute().Left(left).Top(max(0, min(v.height-28, r.y+r.h/2-14))).W(el.Dp(labelW)).Child(label)
		box.Decorate(func(gtx core.C, draw func()) {
			origin, _ := cx.PaintGeometry()
			if labels.take(image.Rectangle{Min: origin, Max: origin.Add(gtx.Constraints.Max)}) {
				draw()
			}
		})
		plot.Child(box)
	}
	if v.hover >= 0 && v.hover < len(v.nodes) {
		data := v.datum(v.hover)
		var content el.Element = el.Text(data.Node.Name + "  " + v.format(data.Value)).TextSize(theme.TextSm)
		if v.tooltip != nil {
			content = v.tooltip(cx, data)
		}
		plot.Child(el.Div().Absolute().Top(4).Left(4).MaxW(el.Full).P(theme.SpaceSm).Bg(theme.Surface).Border(1, theme.Border).Child(content))
	}
	return box.Child(plot)
}
func (v *SankeyChartView) draw(gtx core.C) core.D {
	size := gtx.Constraints.Max
	px := gtx.Metric.PxPerDp
	if px <= 0 {
		px = 1
	}
	w, h := float32(size.X)/px, float32(size.Y)/px
	g := v.layout(w, h)
	if gtx.Enabled() && v.width != w {
		v.width = w
		gtx.Execute(op.InvalidateCmd{})
	}
	if gtx.Enabled() {
		for {
			ev, ok := gtx.Event(pointer.Filter{Target: &v.tag, Kinds: pointer.Move | pointer.Enter | pointer.Leave})
			if !ok {
				break
			}
			e := ev.(pointer.Event)
			next := -1
			if e.Kind != pointer.Leave {
				x, y := e.Position.X/px, e.Position.Y/px
				for i, r := range g.nodes {
					if x >= r.x && x <= r.x+r.w && y >= r.y && y <= r.y+r.h {
						next = i
						break
					}
				}
			}
			if next != v.hover {
				v.hover = next
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
	v.hoverMotion.update(gtx, len(v.nodes), v.hover, !v.noHoverAnimation && !v.disabled)
	hoverTotal := v.hoverMotion.total()
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	defer area.Pop()
	event.Op(gtx.Ops, &v.tag)
	for i, l := range v.links {
		if l.Value == 0 {
			continue
		}
		a, b := g.nodes[l.Source], g.nodes[l.Target]
		band := g.links[i]
		var p clip.Path
		p.Begin(gtx.Ops)
		// Two smooth boundaries permit different endpoint widths under sqrt scaling.
		x0, x1 := (a.x+a.w)*px, b.x*px
		for step := 0; step <= 24; step++ {
			t := float32(step) / 24
			s := t * t * (3 - 2*t)
			pt := f32.Pt(x0+(x1-x0)*t, (band[0]+(band[2]-band[0])*s)*px)
			if step == 0 {
				p.MoveTo(pt)
			} else {
				p.LineTo(pt)
			}
		}
		for step := 24; step >= 0; step-- {
			t := float32(step) / 24
			s := t * t * (3 - 2*t)
			p.LineTo(f32.Pt(x0+(x1-x0)*t, (band[0]+band[1]+(band[2]+band[3]-band[0]-band[1])*s)*px))
		}
		p.Close()
		outline := clip.Outline{Path: p.End()}.Op().Push(gtx.Ops)
		aColor, bColor := v.nodeColor(l.Source), v.nodeColor(l.Target)
		alpha := v.opacity
		unrelated := max(0, hoverTotal-v.hoverMotion.weight(l.Source)-v.hoverMotion.weight(l.Target))
		alpha *= 1 - .8*unrelated
		aColor.A = uint8(float32(aColor.A) * alpha)
		bColor.A = uint8(float32(bColor.A) * alpha)
		paint.LinearGradientOp{Stop1: f32.Pt(x0, 0), Stop2: f32.Pt(x1, 0), Color1: aColor, Color2: bColor}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		outline.Pop()
	}
	for i, r := range g.nodes {
		rect := image.Rect(int(r.x*px), int(r.y*px), int((r.x+r.w)*px), int((r.y+r.h)*px))
		radius := int(min(v.radius, r.w/2, r.h/2) * px)
		paint.FillShape(gtx.Ops, v.nodeColor(i), clip.UniformRRect(rect, radius).Op(gtx.Ops))
		area := clip.Rect(rect).Push(gtx.Ops)
		core.Role("img", v.format(v.values[i])).Add(gtx.Ops)
		semantic.LabelOp(v.nodes[i].Name).Add(gtx.Ops)
		area.Pop()
	}
	return core.D{Size: size}
}
