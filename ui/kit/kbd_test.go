package kit

import (
	"github.com/dyike/keel/ui/core"
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

func TestKbdExplicitSizeStyleAndBinding(t *testing.T) {
	defer core.Bind("test.kbd.size")
	for _, scale := range []int{1, 2} {
		core.Bind("test.kbd.size", "ctrl+s")
		v := KbdFor("test.kbd.size")
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().TextSize(24).Child(v.Render(cx)) }), 200, scale)
		inherited := bounds(h, "ctrl+s")
		v.Size(12)
		h.Frame()
		small := bounds(h, "ctrl+s")
		if small.Dy() >= inherited.Dy() {
			t.Fatal("explicit size did not override parent")
		}
		v.Size(24)
		h.Frame()
		large := bounds(h, "ctrl+s")
		if large.Dy() <= small.Dy() {
			t.Fatal("explicit size not applied")
		}
		v.Plain()
		h.Frame()
		if bounds(h, "ctrl+s") != large {
			t.Fatal("plain changed size")
		}
		v.Style(func(e *el.TextEl) { e.P(20) })
		h.Frame()
		if bounds(h, "ctrl+s").Dy() <= large.Dy() {
			t.Fatal("style not applied")
		}
		v.Style(nil).Size(0)
		h.Frame()
		if bounds(h, "ctrl+s") != inherited {
			t.Fatal("defaults not restored")
		}
		v.Size(16)
		core.Bind("test.kbd.size", "ctrl+k")
		h.Frame()
		if shown(h, "ctrl+s") || !shown(h, "ctrl+k") {
			t.Fatal("rebind failed")
		}
		core.Bind("test.kbd.size")
		h.Frame()
		if shown(h, "ctrl+k") {
			t.Fatal("unbound keycap remained")
		}
	}
}
