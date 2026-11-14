package kit

import (
	"testing"

	"github.com/dyike/keel/ui/locale"
)

func TestSankeyNarrowHeaderTableToggle(t *testing.T) {
	v := SankeyChart([]SankeyNode{{Name: "Revenue"}, {Name: "Cost"}}, []SankeyLink{{Source: 0, Target: 1, Value: 55}}).Title("Revenue flow")
	h := renderView(v, 240, 1)
	b := bounds(h, locale.Current().ShowTable)
	if b.Empty() || b.Min.X < 0 || b.Max.X > 240 {
		t.Fatalf("table toggle outside narrow chart: %v", b)
	}
	click(t, h, locale.Current().ShowTable)
	if !v.table {
		t.Fatal("narrow chart table toggle did not respond")
	}
	click(t, h, locale.Current().ShowChart)
	if v.table {
		t.Fatal("could not return to chart")
	}
}
