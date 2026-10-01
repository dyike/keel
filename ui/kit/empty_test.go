package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestEmptyContentAndConstraints(t *testing.T) {
	v := Empty("暂无结果 0").Description("尝试其他关键词 Search again，或者调整筛选条件。")
	for _, scale := range []int{1, 2} {
		h := renderView(v, 140, scale)
		n, ok := node(h, "暂无结果 0")
		if !ok || n.Desc.Label != "暂无结果 0" || n.Desc.Bounds.Dx() > 140*scale {
			t.Fatalf("invalid empty state: %+v", n)
		}
		v.SetTitle("无数据")
		v.SetDescription("")
		h.Frame()
		if n, ok = node(h, "无数据"); !ok || n.Desc.Label != "无数据" {
			t.Fatal("stale empty state")
		}
		v.SetTitle("暂无结果 0")
	}
}

func TestEmptyAction(t *testing.T) {
	calls := 0
	v := Empty("空").Action(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Name("创建").OnClick(func() { calls++ }).Child(el.Text("创建"))
	}))
	h := renderView(v, 240, 1)
	click(t, h, "创建")
	if calls != 1 {
		t.Fatal("action failed")
	}
}
