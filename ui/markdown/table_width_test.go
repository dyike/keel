package markdown

import (
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
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
	// Wider than the document: its viewport fits and columns scroll without wrapping.
	wide := "| A | B |\n|---|---|\n| " + strings.Repeat("长内容 ", 60) + " | " + strings.Repeat("更长的内容 ", 60) + " |"
	if size := layoutDoc(wide, 500); size.X > 500 || size.Y > 100 {
		t.Fatalf("a wide table measured %v in 500", size)
	}
}

func TestTableColumnsFollowFontScale(t *testing.T) {
	const source = "| 项目名称 | 运行状态 | 详细说明 |\n|---|---|---|\n| 示例任务 | 执行成功 | 无需换行 |"
	doc := New(source)
	root := el.Embed(doc)
	var ops op.Ops
	var initial int
	for _, size := range []float32{1, .85, 1.5, 2, 1} {
		ops.Reset()
		root.Layout(layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(350, 500)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: size}})
		width := doc.selection.parts[0].rect.Dx()
		if initial == 0 {
			initial = width
		}
		if size > 1 && width <= initial {
			t.Fatalf("font scale %v did not expand column: %d <= %d", size, width, initial)
		}
		for _, p := range doc.selection.parts {
			if len(p.r.rt.pieces) == 0 {
				continue
			}
			firstY := p.r.rt.pieces[0].rect.Min.Y
			for _, piece := range p.r.rt.pieces {
				if piece.rect.Min.Y != firstY {
					t.Fatalf("cell %q wrapped at scale %v", p.r.plain, size)
				}
			}
		}
	}
}

func TestWideTableSelectionAfterHorizontalScroll(t *testing.T) {
	d := New("| 第一列 | 第二列 | 第三列 | 第四列 | 第五列 |\n|---|---|---|---|---|\n| 长列内容 | 长列内容 | 长列内容 | 长列内容 | 长列内容 |")
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(220)).Child(d.Render(cx)) }))
	h := uitest.New(root)
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(100, 20), Scroll: f32.Pt(65, 0)})
	h.Frame()
	h.Frame()
	st := d.chunks[0].blocks[0].view.(*tableView)
	if st.scrollX <= 0 {
		t.Fatal("wide table did not scroll")
	}
	x0, y0 := textPoint(t, h, d, "第二列", 0)
	x1, y1 := textPoint(t, h, d, "第二列", 3)
	h.Drag(x0, y0, x1-1, y1)
	h.Frame()
	assertSelectionCopy(t, h, d, "第二列")
}
