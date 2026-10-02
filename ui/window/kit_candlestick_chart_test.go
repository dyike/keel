package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitCandlestickChartAgentTable(t *testing.T) {
	c := kit.CandlestickChart(kit.Candle{Label: "Mon", Open: 10, High: 14, Low: 8, Close: 12}).Title("OHLC")
	w := openTest(t, kitPage(c))
	if roleOfName(w, "OHLC") != "figure" {
		t.Fatal("figure missing")
	}
	w.click(element(t, w, "查看数据表").center())
	if roleOfName(w, "Mon | 10 | 14 | 8 | 12") != "row" {
		t.Fatal("OHLC row missing")
	}
}
