package kit

import "testing"

func TestAlertWrapsAndRetainsStatus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		a := Alert("保存 123", "中英文 Mixed description that wraps in a narrow container. 多行提示内容。").Tone(Success)
		h := renderView(a, 160, scale)
		for i := 0; i < 3; i++ {
			h.Frame()
		}
		n, ok := semanticNode(h, "alert:success")
		if !ok || n.Desc.Label != "保存 123" || n.Desc.Bounds.Dx() > 160*scale || n.Desc.Bounds.Dy() <= 40*scale {
			t.Fatalf("invalid alert: %+v", n)
		}
		a.SetTitle("更新")
		a.SetDescription("")
		a.Tone(Warning)
		h.Frame()
		if n, ok = semanticNode(h, "alert:warning"); !ok || n.Desc.Label != "更新" {
			t.Fatal("updated alert not rendered")
		}
	}
}
