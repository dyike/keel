package kit

import (
	"reflect"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func dockPoint(r dockRect) (float32, float32) { return r.x + r.w/2, r.y + r.h/2 }
func dockPress(h *uitest.Harness, x, y float32) {
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
}
func dockMove(h *uitest.Harness, x, y float32) {
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
}
func dockRelease(h *uitest.Harness, x, y float32) {
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x, y)})
	h.Frame()
	h.Frame()
}
func TestDockDragReorderJoinSplitAndFocus(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := nestedDock()
		calls := 0
		v.OnLayoutChange(func(DockLayout) { calls++ })
		var cx *el.Context
		h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element {
			cx = ctx
			return el.Div().W(el.Dp(900)).P(13).H(el.Dp(600)).Child(v.Render(ctx))
		}), 900, scale)
		drag := func(id string, x, y float32) {
			a, b := dockPoint(v.tabRects[id])
			dockPress(h, a*float32(scale), b*float32(scale))
			dockMove(h, x*float32(scale), y*float32(scale))
			dockRelease(h, x*float32(scale), y*float32(scale))
		}
		b := v.tabRects["b"]
		drag("a", b.x+b.w-2, b.y+b.h/2)
		if !reflect.DeepEqual(v.Layout().Left, []string{"b", "a"}) || calls != 1 {
			t.Fatalf("reorder scale %d: %v callbacks=%d", scale, v.Layout().Left, calls)
		}
		c := v.groupRects[v.layout.RightTree]
		drag("a", c.x+c.w/2, c.y+c.h/2)
		if !reflect.DeepEqual(v.Layout().Right, []string{"c", "a"}) || v.layout.RightTree.Active != "a" || calls != 2 || !cx.Focused(v.tabID("a")) {
			t.Fatalf("join scale %d: %v callbacks=%d focus=%v", scale, v.Layout().Right, calls, cx.Focused(v.tabID("a")))
		}
		left := v.groupRects[v.layout.LeftTree]
		a, y := dockPoint(v.tabRects["a"])
		dockPress(h, a*float32(scale), y*float32(scale))
		dockMove(h, (left.x+left.w/2)*float32(scale), (left.y+left.h-10)*float32(scale))
		if v.drag.drop.kind != 2 || v.drag.drop.placement != DockPlacementBottom || v.drag.drop.preview.h != left.h/2 {
			t.Fatalf("split preview: %+v", v.drag)
		}
		dockRelease(h, (left.x+left.w/2)*float32(scale), (left.y+left.h-10)*float32(scale))
		if v.layout.LeftTree.First == nil || v.layout.LeftTree.Second.Active != "a" || calls != 3 {
			t.Fatal("edge split")
		}
	}
}
func TestDockDragCancelOutsideAndDisabled(t *testing.T) {
	for _, mode := range []string{"cancel", "escape", "outside", "self-disabled", "ancestor-disabled"} {
		v := nestedDock()
		calls := 0
		v.OnLayoutChange(func(DockLayout) { calls++ })
		disabled := false
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(900)).H(el.Dp(600)).Disabled(disabled).Child(v.Render(cx))
		}), 900, 1)
		before := v.Layout()
		x, y := dockPoint(v.tabRects["a"])
		c := v.groupRects[v.layout.RightTree]
		dockPress(h, x, y)
		dockMove(h, c.x+c.w/2, c.y+c.h/2)
		if v.drag.drop.kind == 0 {
			t.Fatal("no preview")
		}
		switch mode {
		case "escape":
			h.Key(key.NameEscape, 0)
			dockRelease(h, c.x+c.w/2, c.y+c.h/2)
		case "cancel":
			h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
			h.Frame()
		case "outside":
			dockMove(h, 1000, 700)
			dockRelease(h, 1000, 700)
		case "self-disabled":
			v.SetDisabled(true)
			h.Frame()
			dockRelease(h, c.x, c.y)
		case "ancestor-disabled":
			disabled = true
			h.Frame()
			h.Frame()
			dockRelease(h, c.x, c.y)
		}
		if !reflect.DeepEqual(before, v.Layout()) || calls != 0 || v.drag.id != "" {
			t.Fatalf("%s changed layout or kept drag: calls=%d", mode, calls)
		}
	}
}
func TestDockDragToEmptyRegion(t *testing.T) {
	v := nestedDock()
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(900)).H(el.Dp(600)).Child(v.Render(cx)) }), 900, 1)
	x, y := dockPoint(v.tabRects["b"])
	target := v.centerRect
	dockPress(h, x, y)
	dockMove(h, target.x+target.w/2, target.y+target.h-5)
	if v.drag.drop.kind != 3 || v.drag.drop.side != DockBottom {
		t.Fatalf("outer target %+v", v.drag.drop)
	}
	dockRelease(h, target.x+target.w/2, target.y+target.h-5)
	if v.where("b") != int(DockBottom) || v.layout.BottomTree.Active != "b" {
		t.Fatal("empty region not created")
	}
}
