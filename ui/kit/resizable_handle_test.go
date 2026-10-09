package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestResizableHandleStates(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	for _, vertical := range []bool{false, true} {
		v := Resizable(nil, nil).Min(40, 40)
		v.SetValue(160)
		if vertical {
			v.Vertical()
		}
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().W(el.Dp(400)).H(el.Dp(400)).Items(el.Stretch).Child(v.Render(cx))
		}), 400, 1)
		h.Frame()
		if v.handleMotion.value != 1 {
			t.Fatal("idle", v.handleMotion)
		}
		b := bounds(h, locale.Current().Resize)
		x, y := center(b)
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)})
		h.Frame()
		h.Frame()
		if v.handleMotion.value != 2 {
			t.Fatal("hover", v.handleMotion)
		}
		h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
		h.Frame()
		if !v.pressed || v.handleMotion.value != 3 {
			t.Fatal("press", v.pressed, v.handleMotion)
		}
		if vertical {
			y += 25
		} else {
			x += 25
		}
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
		h.Frame()
		h.Frame()
		if !v.dragging || v.handleMotion.value != 4 {
			t.Fatal("drag", v.dragging, v.handleMotion)
		}
		before := v.Value()
		h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
		h.Frame()
		h.Frame()
		if v.pressed || v.dragging || v.Value() != before {
			t.Fatal("cancel moved split or stuck", v.Value(), before)
		}
		v.HandleAppearance(func(a ResizableHandleAppearance) ResizableHandleAppearance {
			a.Idle = 2
			a.Hover = 5
			a.Duration = 0
			return a
		})
		v.SetDisabled(true)
		h.Frame()
		h.Frame()
		if v.handleMotion.value != 2 {
			t.Fatal("disabled did not use idle", v.handleMotion)
		}
	}
}
