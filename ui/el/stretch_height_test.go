package el

import (
	"image"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
)

// A content-sized column measures a table-like child at its natural width
// (cells growing from a 1dp basis wrap per character, very tall), then
// stretches it to the column's width, where it is short. The column must
// take the stretched height: keeping the first one left a large blank gap.
func TestStretchedChildUpdatesAutoHeight(t *testing.T) {
	cell := func(s string) Element { return Div().Grow().W(Dp(1)).Px(10).Py(8).Child(Text(s)) }
	height := func(fullWidth bool) int {
		root := Embed(ViewFunc(func(cx *Context) Element {
			grid := Div()
			if fullWidth {
				grid.WFull() // definite from the start: the reference
			}
			for range 3 {
				grid.Child(Div().Row().Child(cell("城市名称"), cell("天气"), cell("温度"), cell("降水概率")))
			}
			text := Text("明天上海和北京都是阴天，基本不下雨，比今天上海的毛毛雨好很多")
			col := Div().Gap(12).Child(text, grid)
			if fullWidth {
				col.WFull()
			}
			return col
		}))
		var ops op.Ops
		return root.Layout(layout.Context{Ops: &ops, Constraints: layout.Constraints{Max: image.Pt(888, 1<<24)}, Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}}).Size.Y
	}
	if got, want := height(false), height(true); got != want {
		t.Fatalf("stretched height %d, want %d as when the width is definite", got, want)
	}
}
