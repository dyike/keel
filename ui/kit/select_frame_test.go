package kit

import (
	"image"
	"testing"

	"gioui.org/gpu/headless"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// The clickable/select semantic node must cover the frame, not just its text.
// Check all four padding corners and keyboard activation after dismissal.
func TestSelectWholeFieldTrigger(t *testing.T) {
	for _, plain := range []bool{false, true} {
		for _, height := range []float32{32, 36, 48} {
			v := Select("", "完全访问", "只读访问").Hint("访问权限").Size(height).Appearance(!plain)
			v.SetValue("完全访问")
			h := renderView(viewFunc(func(cx *el.Context) el.Element {
				return el.Div().W(el.Dp(320)).Child(v.Render(cx), el.Div().Role("group").Name("after").H(el.Dp(1)))
			}), 320, 1)
			h.Frame()
			box, ok := node(h, "访问权限")
			if !ok {
				t.Fatal("select missing")
			}
			inset := 0
			if !plain {
				inset = 2
			}
			if box.Desc.Bounds.Dy() != int(height)-inset || bounds(h, "after").Min.Y != int(height) {
				t.Fatalf("plain=%v height=%v: trigger=%v after=%v", plain, height, box.Desc.Bounds, bounds(h, "after"))
			}
			for _, x := range []int{box.Desc.Bounds.Min.X + 1, box.Desc.Bounds.Max.X - 2} {
				for _, y := range []int{box.Desc.Bounds.Min.Y + 1, box.Desc.Bounds.Max.Y - 2} {
					h.Click(float32(x), float32(y))
					h.Frame()
					if !v.open {
						t.Fatalf("plain=%v padding corner (%d,%d) did not open", plain, x, y)
					}
					h.Key(key.NameEscape, 0)
					h.Frame()
				}
			}
			h.Key(key.NameSpace, 0)
			h.Frame()
			if !v.open {
				t.Fatal("Escape did not restore keyboard activation")
			}
		}
	}
}

// A pointer pick returns focus to the trigger without its focus ring, as a
// click on a button does; keyboard focus still shows the ring.
func TestSelectRingsOnlyForKeyboardFocus(t *testing.T) {
	v := Select("", "完全访问", "只读访问").Hint("访问权限")
	gpu, err := headless.NewWindow(320, 240)
	if err != nil {
		t.Skip("no headless GPU:", err)
	}
	defer gpu.Release()
	root := el.Root(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(10).Gap(10).W(el.Dp(320)).Items(el.Stretch).Child(v.Render(cx), Button("下一步", func() {}).Render(cx))
	}))
	img := image.NewRGBA(image.Rect(0, 0, 320, 240))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}
		gtx.Constraints = layout.Exact(image.Pt(320, 240))
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
	})
	h.Frame()
	trigger, _ := node(h, "访问权限")
	b := trigger.Desc.Bounds
	// The frame's 1dp border is just outside the trigger. Rendering blends
	// it, so tell the gray border from the primary ring by its blue.
	ringed := func() bool {
		// Render reads focus from the previous frame.
		h.Frame()
		h.Frame()
		c := img.RGBAAt(b.Min.X-1, (b.Min.Y+b.Max.Y)/2)
		return int(c.B)-int(c.R) > 32
	}
	if ringed() {
		t.Fatal("ring before focus")
	}
	h.Click(float32(b.Min.X+b.Max.X)/2, float32(b.Min.Y+b.Max.Y)/2)
	h.Frame()
	option := innermost(h, "只读访问")
	h.Click(float32(option.Min.X+option.Max.X)/2, float32(option.Min.Y+option.Max.Y)/2)
	h.Frame()
	if v.Value() != "只读访问" || v.open {
		t.Fatalf("pointer pick: value %q open %v", v.Value(), v.open)
	}
	if ringed() {
		t.Fatal("ring after a pointer pick")
	}
	h.Key(key.NameTab, 0) // to the button
	h.Key(key.NameTab, key.ModShift)
	if !ringed() {
		t.Fatal("no ring with keyboard focus")
	}
}
