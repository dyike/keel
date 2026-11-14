package kit

import (
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestSidebarResizeByDragAndKeys(t *testing.T) {
	for _, side := range []el.Side{el.Left, el.Right} {
		got := float32(0)
		v := Sidebar().Width(200).Resizable(160, 320).Side(side).OnResize(func(dp float32) { got = dp }).
			Section("Pages", SidebarItem{ID: "a", Label: "Alpha"})
		h := renderView(viewFunc(func(cx *el.Context) el.Element {
			return el.Div().Row().W(el.Dp(600)).H(el.Dp(400)).Items(el.Stretch).Child(v.Render(cx), el.Div().Grow())
		}), 600, 1)
		b := bounds(h, locale.Current().Resize)
		if b.Empty() {
			t.Fatal("missing resize handle")
		}
		x, y := center(b)
		dx := float32(40)
		if side == el.Right {
			dx = -40
		}
		h.Router.Queue(
			pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, y)},
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)},
			pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+dx, y)},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x+dx, y)},
		)
		h.Frame()
		h.Frame()
		if v.CurrentWidth() != 240 || got != 240 {
			t.Fatalf("%v: drag resized to %v (callback %v), want 240", side, v.CurrentWidth(), got)
		}
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+10*dx, y)})
		h.Frame()
		if v.CurrentWidth() != 240 {
			t.Fatal("moved without a press", v.CurrentWidth())
		}
		h.Key(key.NameEnd, 0)
		h.Frame()
		if v.CurrentWidth() != 320 {
			t.Fatal("End did not reach the maximum", v.CurrentWidth())
		}
		h.Key(key.NameHome, 0)
		h.Frame()
		if v.CurrentWidth() != 160 {
			t.Fatal("Home did not reach the minimum", v.CurrentWidth())
		}
		v.SetCollapsed(true)
		h.Frame()
		if !bounds(h, locale.Current().Resize).Empty() {
			t.Fatal("collapsed rail is resizable")
		}
	}
}

func TestSidebarSectionTitleColor(t *testing.T) {
	v := Sidebar().Section("Pages", SidebarItem{ID: "a", Label: "Alpha"})
	ink := color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xff}
	v.SectionTitleColor(ink)
	if v.titleColor == nil || *v.titleColor != ink {
		t.Fatal("title color not kept")
	}
	renderView(v, 300, 1)
}

func TestSidebarResizeAcrossFramesAndDensity(t *testing.T) {
	for _, density := range []float32{1, 1.25, 2} {
		for _, side := range []el.Side{el.Left, el.Right} {
			v := Sidebar().Width(200).Resizable(160, 320).Side(side).Section("Pages", SidebarItem{ID: "a", Label: "Alpha"})
			root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
				row := el.Div().Row().WFull().HFull().Items(el.Stretch)
				if side == el.Right {
					return row.Child(el.Div().Grow(), v.Render(cx))
				}
				return row.Child(v.Render(cx), el.Div().Grow())
			}))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: density, PxPerSp: density}
				gtx.Constraints.Max = image.Pt(int(600*density), int(400*density))
				root.Layout(gtx)
			})
			x, y := center(bounds(h, locale.Current().Resize))
			sign := float32(1)
			if side == el.Right {
				sign = -1
			}
			send := func(kind pointer.Kind, dx float32) {
				h.Router.Queue(pointer.Event{Kind: kind, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+dx*density*sign, y)})
				h.Frame()
			}
			send(pointer.Press, 0)
			for _, dx := range []float32{20, 40, 400, 60, -100, 0} {
				send(pointer.Move, dx)
				want := min(max(200+dx, float32(160)), float32(320))
				if math.Abs(float64(v.CurrentWidth()-want)) > 1 {
					t.Fatalf("density %v side %v dx %v: width=%v want %v", density, side, dx, v.CurrentWidth(), want)
				}
			}
			send(pointer.Release, 0)
			if v.resizing {
				t.Fatal("release left resize active")
			}
			v.SetDisabled(true)
			h.Frame()
			if !bounds(h, locale.Current().Resize).Empty() {
				t.Fatal("disabled sidebar has active handle")
			}
		}
	}
}
