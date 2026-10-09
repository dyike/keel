package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"slices"
	"testing"
	"time"
)

func TestAccordionSeparatePartsKeyboardAndDisabledAncestors(t *testing.T) {
	a := Accordion().Multiple().Add("First", text("First body")).Add("Unavailable", text("Second body")).Add("Third", text("Third body")).Heading(0, text("Custom heading"))
	var cx *el.Context
	h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().W(el.Dp(300)).Child(a.Trigger(0).Render(ctx), el.Div().Disabled(true).Child(a.Trigger(1).Render(ctx)), a.Trigger(2).Render(ctx), a.Content(0).Render(ctx), a.Content(1).Render(ctx), a.Content(2).Render(ctx))
	}), 300, 1)
	click(t, h, "First")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if !cx.Focused(a.triggerID(2)) {
		t.Fatal("keyboard entered disabled ancestor")
	}
	h.Key(key.NameReturn, 0)
	if !slices.Equal(a.Value(), []int{0, 2}) {
		t.Fatal("independent multiple state")
	}
	a.SetDisabled(true)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if !slices.Equal(a.Value(), []int{0, 2}) {
		t.Fatal("disabled toggled")
	}
	snapshot := a.Value()
	snapshot[0] = 99
	if a.Value()[0] != 0 {
		t.Fatal("value aliases")
	}
}
func TestAccordionAnimationReversalAndInputState(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	input := Input("Name")
	input.SetValue("kept")
	a := Accordion().Add("First", input).Add("Second", text("Second body"))
	calls := 0
	a.OnChange(func([]int) { calls++ })
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(300)).Child(a.Render(cx), el.Text("After")) })
	start := bounds(h, "After").Min.Y
	click(t, h, "First")
	c.advance(h, DisclosureDuration/2)
	mid := bounds(h, "After").Min.Y
	if mid <= start {
		t.Fatal("no intermediate animation height")
	}
	a.SetValue()
	h.Frame()
	if bounds(h, "After").Min.Y != mid {
		t.Fatal("reversal jumped")
	}
	c.advance(h, DisclosureDuration)
	if bounds(h, "After").Min.Y != start {
		t.Fatal("close did not finish")
	}
	a.SetValue(0)
	h.Frame()
	c.advance(h, DisclosureDuration)
	if input.Value() != "kept" || calls != 1 {
		t.Fatal("input state or setter callback")
	}
	a.SetItemDisabled(0, true)
	h.Frame()
	click(t, h, "First")
	if !slices.Equal(a.Value(), []int{0}) {
		t.Fatal("disabled item toggled")
	}
}
