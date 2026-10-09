package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestHoverCardFocusEscapeAndDisabled(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	card := HoverCard(Button("Target", nil).ID("target"), Button("Inside", nil))
	var cx *el.Context
	h := c.harness(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().P(10).Items(el.Start).Child(card.Render(ctx), Button("Other", nil).ID("other").Render(ctx))
	})
	cx.Focus("target")
	h.Frame()
	h.Frame()
	if !card.open || !shown(h, "Inside") {
		t.Fatal("focus did not open")
	}
	h.Key(key.NameEscape, 0)
	c.advance(h, time.Second)
	h.Frame()
	if card.open {
		t.Fatal("Esc reopened under stationary focus")
	}
	cx.Focus("other")
	h.Frame()
	h.Frame()
	cx.Focus("target")
	h.Frame()
	h.Frame()
	if !card.open {
		t.Fatal("focus leave did not rearm")
	}
	card.SetDisabled(true)
	h.Frame()
	c.advance(h, time.Second)
	if card.open {
		t.Fatal("disabled card opened")
	}
}
func TestHoverCardNestedMenuSurvivesCloseDelay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	m := Menu().Item("Action", "", nil)
	m.Trigger(Button("Menu", m.Toggle))
	card := HoverCard(text("Target"), m)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(10).Items(el.Start).Child(card.Render(cx)) })
	x, y := center(bounds(h, "Target"))
	h.Move(x, y)
	c.advance(h, HoverCardOpenDelay)
	h.Frame()
	click(t, h, "Menu")
	h.Frame()
	c.advance(h, time.Second)
	h.Frame()
	if !card.open || !m.Value() || !shown(h, "Action") {
		t.Fatal("nested menu was dismissed by parent's close timer")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !card.open || m.Value() {
		t.Fatal("nested Esc")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	c.advance(h, time.Second)
	if card.open {
		t.Fatal("card Esc reopened")
	}
}

func TestHoverCardCustomDelaysAndPlacement(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	card := HoverCard(text("Target"), text("Preview")).Width(80).
		OpenDelay(100*time.Millisecond).CloseDelay(500*time.Millisecond).Placement(el.Right, el.Start).Offset(12)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(100).Items(el.Start).Child(card.Render(cx)) })
	x, y := center(bounds(h, "Target"))
	h.Move(x, y)
	c.advance(h, 90*time.Millisecond)
	if shown(h, "Preview") {
		t.Fatal("opened early")
	}
	c.advance(h, 20*time.Millisecond)
	h.Frame()
	panel, ok := semanticNode(h, "dialog")
	if !ok {
		t.Fatal("custom opening delay ignored")
	}
	anchor := bounds(h, "Target")
	if panel.Desc.Bounds.Min.X != anchor.Max.X+12 {
		t.Fatalf("right offset: anchor %v panel %v", anchor, panel.Desc.Bounds)
	}
	h.Move(5, 5)
	h.Frame()
	c.advance(h, 400*time.Millisecond)
	if !shown(h, "Preview") {
		t.Fatal("closed early")
	}
	c.advance(h, 110*time.Millisecond)
	h.Frame()
	if shown(h, "Preview") {
		t.Fatal("custom closing delay ignored")
	}
	// A changed delay restarts a pending timer without using the old deadline.
	h.Move(x, y)
	h.Frame()
	c.advance(h, 50*time.Millisecond)
	card.OpenDelay(300 * time.Millisecond)
	h.Frame()
	c.advance(h, 100*time.Millisecond)
	h.Frame()
	if shown(h, "Preview") {
		t.Fatal("old deadline fired")
	}
	c.advance(h, 210*time.Millisecond)
	h.Frame()
	if !shown(h, "Preview") {
		t.Fatal("restarted timer never fired")
	}
}

func TestHoverCardPlacementUpdatesAndImmediateDelay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	card := HoverCard(Button("Target", nil).ID("target"), text("Preview")).Width(60).OpenDelay(-1).CloseDelay(0)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(100).Items(el.Start).Child(card.Render(cx)) })
	x, y := center(bounds(h, "Target"))
	h.Move(x, y)
	h.Frame()
	h.Frame()
	if !shown(h, "Preview") {
		t.Fatal("zero delay did not open")
	}
	for _, side := range []el.Side{el.Top, el.Left, el.Bottom, el.Right} {
		card.Placement(side, el.Start).Offset(8)
		h.Frame()
		panel, ok := semanticNode(h, "dialog")
		if !ok {
			t.Fatal("lost panel")
		}
		a, b := bounds(h, "Target"), panel.Desc.Bounds
		switch side {
		case el.Top:
			if b.Max.Y != a.Min.Y-8 {
				t.Fatalf("top %v %v", a, b)
			}
		case el.Bottom:
			if b.Min.Y != a.Max.Y+8 {
				t.Fatalf("bottom %v %v", a, b)
			}
		case el.Left:
			if b.Max.X != a.Min.X-8 {
				t.Fatalf("left %v %v", a, b)
			}
		case el.Right:
			if b.Min.X != a.Max.X+8 {
				t.Fatalf("right %v %v", a, b)
			}
		}
	}
	h.Move(5, 5)
	h.Frame()
	h.Frame()
	if shown(h, "Preview") {
		t.Fatal("zero close delay did not close")
	}
}
