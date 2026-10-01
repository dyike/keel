package kit

import "testing"

func TestStatusBarWrapsAndUpdates(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := StatusBar("连接正常 Connected 123", "第 100 行，共 999 行")
		h := renderView(v, 160, scale)
		n, ok := semanticNode(h, "status:第 100 行，共 999 行")
		if !ok || n.Desc.Bounds.Dx() > 160*scale {
			t.Fatalf("invalid status bounds: ok=%v node=%+v", ok, n)
		}
		v.SetStatus("离线")
		v.SetDetail("重连中")
		h.Frame()
		if n, ok = semanticNode(h, "status:重连中"); !ok || n.Desc.Label != "离线" {
			t.Fatal("stale status")
		}
	}
}
