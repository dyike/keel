package el

import (
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTouchHoldOpensContextMenu(t *testing.T) {
	menus, clicks := 0, 0
	root := Embed(ViewFunc(func(*Context) Element {
		return Div().Child(Div().W(Dp(200)).H(Dp(100)).OnClick(func() { clicks++ }).OnContextMenu(func() { menus++ }))
	}))
	now := time.Unix(100, 0)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Now = now; root.Layout(gtx) })
	h.Frame()
	touch := func(kind pointer.Kind, x float32) {
		h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Touch, PointerID: 1, Position: f32.Pt(x, 20)})
		h.Frame()
	}
	wait := func(d time.Duration) { now = now.Add(d); h.Frame() }

	// A tap is a click.
	touch(pointer.Press, 20)
	touch(pointer.Release, 20)
	if clicks != 1 || menus != 0 {
		t.Fatal("tap", clicks, menus)
	}
	// Holding still opens the menu once, and lifting is no click.
	touch(pointer.Press, 20)
	wait(300 * time.Millisecond)
	if menus != 0 {
		t.Fatal("menu before the hold delay")
	}
	wait(300 * time.Millisecond)
	wait(300 * time.Millisecond)
	touch(pointer.Release, 20)
	if menus != 1 || clicks != 1 {
		t.Fatal("hold", menus, clicks)
	}
	// Drifting turns the hold into a drag.
	touch(pointer.Press, 20)
	touch(pointer.Move, 60)
	wait(time.Second)
	touch(pointer.Release, 60)
	if menus != 1 {
		t.Fatal("drifting finger opened the menu")
	}
	// The mouse keeps its secondary button.
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonSecondary, Position: f32.Pt(20, 20)})
	h.Frame()
	if menus != 2 {
		t.Fatal("right click", menus)
	}
}
