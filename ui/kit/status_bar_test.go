package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestStatusBarWrapsAndUpdates(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := StatusBar().Left(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("连接正常 Connected 123") })).Right(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("第 100 行，共 999 行") }))
		h := renderView(v, 160, scale)
		n, ok := semanticNode(h, "status")
		if !ok || n.Desc.Bounds.Dx() > 160*scale {
			t.Fatalf("invalid status bounds: ok=%v node=%+v", ok, n)
		}
		v.Left(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("离线") }))
		v.Right(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("重连中") }))
		h.Frame()
		if n, ok = node(h, "离线"); !ok || n.Desc.Label != "离线" {
			t.Fatal("stale status")
		}
	}
}

func TestStatusBarFixedHeight(t *testing.T) {
	for _, scale := range []int{1, 2} {
		h := renderView(StatusBar().Left(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("很长很长的左侧文字需要截断") })).Right(el.ViewFunc(func(cx *el.Context) el.Element { return el.Text("右侧") })), 160, scale)
		n, ok := semanticNode(h, "status")
		if !ok || n.Desc.Bounds.Dy() != 24*scale {
			t.Fatalf("status bar height %v", n.Desc.Bounds)
		}
	}
}
