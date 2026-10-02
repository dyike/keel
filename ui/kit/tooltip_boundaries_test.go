package kit

import (
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestTooltipSwitchingTargetsCancelsPriorDelay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	a := WithTooltip(text("A"), "Hint A")
	b := WithTooltip(text("B"), "Hint B")
	h := c.harness(func(cx *el.Context) el.Element {
		return el.Div().P(10).Gap(20).Items(el.Start).Child(a.Render(cx), b.Render(cx))
	})
	x, y := center(bounds(h, "A"))
	h.Move(x, y)
	c.advance(h, 300*time.Millisecond)
	x, y = center(bounds(h, "B"))
	h.Move(x, y)
	c.advance(h, 300*time.Millisecond)
	if shown(h, "Hint A") || shown(h, "Hint B") {
		t.Fatal("delay leaked between targets")
	}
	c.advance(h, 200*time.Millisecond)
	h.Frame()
	if !shown(h, "Hint B") || shown(h, "Hint A") {
		t.Fatal("second tooltip delay")
	}
	b.SetDisabled(true)
	h.Frame()
	c.advance(h, time.Second)
	if shown(h, "Hint B") {
		t.Fatal("disabled hint shown")
	}
}
func TestTooltipDisabledAncestorCancelsPendingDelay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	tip := WithTooltip(text("Target"), "Hint")
	disabled := false
	h := c.harness(func(cx *el.Context) el.Element {
		return el.Div().P(10).Disabled(disabled).Items(el.Start).Child(tip.Render(cx))
	})
	x, y := center(bounds(h, "Target"))
	h.Move(x, y)
	c.advance(h, 400*time.Millisecond)
	disabled = true
	h.Frame()
	c.advance(h, time.Second)
	h.Frame()
	if shown(h, "Hint") {
		t.Fatal("pending disabled tooltip fired")
	}
	disabled = false
	h.Frame()
	h.Move(x, y)
	c.advance(h, 400*time.Millisecond)
	h.Frame()
	if shown(h, "Hint") {
		t.Fatal("disabled delay was carried over")
	}
	c.advance(h, 100*time.Millisecond)
	h.Frame()
	if !shown(h, "Hint") {
		t.Fatal("fresh delay did not finish")
	}
}
