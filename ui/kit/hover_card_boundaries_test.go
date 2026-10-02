package kit

import (
	"gioui.org/io/key"
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
