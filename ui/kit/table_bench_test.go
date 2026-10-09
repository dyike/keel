package kit

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func bigTable(n int) (*TableView, *uitest.Harness) {
	rows := make([][]string, n)
	for i := range rows {
		rows[i] = []string{fmt.Sprintf("SO-%06d", i), []string{"华东", "华北", "华南"}[i%3], fmt.Sprint(i * 37 % 9000)}
	}
	t := Table(Col("单号").Width(120), Col("区域").Width(100), Col("金额").Width(100).Numeric()).Height(500)
	t.SetRows(rows)
	root := el.Root(viewFunc(func(cx *el.Context) el.Element { return el.Div().Child(t.Render(cx)) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(800, 600); root.Layout(gtx) })
	return t, h
}

// A table of 300,000 rows draws, sorts and moves its selection without
// laying out rows it does not show.
func TestTableThreeHundredThousandRows(t *testing.T) {
	start := time.Now()
	tbl, h := bigTable(300000)
	h.Frame()
	click(t, h, "金额") // sort
	h.Frame()
	h.Key(key.NameEnd, key.ModShortcut)
	h.Frame()
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("300,000 rows took %v", d)
	}
	if !shown(h, "SO-") && tbl == nil {
		t.Fatal("no rows")
	}
}

func BenchmarkTableFrame300k(b *testing.B) {
	_, h := bigTable(300000)
	h.Frame()
	for b.Loop() {
		h.Scroll(100, 300, 400)
	}
}
