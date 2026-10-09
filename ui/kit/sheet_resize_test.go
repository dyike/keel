package kit

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"github.com/dyike/keel/ui/theme"
	"testing"
)

func TestSheetResizeDirectionsAndLimits(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	for _, scale := range []int{1, 2} {
		for _, side := range []el.Side{el.Left, el.Right, el.Top, el.Bottom} {
			calls := 0
			s := Sheet(side, "Panel").Size(200).MarginTop(32).OnResize(func(float32) { calls++ })
			s.SetValue(true)
			h := renderView(s, 400, scale)
			x, y := center(bounds(h, locale.Current().Resize))
			dx, dy := float32(50*scale), float32(0)
			if side == el.Right {
				dx = -dx
			}
			if side == el.Top {
				dx, dy = 0, dx
			}
			if side == el.Bottom {
				dx, dy = 0, -dx
			}
			h.Drag(x, y, x+dx, y+dy)
			h.Frame()
			if s.PanelSize() != 250 {
				t.Fatal("drag", side, scale, s.PanelSize())
			}
			if calls == 0 {
				t.Fatal("missing callback")
			}
			h.Key(key.NameHome, 0)
			if s.PanelSize() != 80 {
				t.Fatal("minimum", s.PanelSize())
			}
			h.Key(key.NameEnd, 0)
			expected := float32(400)
			if side == el.Top || side == el.Bottom {
				expected = 968
			}
			if s.PanelSize() != expected {
				t.Fatal("maximum", side, s.PanelSize(), expected)
			}
			before := calls
			s.Size(200)
			h.Frame()
			if calls != before {
				t.Fatal("program size notified")
			}
			s.Resizable(false)
			h.Frame()
			if shown(h, locale.Current().Resize) {
				t.Fatal("handle not removed")
			}
			s.Resizable(true)
			h.Frame()
			if !shown(h, locale.Current().Resize) {
				t.Fatal("handle not restored")
			}
		}
	}
}

func TestSheetResizeCancellationAndAncestorDisable(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	disabled := false
	s := Sheet(el.Left, "Panel").Size(200)
	s.SetValue(true)
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(s.Render(cx)) }), 400, 1)
	x, y := center(bounds(h, locale.Current().Resize))
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+40, y)})
	h.Frame()
	if s.PanelSize() != 240 {
		t.Fatal("move", s.PanelSize())
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Cancel, Source: pointer.Mouse})
	h.Frame()
	if s.PanelSize() != 200 || s.resizing {
		t.Fatal("cancel did not restore", s.PanelSize())
	}
	disabled = true
	h.Frame()
	if s.Value() || s.resizing {
		t.Fatal("disabled owner retained panel")
	}
}

func TestSheetResizeStartsFromFittedSizeAndCanStopMidDrag(t *testing.T) {
	old := el.ReducedMotion()
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(old)
	s := Sheet(el.Right, "Panel").Size(1000)
	s.SetValue(true)
	h := renderView(s, 400, 1)
	x, y := center(bounds(h, locale.Current().Resize))
	h.Drag(x, y, x+50, y)
	h.Frame()
	if s.PanelSize() != 350 {
		t.Fatal("drag used requested rather than fitted size", s.PanelSize())
	}
	x, y = center(bounds(h, locale.Current().Resize))
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
	s.Resizable(false)
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+40, y)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x+40, y)})
	h.Frame()
	if s.PanelSize() != 350 || s.resizing {
		t.Fatal("disabled drag changed size")
	}
	s.Resizable(true)
	h.Frame()
	x, y = center(bounds(h, locale.Current().Resize))
	h.Drag(x, y, x+30, y)
	h.Frame()
	if s.PanelSize() != 320 {
		t.Fatal("new drag retained old grab", s.PanelSize())
	}
}
