package kit

import (
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestStepperNarrowScrollingNavigationAndOwnership(t *testing.T) {
	for _, scale := range []int{1, 2} {
		labels := []string{"填写订单", "确认付款", "安排发货", "完成 Done 123"}
		calls := 0
		v := Stepper(labels...).Navigable().OnChange(func(int) { calls++ })
		v.SetValue(4)
		labels[3] = "mutated"
		h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(224)).Child(v.Render(cx)) }), 224, scale)
		h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(100*float32(scale), 10*float32(scale)), Scroll: f32.Pt(1000*float32(scale), 0)})
		h.Frame()
		h.Frame()
		b := bounds(h, "完成 Done 123")
		if b.Empty() || b.Min.X < 0 || b.Max.X > 224*scale {
			t.Fatal("last step unreachable", b)
		}
		v.SetDisabled(true)
		h.Frame()
		click(t, h, "完成 Done 123")
		h.Frame()
		if calls != 0 || v.Value() != 4 {
			t.Fatal("disabled step changed")
		}
		v.SetDisabled(false)
		h.Frame()
		click(t, h, "完成 Done 123")
		h.Frame()
		if calls != 1 || v.Value() != 3 {
			t.Fatal("last step navigation", calls, v.Value())
		}
	}
}

func TestStepperVerticalEntriesSizeAndKeyboard(t *testing.T) {
	for _, scale := range []int{1, 2} {
		calls := 0
		items := []StepperItem{{Label: "First", Icon: IconUser, Disabled: true}, {Label: "Second", Icon: IconInbox}, {Label: "Third"}}
		v := Stepper().Vertical().Size(32).Navigable().OnChange(func(int) { calls++ })
		v.SetEntries(items...)
		v.SetValue(3)
		items[1].Label = "mutated"
		copy := v.Entries()
		copy[1].Label = "also mutated"
		h := renderView(v, 200, scale)
		first, second := bounds(h, "First"), bounds(h, "Second")
		if first.Min.X != second.Min.X || second.Min.Y <= first.Max.Y || first.Dy() != 32*scale {
			t.Fatalf("vertical size/alignment: %v %v", first, second)
		}
		click(t, h, "First")
		if v.Value() != 3 || calls != 0 {
			t.Fatal("disabled step activated")
		}
		h.Router.MoveFocus(key.FocusForward)
		h.Frame()
		h.Key(key.NameReturn, 0)
		if v.Value() != 1 || calls != 1 {
			t.Fatalf("Tab did not skip disabled step: value=%d calls=%d", v.Value(), calls)
		}
		v.SetItemDisabled(0, false)
		v.SetValue(3)
		h.Frame()
		click(t, h, "First")
		if v.Value() != 0 || calls != 2 {
			t.Fatal("reenabled step did not activate")
		}
		v.SetValue(3)
		v.SetDisabled(true)
		h.Frame()
		click(t, h, "Second")
		h.Key(key.NameReturn, 0)
		if v.Value() != 3 || calls != 2 {
			t.Fatal("disabled stepper activated")
		}
		v.SetItemDisabled(-1, true)
		v.SetItemDisabled(99, true)
		v.SetEntries(StepperItem{Label: "Only"})
		if v.Value() != 1 || calls != 2 {
			t.Fatal("entry update clamp/callback")
		}
		v.SetEntries()
		h.Frame()
		if v.Value() != 0 {
			t.Fatal("empty entry update did not clamp")
		}
	}
}

func TestStepperVerticalScrollAndRichContent(t *testing.T) {
	v := Stepper().Vertical().Navigable()
	v.SetEntries(StepperItem{Label: "First", Icon: IconUser, Content: viewFunc(func(*el.Context) el.Element {
		return el.Div().Child(el.Text("Account"), el.Text("Add details"))
	})}, StepperItem{Label: "Second"}, StepperItem{Label: "Last"})
	v.SetValue(3)
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().H(el.Dp(100)).Child(v.Render(cx)) }), 200, 1)
	if !shown(h, "Account") || !shown(h, "Add details") {
		t.Fatal("rich content missing")
	}
	h.Scroll(60, 40, 1000)
	h.Frame()
	b := bounds(h, "Last")
	if b.Empty() || b.Min.Y < 0 || b.Max.Y > 100 {
		t.Fatalf("last step unreachable: %v", b)
	}
	click(t, h, "Last")
	if v.Value() != 2 {
		t.Fatal("scrolled step not clickable")
	}
}

func TestStepperInvalidSizePreservesLayout(t *testing.T) {
	v := Stepper("First").Size(20)
	for _, size := range []float32{0, -1, float32(math.NaN()), float32(math.Inf(1))} {
		v.Size(size)
		h := renderView(v, 200, 1)
		if b := bounds(h, "First"); b.Dy() != 20 {
			t.Fatalf("invalid size %v changed layout: %v", size, b)
		}
	}
}

func TestStepperAncestorDisabledNavigation(t *testing.T) {
	v := Stepper("First", "Second").Vertical().Navigable()
	v.SetValue(2)
	disabled := true
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 200, 1)
	click(t, h, "First")
	if v.Value() != 2 {
		t.Fatal("ancestor disabled navigation")
	}
	disabled = false
	h.Frame()
	click(t, h, "First")
	if v.Value() != 0 {
		t.Fatal("ancestor reenable did not restore navigation")
	}
}
