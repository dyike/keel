package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestStepperNavigationModesAndCurrentNoop(t *testing.T) {
	calls := 0
	v := Stepper("First", "Second", "Third").TextCenter(true).OnChange(func(int) { calls++ })
	v.SetValue(1)
	h := renderView(v, 360, 1)
	click(t, h, "First")
	if v.Value() != 1 || calls != 0 {
		t.Fatal("read-only navigation")
	}
	v.Navigation(StepperNavigationCompleted)
	h.Frame()
	click(t, h, "Third")
	if v.Value() != 1 || calls != 0 {
		t.Fatal("completed mode allowed future step")
	}
	click(t, h, "First")
	if v.Value() != 0 || calls != 1 {
		t.Fatal("completed step unavailable")
	}
	v.Navigation(StepperNavigationAll)
	h.Frame()
	click(t, h, "Third")
	if v.Value() != 2 || calls != 2 {
		t.Fatal("all mode future step unavailable")
	}
	click(t, h, "Third")
	if calls != 2 {
		t.Fatal("current step emitted callback")
	}
	v.SetItemDisabled(1, true)
	v.SetValue(0)
	h.Frame()
	click(t, h, "First")
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if v.Value() != 2 || calls != 3 {
		t.Fatal("keyboard did not skip disabled step", v.Value(), calls)
	}
	v.Navigation(StepperNavigationNone)
	h.Frame()
	h.Key(key.NameReturn, 0)
	if calls != 3 {
		t.Fatal("navigation mode left stale handler")
	}
}

func TestStepperCenteredColumnsDirectionAndOverflow(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Stepper("One", "Two", "Three").TextCenter(true).Navigation(StepperNavigationAll)
		width := float32(360)
		h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(width)).Child(v.Render(cx)) }), 400, scale)
		h.Frame()
		a, b, c := bounds(h, "One"), bounds(h, "Two"), bounds(h, "Three")
		if a.Dx() != 120*scale || b.Dx() != a.Dx() || c.Dx() != a.Dx() || a.Min.Y != b.Min.Y || a.Dy() <= 24*scale {
			t.Fatal("centered columns", a, b, c)
		}
		v.Vertical()
		h.Frame()
		a, b = bounds(h, "One"), bounds(h, "Two")
		if b.Min.Y <= a.Max.Y {
			t.Fatal("vertical direction failed", a, b)
		}
		v.Horizontal()
		width = 100
		h.Frame()
		h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(50*float32(scale), 10*float32(scale)), Scroll: f32.Pt(1000*float32(scale), 0)})
		h.Frame()
		h.Frame()
		c = bounds(h, "Three")
		if c.Empty() || c.Max.X > 100*scale || c.Min.X < 0 {
			t.Fatal("centered last step unreachable", c)
		}
		click(t, h, "Three")
		if v.Value() != 2 {
			t.Fatal("scrolled centered step cannot activate")
		}
	}
}
