package kit

import (
	"image"
	"slices"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/locale"
)

func navMobileHarness(v el.View) *uitest.Harness {
	root := el.Root(v)
	return uitest.NewFunc(func(g core.C) { g.Constraints.Max = image.Pt(320, 600); root.Layout(g) })
}

func TestNavStackModalDismissAndRetainedState(t *testing.T) {
	nav := NavStack()
	draft := Input("Draft")
	nav.Page(NavPage{ID: "home", Title: "Home", Create: func() el.View { return draft }})
	nav.Page(NavPage{ID: "modal", Title: "Modal", Modal: true, Create: func() el.View { return Input("Modal notes") }})
	nav.Page(NavPage{ID: "detail", Title: "Details"})
	changes := 0
	nav.OnChange(func([]string) { changes++ })
	h := navMobileHarness(nav)
	clickClass(t, h, "Editor", "Draft")
	h.Type("retained")
	nav.Push("modal")
	h.Frame()
	h.Frame()
	if !shown(h, "Modal notes") {
		t.Fatal("modal not shown")
	}
	clickClass(t, h, "Editor", "Modal notes")
	h.Type("private")
	if draft.Value() != "retained" {
		t.Fatal("modal input reached background")
	}
	nav.Push("detail")
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if !slices.Equal(nav.Value(), []string{"home"}) || changes != 3 {
		t.Fatal("dismiss did not close presentation once", nav.Value(), changes)
	}
	if draft.Value() != "retained" {
		t.Fatal("parent state lost")
	}
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	if draft.Value() != "retaine" {
		t.Fatal("focus not restored", draft.Value())
	}
	nav.Push("modal")
	nav.Push("modal")
	h.Frame()
	h.Frame()
	click(t, h, locale.Current().Close)
	h.Frame()
	if len(nav.Value()) != 2 {
		t.Fatal("nested close removed outer modal", nav.Value())
	}
	nav.SetValue([]string{"home", "modal", "detail"})
	h.Frame()
	if !nav.Dismiss() || !slices.Equal(nav.Value(), []string{"home"}) {
		t.Fatal("restored modal did not dismiss")
	}
	if nav.Dismiss() {
		t.Fatal("root dismissed")
	}
}

func TestNavStackEdgeSwipeAndCancellation(t *testing.T) {
	for _, tc := range []struct {
		name         string
		x, y, dx, dy float32
		cancel, off  bool
		want         int
	}{
		{name: "commit", x: 10, y: 150, dx: 150, want: 1},
		{name: "short", x: 10, y: 150, dx: 30, want: 2},
		{name: "middle", x: 100, y: 150, dx: 150, want: 2},
		{name: "vertical", x: 10, y: 150, dx: 10, dy: 150, want: 2},
		{name: "left", x: 10, y: 150, dx: -50, want: 2},
		{name: "cancel", x: 10, y: 150, dx: 150, cancel: true, want: 2},
		{name: "disabled", x: 10, y: 150, dx: 150, off: true, want: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nav := NavStack().Page(NavPage{ID: "home"}).Page(NavPage{ID: "detail", Create: func() el.View {
				return el.ViewFunc(func(*el.Context) el.Element { return el.Div().H(el.Dp(1200)).Child(el.Text("Long page")) })
			}})
			nav.Push("detail")
			nav.SwipeBack(!tc.off)
			changes := 0
			nav.OnChange(func([]string) { changes++ })
			h := navMobileHarness(nav)
			send := func(k pointer.Kind, x, y float32) {
				h.Router.Queue(pointer.Event{Kind: k, Source: pointer.Touch, PointerID: 1, Position: f32.Pt(x, y)})
				h.Frame()
			}
			send(pointer.Press, tc.x, tc.y)
			send(pointer.Move, tc.x+tc.dx, tc.y+tc.dy)
			end := pointer.Release
			if tc.cancel {
				end = pointer.Cancel
			}
			send(end, tc.x+tc.dx, tc.y+tc.dy)
			h.Frame()
			if len(nav.Value()) != tc.want {
				t.Fatal("unexpected path", nav.Value())
			}
			if changes != 2-tc.want {
				t.Fatal("unexpected callback count", changes)
			}
			if len(nav.entries) > 1 && (nav.entries[1].swiping || nav.entries[1].swipeOffset != 0) {
				t.Fatal("gesture did not reset")
			}
		})
	}
}

func TestNavStackSwipeLeavesVerticalScrollingAndChildInput(t *testing.T) {
	offset := float32(0)
	nav := NavStack().Page(NavPage{ID: "home"})
	field := Input("Notes")
	nav.Page(NavPage{ID: "detail", Create: func() el.View {
		return el.ViewFunc(func(cx *el.Context) el.Element {
			offset, _, _ = cx.ScrollState("nav-stack/content")
			return el.Div().Gap(20).Child(field.Render(cx), el.Div().H(el.Dp(1200)).Child(el.Text("Scroll content")))
		})
	}})
	nav.Push("detail")
	h := navMobileHarness(nav)
	clickClass(t, h, "Editor", "Notes")
	h.Type("kept")
	if field.Value() != "kept" {
		t.Fatal("gesture stole input")
	}
	send := func(k pointer.Kind, x, y float32) {
		h.Router.Queue(pointer.Event{Kind: k, Source: pointer.Touch, PointerID: 1, Position: f32.Pt(x, y)})
		h.Frame()
	}
	send(pointer.Press, 10, 350)
	send(pointer.Move, 10, 250)
	send(pointer.Move, 10, 180)
	send(pointer.Release, 10, 180)
	h.Frame()
	if offset <= 0 || len(nav.Value()) != 2 {
		t.Fatal("edge gesture blocked vertical scrolling", offset, nav.Value())
	}
	send(pointer.Press, 10, 350)
	send(pointer.Move, 180, 350)
	nav.SetDisabled(true)
	h.Frame()
	send(pointer.Release, 180, 350)
	nav.SetDisabled(false)
	h.Frame()
	if len(nav.Value()) != 2 || nav.entries[1].swiping || nav.entries[1].swipeOffset != 0 {
		t.Fatal("disabled drag committed or stayed active")
	}
	nav.SwipeBack(false)
	h.Frame()
	send(pointer.Press, 10, 350)
	send(pointer.Move, 180, 350)
	send(pointer.Release, 180, 350)
	if len(nav.Value()) != 2 {
		t.Fatal("opted-out swipe navigated")
	}
}

func TestNavStackNestedModalReturnsInputFocus(t *testing.T) {
	nav := NavStack().Page(NavPage{ID: "home"})
	var inputs []*InputView
	nav.Page(NavPage{ID: "edit", Title: "Edit", Modal: true, Create: func() el.View { v := Input("Notes"); inputs = append(inputs, v); return v }})
	h := navMobileHarness(nav)
	nav.Push("edit")
	h.Frame()
	h.Frame()
	clickClass(t, h, "Editor", "Notes")
	h.Type("outer")
	nav.Push("edit")
	h.Frame()
	h.Frame()
	clickClass(t, h, "Editor", "Notes")
	h.Type("inner")
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Key(key.NameEnd, 0)
	h.Key(key.NameDeleteBackward, 0)
	if len(nav.Value()) != 2 || inputs[0].Value() != "oute" || inputs[1].Value() != "inner" {
		t.Fatal("nested focus/state not restored", nav.Value(), inputs[0].Value(), inputs[1].Value())
	}
}
