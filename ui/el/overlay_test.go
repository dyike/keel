package el

import (
	"image"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/op"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestOverlayPlacement(t *testing.T) {
	for _, tc := range []struct {
		name       string
		side       Side
		x, y, w, h int
		want       image.Point
	}{
		{"bottom", Bottom, 150, 100, 80, 40, image.Pt(150, 124)},
		{"top", Top, 150, 100, 80, 40, image.Pt(150, 56)},
		{"left", Left, 150, 100, 80, 40, image.Pt(66, 100)},
		{"right", Right, 150, 100, 80, 40, image.Pt(194, 100)},
		{"flip_bottom", Bottom, 150, 280, 80, 40, image.Pt(150, 236)},
		{"flip_top", Top, 150, 0, 80, 40, image.Pt(150, 24)},
		{"flip_left", Left, 0, 100, 80, 40, image.Pt(44, 100)},
		{"flip_right", Right, 360, 100, 80, 40, image.Pt(276, 100)},
		{"neither_fits", Bottom, 150, 100, 80, 240, image.Pt(150, 60)},
		{"cross_axis", Bottom, 360, 100, 80, 40, image.Pt(320, 124)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
				cx.Overlay("layer", Anchored("anchor", Div().Name("layer").W(Dp(float32(tc.w))).H(Dp(float32(tc.h)))).Placement(tc.side, Start))
				return Div().Child(Div().Absolute().Left(float32(tc.x)).Top(float32(tc.y)).W(Dp(40)).H(Dp(20)).ID("anchor"))
			})))
			if b := nodeBounds(h, "layer"); b.Min != tc.want || b.Size() != image.Pt(tc.w, tc.h) {
				t.Fatalf("bounds %v want %v", b, tc.want)
			}
		})
	}
}

func TestOverlayAnchorWidthSubmenuAndCurrentFrame(t *testing.T) {
	x := 20
	root := Root(ViewFunc(func(cx *Context) Element {
		cx.Overlay("parent", Anchored("anchor", Div().Name("parent").H(Dp(40)).Child(Div().ID("item").Size(Dp(20)))).MatchAnchorWidth())
		cx.Overlay("child", Anchored("item", Div().Name("child").Size(Dp(30))).Placement(Right, Start))
		return Div().Child(Div().Absolute().Left(float32(x)).Top(30).W(Dp(100)).H(Dp(20)).ID("anchor"))
	}))
	h := uitest.New(root)
	if b := nodeBounds(h, "parent"); b.Dx() != 100 {
		t.Fatalf("anchor width: %v", b)
	}
	if b := nodeBounds(h, "child"); b.Min != image.Pt(44, 54) {
		t.Fatalf("submenu: %v", b)
	}
	x = 100
	h.Frame()
	if b := nodeBounds(h, "parent"); b.Min.X != 100 {
		t.Fatalf("stale anchor: %v", b)
	}
}

func TestOverlayOutsidePress(t *testing.T) {
	for _, modal := range []bool{false, true} {
		t.Run(map[bool]string{false: "pass", true: "modal"}[modal], func(t *testing.T) {
			open := true
			calls, dismiss := 0, 0
			root := Root(ViewFunc(func(cx *Context) Element {
				if open {
					l := Anchored("anchor", Div().Size(Dp(60)).Name("layer")).OnDismiss(func() { dismiss++; open = false })
					if modal {
						l.Modal()
					}
					cx.Overlay("layer", l)
				}
				return Div().Child(Div().ID("anchor").Size(Dp(20)), Div().Absolute().Left(200).Top(200).Size(Dp(60)).Name("under").OnClick(func() { calls++ }))
			}))
			h := uitest.New(root)
			h.Click(220, 220)
			h.Frame()
			want := 1
			if modal {
				want = 0
			}
			if dismiss != 1 || calls != want {
				t.Fatalf("dismiss=%d calls=%d want=%d", dismiss, calls, want)
			}
			if _, ok := labelBounds(h, "layer"); ok {
				t.Fatal("omitted layer remains")
			}
		})
	}
}

func TestOverlayEscapeTopmost(t *testing.T) {
	a, b := true, true
	ca, cb := 0, 0
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		if a {
			cx.Overlay("a", Modal(Div().Size(Dp(100)).Name("a")).OnDismiss(func() { a = false; ca++ }))
		}
		if b {
			cx.Overlay("b", Modal(Div().Size(Dp(50)).Name("b")).OnDismiss(func() { b = false; cb++ }))
		}
		return Div()
	})))
	h.Key(key.NameEscape, 0)
	h.Frame()
	if ca != 0 || cb != 1 {
		t.Fatalf("escape leaked: %d %d", ca, cb)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if ca != 1 || cb != 1 {
		t.Fatalf("second escape: %d %d", ca, cb)
	}
}

func TestOverlayFocusTrapAndRestore(t *testing.T) {
	for _, remove := range []bool{false, true} {
		t.Run(map[bool]string{false: "restore", true: "removed"}[remove], func(t *testing.T) {
			open := false
			trigger := true
			root := Root(ViewFunc(func(cx *Context) Element {
				if open {
					cx.Overlay("dialog", Modal(Div().W(Dp(120)).Child(Div().ID("first").Name("first").OnClick(func() {}).H(Dp(20)), Input().ID("last").Name("last"))).OnDismiss(func() { open = false }))
				}
				return Div().When(trigger, func(d *DivEl) {
					d.Child(Div().ID("trigger").Name("trigger").Size(Dp(40)).OnClick(func() { open = true }))
				})
			}))
			h := uitest.New(root)
			h.Click(10, 10)
			h.Frame()
			focused := func(id string) bool { return (&Context{root: root}).Focused(id) }
			if !focused("first") {
				t.Fatal("first element not focused")
			}
			h.Router.MoveFocus(key.FocusForward)
			h.Frame()
			if !focused("last") {
				t.Fatal("Tab did not reach input")
			}
			h.Router.MoveFocus(key.FocusForward)
			h.Frame()
			if !focused("first") {
				t.Fatal("Tab escaped trap")
			}
			h.Router.MoveFocus(key.FocusBackward)
			h.Frame()
			if !focused("last") {
				t.Fatal("reverse Tab escaped trap")
			}
			if remove {
				trigger = false
			}
			h.Key(key.NameEscape, 0)
			h.Frame()
			if remove {
				if root.focusedTag() != nil {
					t.Fatal("removed trigger retained focus")
				}
			} else if !focused("trigger") {
				t.Fatalf("trigger focus not restored: tag=%#v states=%#v", root.focusedTag(), root.store.states)
			}
		})
	}
}

func TestOverlayAnchorLossAndMeasurement(t *testing.T) {
	show, anchor := true, true
	dismiss := 0
	root := Embed(ViewFunc(func(cx *Context) Element {
		if show {
			cx.Overlay("layer", Anchored("anchor", Div().Name("layer").Size(Dp(40))).OnDismiss(func() { dismiss++ }))
		}
		return Div().Child(Div().ID("anchor").Hidden(!anchor).Size(Dp(20)))
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		// Measurement deliberately omits the layer. It must not close the live one.
		saved := show
		show = false
		for i := 0; i < 3; i++ {
			var ops op.Ops
			m := gtx
			m.Ops = &ops
			m.Source = input.Source{}
			root.Layout(m)
		}
		show = saved
		root.Layout(gtx)
	})
	first := root.layers["layer"]
	h.Frame()
	if root.layers["layer"] != first || !first.active || dismiss != 0 {
		t.Fatal("measurement changed layer lifecycle")
	}
	anchor = false
	h.Frame()
	h.Frame()
	if dismiss != 1 {
		t.Fatalf("missing anchor dismiss count %d", dismiss)
	}
	show = false
	h.Frame()
	if len(root.activeLayers) != 0 || len(root.layers) != 0 {
		t.Fatal("omitted layer retained")
	}
}

func TestHoveredWithoutClickHandler(t *testing.T) {
	disabled := false
	hovered := false
	root := Root(ViewFunc(func(cx *Context) Element {
		hovered = cx.Hovered("anchor")
		return Div().Child(Div().ID("anchor").Disabled(disabled).Size(Dp(50)))
	}))
	h := uitest.New(root)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(20, 20)})
	h.Frame()
	h.Frame()
	if !hovered {
		t.Fatal("plain element not hovered")
	}
	disabled = true
	h.Frame()
	h.Frame()
	if hovered {
		t.Fatal("disabled element hovered")
	}
	disabled = false
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(200, 200)})
	h.Frame()
	h.Frame()
	if hovered {
		t.Fatal("hover survives pointer leave")
	}
}

func TestOverlayExplicitFocusAndNestedRestore(t *testing.T) {
	outer, inner := false, false
	root := Root(ViewFunc(func(cx *Context) Element {
		if outer {
			cx.Overlay("outer", Modal(Div().W(Dp(140)).Child(
				Div().ID("default").Name("default").OnClick(func() {}).H(Dp(20)),
				Div().ID("nested").Name("nested").OnClick(func() { inner = true }).H(Dp(20)),
			)).OnDismiss(func() { outer = false }))
			if !inner {
				cx.Focus("nested")
			}
		}
		if inner {
			cx.Overlay("inner", Modal(Div().W(Dp(100)).Child(Div().ID("inside").Name("inside").OnClick(func() {}).H(Dp(20)))).OnDismiss(func() { inner = false }))
		}
		return Div().Child(Div().ID("trigger").Name("trigger").Size(Dp(40)).OnClick(func() { outer = true }))
	}))
	h := uitest.New(root)
	h.Click(10, 10)
	h.Frame()
	focused := func(id string) bool { return (&Context{root: root}).Focused(id) }
	if !focused("nested") {
		t.Fatal("explicit focus did not override first element")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !focused("inside") {
		t.Fatal("nested trap did not focus child")
	}
	if root.layers["outer"].seed == root.layers["inner"].seed {
		t.Fatal("layers share element state namespace")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !outer || inner || !focused("nested") {
		t.Fatal("nested close failed to restore parent")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if outer || !focused("trigger") {
		t.Fatal("outer close failed to restore original trigger")
	}
}

func TestOverlayNonmodalKeepsFocusAndBlocksContentClicks(t *testing.T) {
	open := false
	calls := 0
	request := true
	root := Root(ViewFunc(func(cx *Context) Element {
		if request {
			cx.Focus("input")
			request = false
		}
		if open {
			cx.Overlay("layer", Anchored("anchor", Div().Name("layer").Size(Dp(60))))
		}
		return Div().Child(Input().ID("input"), Div().ID("anchor").Absolute().Top(60).Size(Dp(20)), Div().Absolute().Top(84).Size(Dp(60)).OnClick(func() { calls++ }))
	}))
	h := uitest.New(root)
	open = true
	h.Frame()
	if !(&Context{root: root}).Focused("input") {
		t.Fatal("nonmodal overlay stole focus")
	}
	h.Click(20, 100)
	h.Frame()
	if calls != 0 {
		t.Fatal("overlay content click reached underlying button")
	}
}

func TestOverlayMeasurementDoesNotDismissMissingAnchor(t *testing.T) {
	missing := false
	calls := 0
	root := Embed(ViewFunc(func(cx *Context) Element {
		cx.Overlay("layer", Anchored("anchor", Div().Size(Dp(30))).OnDismiss(func() { calls++ }))
		return Div().Child(Div().ID("anchor").Hidden(missing).Size(Dp(20)))
	}))
	h := uitest.NewFunc(func(gtx core.C) {
		missing = true
		for i := 0; i < 3; i++ {
			var ops op.Ops
			m := gtx
			m.Ops = &ops
			m.Source = input.Source{}
			root.Layout(m)
		}
		missing = false
		root.Layout(gtx)
	})
	h.Frame()
	if calls != 0 {
		t.Fatalf("measurement dismissed a live layer %d times", calls)
	}
}

func TestOverlayKeysPreserveEditorState(t *testing.T) {
	insert := false
	root := Root(ViewFunc(func(cx *Context) Element {
		if insert {
			cx.Overlay("insert", Anchored("anchor", Div().Name("insert").Size(Dp(20))))
		}
		cx.Overlay("editor", Anchored("anchor", Input().ID("editor").Name("editor").W(Dp(120))))
		return Div().Child(Div().ID("anchor").Size(Dp(20)))
	}))
	h := uitest.New(root)
	h.Click(center(nodeBounds(h, "editor")))
	h.Type("持久 123")
	insert = true
	h.Frame()
	if got := description(h, "editor"); got != "持久 123" {
		t.Fatalf("conditional insertion reset editor: %q", got)
	}
	insert = false
	h.Frame()
	if got := description(h, "editor"); got != "持久 123" {
		t.Fatalf("conditional removal reset editor: %q", got)
	}
}

func TestHoveredInteractiveAnchorAndModalIsolation(t *testing.T) {
	open := false
	hovered := false
	root := Root(ViewFunc(func(cx *Context) Element {
		hovered = cx.Hovered("anchor")
		if open {
			cx.Overlay("modal", Modal(Div().Size(Dp(60))))
		}
		return Div().Child(Div().ID("anchor").OnClick(func() {}).Size(Dp(50)))
	}))
	h := uitest.New(root)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(20, 20)})
	h.Frame()
	h.Frame()
	if !hovered {
		t.Fatal("interactive anchor not hovered")
	}
	open = true
	h.Frame()
	h.Frame()
	if hovered {
		t.Fatal("modal did not suppress background hover")
	}
}

func TestHoverIDDoesNotClipOverflow(t *testing.T) {
	h := uitest.New(Root(ViewFunc(func(*Context) Element {
		return Div().Child(Div().ID("parent").Size(Dp(20)).Child(Div().Absolute().Left(30).Size(Dp(30)).Name("overflow")))
	})))
	if b := nodeBounds(h, "overflow"); b != image.Rect(30, 0, 60, 30) {
		t.Fatalf("adding hover ID clipped child: %v", b)
	}
}
