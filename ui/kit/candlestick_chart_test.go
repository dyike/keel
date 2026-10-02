package kit

import (
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestCandlestickDataValidationTableAndHover(t *testing.T) {
	data := []Candle{{"Mon", 10, 14, 8, 12}, {"Tue", 12, 15, 9, 10}, {"bad", 10, 9, 11, 12}}
	c := CandlestickChart(data...).Title("OHLC")
	data[0].High = 999
	if lo, hi := c.chart.span(); lo != 8 || hi != 15 {
		t.Fatal("invalid or aliased range", lo, hi)
	}
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(400)).Child(c.Render(cx)) }), 400, 1)
	h.Move(110, 130)
	h.Frame()
	if c.chart.hover != 0 || !shown(h, "开盘") {
		t.Fatal("candle tooltip", c.chart.hover)
	}
	click(t, h, "查看数据表")
	if !shown(h, "Mon | 10 | 14 | 8 | 12") || !shown(h, "bad | — | — | — | —") {
		t.Fatal("OHLC table")
	}
	c.SetDisabled(true)
	h.Frame()
	click(t, h, "查看图表")
	if !c.chart.table {
		t.Fatal("disabled chart toggled")
	}
	c.SetData(Candle{"nan", math.NaN(), 1, 0, 0})
	if lo, hi := c.chart.span(); lo != 0 || hi != 0 {
		t.Fatal("nonfinite range")
	}
}
func TestCandlestickPixelBucketsRetainOHLC(t *testing.T) {
	data := []Candle{{"a", 2, 4, 1, 3}, {"b", 3, 8, -2, 5}, {"gap", 0, -1, 1, 0}, {"c", 5, 7, 3, 4}}
	got := candleBuckets(data, 1)
	if len(got) != 1 || got[0].Open != 2 || got[0].Close != 4 || got[0].High != 8 || got[0].Low != -2 {
		t.Fatal(got)
	}
	if got := candleBuckets(data, 10); len(got) != 4 || got[2].valid() {
		t.Fatal("gap removed", got)
	}
}
