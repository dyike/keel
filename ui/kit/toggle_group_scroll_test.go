package kit

import (
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
)

func TestToggleGroupNarrowScrollAndResize(t *testing.T) {
	v := ToggleGroup("Align left", "Center", "Align right").Segmented(true).Size(ToggleSizeLarge)
	v.SetValue("Center")
	var cx *el.Context
	width := float32(180)
	h := render(func(c *el.Context) el.Element {
		cx = c
		return el.Div().W(el.Dp(width)).Child(v.Render(c))
	})
	id := autoID("togglegroup", v)
	_, viewport, content := cx.ScrollStateX(id)
	if content <= viewport {
		t.Fatalf("labels squeezed instead of overflowing: %g / %g", content, viewport)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(50, 20), Scroll: f32.Pt(1000, 0)})
	h.Frame()
	h.Frame()
	b := bounds(h, "Align right")
	if b.Empty() || b.Min.X < 0 || b.Max.X > int(width) {
		t.Fatalf("last option unreachable: %v", b)
	}
	click(t, h, "Align right")
	if got := v.Value(); len(got) != 1 || got[0] != "Align right" {
		t.Fatal("scrolled selection", got)
	}
	width = 600
	h.Frame()
	h.Frame()
	if offset, _, _ := cx.ScrollStateX(id); offset != 0 {
		t.Fatal("resize retained stale scroll", offset)
	}
	if v.Value()[0] != "Align right" {
		t.Fatal("resize lost selection")
	}
}

func TestToggleGroupTouchScrollDoesNotSelect(t *testing.T) {
	v := ToggleGroup("Align left", "Center", "Align right").Segmented(true).Size(ToggleSizeLarge)
	v.SetValue("Center")
	var cx *el.Context
	h := render(func(c *el.Context) el.Element {
		cx = c
		return el.Div().W(el.Dp(180)).Child(v.Render(c))
	})
	for _, event := range []pointer.Event{
		{Kind: pointer.Press, Position: f32.Pt(160, 20)},
		{Kind: pointer.Move, Position: f32.Pt(80, 20)},
		{Kind: pointer.Move, Position: f32.Pt(10, 20)},
		{Kind: pointer.Release, Position: f32.Pt(10, 20)},
	} {
		event.Source, event.PointerID = pointer.Touch, 1
		h.Router.Queue(event)
		h.Frame()
	}
	h.Frame()
	if offset, _, _ := cx.ScrollStateX(autoID("togglegroup", v)); offset <= 0 {
		t.Fatal("touch drag did not scroll", offset)
	}
	if got := v.Value(); len(got) != 1 || got[0] != "Center" {
		t.Fatal("touch scrolling changed selection", got)
	}
}
