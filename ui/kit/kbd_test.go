package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestKbdInheritsTextSizeAndPreservesPlainBounds(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Kbd("mod+shift+p")
		size := float32(12)
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().TextSize(size).Child(v.Render(cx))
		}), 300, scale)
		small := bounds(h, "mod+shift+p")
		size = 24
		h.Frame()
		large := bounds(h, "mod+shift+p")
		if large.Dx() <= small.Dx() || large.Dy() <= small.Dy() {
			t.Fatalf("font size not inherited: %v %v", small, large)
		}
		v.Plain()
		h.Frame()
		if got := bounds(h, "mod+shift+p"); got != large {
			t.Fatalf("Plain moved keycap: %v %v", large, got)
		}
	}
}

func TestKbdNarrowAndLiteralLabels(t *testing.T) {
	for _, scale := range []int{1, 2} {
		for _, label := range []string{"中文 English 123 very long label", "ctrl+", "mod+shift+p", ""} {
			h := renderView(Kbd(label), 80, scale)
			if label == "" {
				continue
			}
			n, ok := node(h, label)
			if !ok || n.Desc.Bounds.Dx() > 80*scale || n.Desc.Bounds.Dy() <= 0 {
				t.Fatalf("bad keycap: %+v", n)
			}
		}
	}
}
