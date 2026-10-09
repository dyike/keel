package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
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

func TestTooltipRichContentBindingAndPlacement(t *testing.T) {
	const action = "test.tooltip.save"
	t.Cleanup(func() { core.Bind(action) })
	if err := core.Bind(action, "ctrl+s"); err != nil {
		t.Fatal(err)
	}
	clicks := 0
	tip := WithTooltip(Button("Target", nil).ID("target"), "Save hint").
		Content(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().Child(el.Text("Rich description"), Button("Not interactive", func() { clicks++ }).Render(cx))
		})).
		Action(action).Placement(el.Right, el.Start).Offset(10)
	c := &clock{now: time.Unix(100, 0)}
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(60).Items(el.Start).Child(tip.Render(cx)) })
	click(t, h, "Target")
	h.Frame()
	n, ok := semanticNode(h, "tooltip")
	if !ok || n.Desc.Label != "Save hint" {
		t.Fatal("missing accessible tooltip name")
	}
	if !shown(h, "Rich description") || !shown(h, "ctrl+s") {
		t.Fatal("missing rich content or key binding")
	}
	a := bounds(h, "Target")
	if n.Desc.Bounds.Min.X != a.Max.X+10 {
		t.Fatalf("placement %v %v", a, n.Desc.Bounds)
	}
	core.Bind(action, "ctrl+k")
	h.Frame()
	if shown(h, "ctrl+s") || !shown(h, "ctrl+k") {
		t.Fatal("rebind did not update")
	}
	click(t, h, "Not interactive")
	h.Frame()
	if clicks != 0 {
		t.Fatal("tooltip content accepted input")
	}
	core.Bind(action)
	tip.Content(nil)
	h.Frame()
	if shown(h, "ctrl+k") || shown(h, "Rich description") {
		t.Fatal("unbinding or content reset failed")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if _, ok := semanticNode(h, "tooltip"); ok {
		t.Fatal("Esc did not dismiss")
	}
}
