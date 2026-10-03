package kit

import (
	"image/color"
	"math"
	"testing"
)

func TestBadgeModesSizesAndChildLayout(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		button := Button("Inbox", func() { calls++ })
		v := Badge(120).Max(9).Child(button)
		h := renderView(v, 220, scale)
		child := bounds(h, "Inbox")
		for _, size := range []float32{12, 18, 24, 32} {
			v.Size(size)
			h.Frame()
			n, ok := semanticNode(h, "badge:9+")
			if !ok || n.Desc.Label != "120" || n.Desc.Bounds.Dy() != int(size)*scale || bounds(h, "Inbox") != child {
				t.Fatal("count size or child layout", n.Desc.Bounds)
			}
			v.Icon(IconCheck).Name("Verified")
			v.SetValue(0)
			h.Frame()
			n, ok = semanticNode(h, "badge:icon")
			if !ok || n.Desc.Label != "Verified" || n.Desc.Bounds.Dy() != int(size)*scale || n.Desc.Bounds.Min.Y < child.Min.Y || bounds(h, "Inbox") != child {
				t.Fatal("icon layout", n.Desc.Bounds, child)
			}
			click(t, h, "Inbox")
			v.Icon(IconNone).Name("")
			v.SetValue(120)
		}
		if calls != 4 {
			t.Fatal("badge blocked child")
		}
		v.Dot().SetValue(1)
		h.Frame()
		if _, ok := semanticNode(h, "badge:dot"); !ok {
			t.Fatal("dot mode")
		}
		v.SetValue(0)
		h.Frame()
		if _, ok := semanticNode(h, "badge:dot"); ok || bounds(h, "Inbox") != child {
			t.Fatal("hidden badge affected child")
		}
		v.Size(float32(math.NaN())).Size(-1).Size(float32(math.Inf(1)))
		if v.diameter() != 32 {
			t.Fatal("invalid size")
		}
		v.Color(color.NRGBA{R: 255, A: 255}).Tone(ToneSuccess)
		if v.color != nil {
			t.Fatal("theme colors did not restore")
		}
	}
}
