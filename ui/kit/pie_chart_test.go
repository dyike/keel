package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestPieChartDataLegendHoverAndTable(t *testing.T) {
	data := []PieSlice{{"A", 3}, {"B", 1}, {"bad", math.NaN()}, {"negative", -2}}
	p := PieChart(data...).Title("share").Donut(.5)
	data[0].Value = 99
	f := p.fractions()
	if f[0] != .75 || f[1] != .25 || f[2] != 0 || f[3] != 0 {
		t.Fatal(f)
	}
	if pieHit(0, 0, 100, .5, f) != -1 || pieHit(80, 0, 100, .5, f) != 0 || pieHit(-80, -40, 100, .5, f) != 1 {
		t.Fatal("sector hit")
	}
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(400)).Child(p.Render(cx)) }), 400, 1)
	h.Move(280, 150)
	h.Frame()
	if p.hover != 0 {
		t.Fatal("pointer did not hover sector", p.hover)
	}
	click(t, h, "B")
	if p.fractions()[0] != 1 {
		t.Fatal("legend did not renormalize")
	}
	p.SetDisabled(true)
	h.Frame()
	click(t, h, "B")
	if !p.hidden[1] {
		t.Fatal("disabled legend fired")
	}
	p.SetDisabled(false)
	h.Frame()
	click(t, h, "查看数据表")
	if !shown(h, "A | 3 | 100.0%") || !shown(h, "bad | — | 0.0%") {
		t.Fatal("table values")
	}
	p.SetData(PieSlice{"huge", math.MaxFloat64}, PieSlice{"huge2", math.MaxFloat64})
	if f := p.fractions(); f[0] != .5 || f[1] != .5 {
		t.Fatal("overflow", f)
	}
	p.SetData(PieSlice{"empty", 0})
	p.table = false
	h.Frame()
	if !shown(h, "暂无数据") {
		t.Fatal("empty chart")
	}
}
