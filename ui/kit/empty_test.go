package kit

import "testing"

func TestEmptyContentAndConstraints(t *testing.T) {
	v := Empty("暂无结果 0", "尝试其他关键词 Search again，或者调整筛选条件。")
	for _, scale := range []int{1, 2} {
		h := renderView(v, 140, scale)
		n, ok := semanticNode(h, "empty")
		if !ok || n.Desc.Label != "暂无结果 0" || n.Desc.Bounds.Dx() > 140*scale {
			t.Fatalf("invalid empty state: %+v", n)
		}
		v.SetTitle("无数据")
		v.SetDescription("")
		h.Frame()
		if n, ok = semanticNode(h, "empty"); !ok || n.Desc.Label != "无数据" {
			t.Fatal("stale empty state")
		}
		v.SetTitle("暂无结果 0")
	}
}
