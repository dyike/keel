package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/theme"
	"testing"
	"time"
)

func TestCollapsibleSeparatePartsAndInterruptedAnimation(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	calls := 0
	v := Collapsible("Details", el.ViewFunc(func(*el.Context) el.Element { return el.Div().H(el.Dp(100)).Child(el.Text("Body")) })).Heading(text("Custom heading")).OnChange(func(bool) { calls++ })
	var cx *el.Context
	h := c.harness(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().W(el.Dp(300)).Child(v.Trigger().Render(ctx), el.Text("Between"), v.Content().Render(ctx), el.Text("After"))
	})
	before := bounds(h, "After").Min.Y
	click(t, h, "Details")
	c.advance(h, DisclosureDuration/2)
	middle := bounds(h, "After").Min.Y
	if middle <= before || middle >= before+114 || calls != 1 || !v.Value() {
		t.Fatalf("opening geometry %d %d", before, middle)
	}
	v.SetValue(false)
	h.Frame()
	if y := bounds(h, "After").Min.Y; y != middle {
		t.Fatalf("reversal jumped %d -> %d", middle, y)
	}
	c.advance(h, DisclosureDuration)
	if bounds(h, "After").Min.Y != before || calls != 1 {
		t.Fatal("programmatic close or completion")
	}
	cx.Focus(autoID("collapsible", v) + "/trigger")
	h.Frame()
	h.Key(key.NameReturn, 0)
	c.advance(h, DisclosureDuration)
	if !v.Value() || calls != 2 || !shown(h, "Body") {
		t.Fatal("keyboard reopen")
	}
	v.SetDisabled(true)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if !v.Value() || calls != 2 {
		t.Fatal("disabled toggled")
	}
}
func TestCollapsibleReducedMotionFocusAndInputPreservation(t *testing.T) {
	old := theme.ReducedMotion
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	input := Input("saved")
	input.SetValue("saved")
	v := Collapsible("Details", input)
	v.SetValue(true)
	var cx *el.Context
	h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return v.Render(ctx) }), 300, 1)
	cx.Focus(input.FocusID())
	h.Frame()
	v.SetValue(false)
	h.Frame()
	h.Frame()
	if !cx.Focused(autoID("collapsible", v)+"/trigger") || shown(h, "saved") {
		t.Fatal("closed focus did not return")
	}
	v.SetValue(true)
	h.Frame()
	if input.Value() != "saved" || v.motion.value != 1 {
		t.Fatal("state or reduced motion")
	}
}
