package kit

import (
	"image"
	"testing"
)

func TestAvatarInitialsAndStableSize(t *testing.T) {
	for name, want := range map[string]string{"张三": "张", "Ada Lovelace": "AL", "Ada Byron Lovelace": "AB", "9 lives": "9L", "  ": "?"} {
		if got := avatarInitials(name); got != want {
			t.Fatalf("%q: %q != %q", name, got, want)
		}
	}
	for _, scale := range []int{1, 2} {
		a := Avatar("Ada Lovelace")
		h := renderView(a, 100, scale)
		before, ok := semanticNode(h, "avatar")
		if !ok {
			t.Fatal("missing initials")
		}
		a.Image(image.NewNRGBA(image.Rect(0, 0, 80, 40)))
		h.Frame()
		after, ok := semanticNode(h, "avatar")
		if !ok || before.Desc.Bounds != after.Desc.Bounds {
			t.Fatal("image changed avatar dimensions")
		}
		a.Image(nil)
		a.SetName("张三")
		h.Frame()
		after, ok = semanticNode(h, "avatar")
		if !ok || after.Desc.Label != "张三" || before.Desc.Bounds != after.Desc.Bounds {
			t.Fatal("fallback changed dimensions")
		}
	}
}

func TestAvatarStatusDoesNotChangeBounds(t *testing.T) {
	a := Avatar("张三").Size(Small)
	h := renderView(a, 100, 1)
	before := bounds(h, "张三")
	for _, status := range []AvatarStatus{Online, Busy, Offline, ""} {
		a.Status(status)
		h.Frame()
		n, ok := node(h, "张三")
		if !ok || n.Desc.Bounds != before || n.Desc.Description != "avatar"+func() string {
			if status != "" {
				return ":" + string(status)
			}
			return ""
		}() {
			t.Fatalf("status changed bounds/semantics: %+v", n)
		}
	}
}
