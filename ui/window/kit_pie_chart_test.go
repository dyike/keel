package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitPieChartAgentTableAndLegend(t *testing.T) {
	p := kit.PieChart(kit.PieSlice{Name: "Online", Value: 3}, kit.PieSlice{Name: "Store", Value: 1}).Title("Revenue")
	w := openTest(t, kitPage(p))
	if roleOfName(w, "Revenue") != "figure" || roleOfName(w, "Store") != "toggle" {
		t.Fatal("missing chart semantics")
	}
	w.click(element(t, w, "Store").center())
	w.click(element(t, w, "查看数据表").center())
	if roleOfName(w, "Online | 3 | 100.0%") != "row" {
		t.Fatal("agent could not inspect renormalized share")
	}
}
