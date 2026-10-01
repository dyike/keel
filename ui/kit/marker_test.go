package kit

import "testing"

func TestMarkerToneDoesNotChangeBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Marker("连接状态 Connected 123")
		h := renderView(v, 130, scale)
		before, ok := semanticNode(h, "marker:info")
		if !ok || before.Desc.Bounds.Dx() > 130*scale {
			t.Fatal("invalid marker")
		}
		v.Tone(Warning)
		h.Frame()
		after, ok := semanticNode(h, "marker:warning")
		if !ok || before.Desc.Bounds != after.Desc.Bounds {
			t.Fatal("tone changed marker bounds")
		}
	}
}
