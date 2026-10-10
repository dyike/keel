package markdown

import (
	"image"
	"strings"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/el"
)

func layoutDoc(src string, width int) image.Point {
	d := New(src)
	root := el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return d.Render(cx) }))
	var ops op.Ops
	return root.Layout(layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(width, 1<<24)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}).Size
}

// A table is as wide as its content, like a web page's, and its cells share
// one width per column. It used to fill the document (or, beside short text,
// shrink to 1dp cells that wrapped every character).
func TestTableFitsItsContent(t *testing.T) {
	const small = "| 城市 | 天气 | 温度 | 降水概率 |\n|------|------|------|----------|\n| **上海** | 阴 | 18~26°C | 4% |"
	alone := layoutDoc(small, 888)
	if alone.X >= 600 || alone.Y > 120 {
		t.Fatalf("a small table measured %v: it should hug its content on two lines", alone)
	}
	if beside := layoutDoc("intro\n\n"+small+"\n\n- a\n- b", 888); beside.Y > 260 {
		t.Fatalf("beside short text the table wrapped: %v", beside)
	}
	// Wider than the document: it fits the width and its cells wrap.
	wide := "| A | B |\n|---|---|\n| " + strings.Repeat("长内容 ", 60) + " | " + strings.Repeat("更长的内容 ", 60) + " |"
	if size := layoutDoc(wide, 500); size.X > 500 || size.Y < 100 {
		t.Fatalf("a wide table measured %v in 500", size)
	}
}
