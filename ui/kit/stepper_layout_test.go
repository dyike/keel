package kit

import (
	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
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
