package kit

import (
	"image"
	"testing"
	"time"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/unit"
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

// ringHarness renders view on a headless GPU. ringed reports whether the
// field frame whose left border passes through (x, y) shows the primary ring.
func ringHarness(t *testing.T, view func(cx *el.Context) el.Element) (*uitest.Harness, func(x, y int) bool) {
	t.Helper()
	gpu, err := headless.NewWindow(320, 240)
	if err != nil {
		t.Skip("no headless GPU:", err)
	}
	t.Cleanup(gpu.Release)
	root := el.Root(viewFunc(view))
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
	return h, func(x, y int) bool {
		// Render reads focus from the previous frame. Rendering blends the
		// 1dp border, so tell the gray border from the ring by its blue.
		h.Frame()
		h.Frame()
		c := img.RGBAAt(x, y)
		return int(c.B)-int(c.R) > 32
	}
}

// A pointer pick returns focus to the trigger without its focus ring, as a
// click on a button does; keyboard focus still shows the ring.
func TestSelectRingsOnlyForKeyboardFocus(t *testing.T) {
	v := Select("", "完全访问", "只读访问").Hint("访问权限")
	h, ringed := ringHarness(t, func(cx *el.Context) el.Element {
		return el.Div().P(10).Gap(10).W(el.Dp(320)).Items(el.Stretch).Child(v.Render(cx), Button("下一步", func() {}).Render(cx))
	})
	trigger, _ := node(h, "访问权限")
	b := trigger.Desc.Bounds
	x, y := b.Min.X-1, (b.Min.Y+b.Max.Y)/2 // the frame's border, just outside the trigger
	if ringed(x, y) {
		t.Fatal("ring before focus")
	}
	h.Click(center(b))
	h.Frame()
	h.Click(center(innermost(h, "只读访问")))
	if v.Value() != "只读访问" || v.open {
		t.Fatalf("pointer pick: value %q open %v", v.Value(), v.open)
	}
	if ringed(x, y) {
		t.Fatal("ring after a pointer pick")
	}
	h.Key(key.NameTab, 0) // to the button
	h.Key(key.NameTab, key.ModShift)
	if !ringed(x, y) {
		t.Fatal("no ring with keyboard focus")
	}
}

// Text fields clicked into show the caret alone, through typing and a
// Combobox pick; reaching them with Tab rings their frame.
func TestTextFieldsRingOnlyForKeyboardFocus(t *testing.T) {
	in := Input("").Placeholder("名称")
	combo := Combobox("", "苹果", "香蕉").Placeholder("水果")
	h, ringed := ringHarness(t, func(cx *el.Context) el.Element {
		return el.Div().P(10).Gap(10).W(el.Dp(320)).Items(el.Stretch).Child(in.Render(cx), combo.Render(cx))
	})
	// Left borders, 10dp in, halfway down each 36dp field.
	inputRing := func() bool { return ringed(10, 28) }
	comboRing := func() bool { return ringed(10, 74) }
	field := innermost(h, "名称")
	h.Click(float32(field.Min.X+4), float32(field.Min.Y+field.Max.Y)/2)
	h.Type("值")
	if in.Value() != "值" || inputRing() {
		t.Fatalf("clicked input: value %q ring %v", in.Value(), inputRing())
	}
	field = innermost(h, "水果")
	h.Click(float32(field.Min.X+4), float32(field.Min.Y+field.Max.Y)/2)
	h.Type("香")
	if !combo.open || comboRing() {
		t.Fatalf("clicked combobox: open %v ring %v", combo.open, comboRing())
	}
	h.Click(center(innermost(h, "香蕉")))
	if combo.Value() != "香蕉" || combo.open || comboRing() {
		t.Fatalf("pointer pick: value %q open %v ring %v", combo.Value(), combo.open, comboRing())
	}
	// Without TabIndex the window moves focus on Tab itself, through Gio.
	h.Router.MoveFocus(key.FocusBackward)
	if !inputRing() || comboRing() {
		t.Fatalf("Shift+Tab to the input: input ring %v combobox ring %v", inputRing(), comboRing())
	}
	h.Router.MoveFocus(key.FocusForward)
	if inputRing() || !comboRing() {
		t.Fatalf("Tab to the combobox: input ring %v combobox ring %v", inputRing(), comboRing())
	}
}

// A field clicked into shows only its caret, so the caret must be there: an
// empty field draws it just before the placeholder, in the frame's gap.
func TestEmptyFieldShowsCaretBeforePlaceholder(t *testing.T) {
	in := Input("搜索").Placeholder("客户名称").Prefix(el.ViewFunc(func(cx *el.Context) el.Element {
		return Icon(IconSearch).Size(16).Render(cx)
	}))
	gpu, err := headless.NewWindow(320, 120)
	if err != nil {
		t.Skip("no headless GPU:", err)
	}
	defer gpu.Release()
	root := el.Root(viewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(10).W(el.Dp(320)).Items(el.Stretch).Child(in.Render(cx))
	}))
	img := image.NewRGBA(image.Rect(0, 0, 320, 120))
	now := time.Unix(1000, 0) // a fixed clock keeps the caret in its visible phase
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = now
		gtx.Metric = unit.Metric{PxPerDp: 1, PxPerSp: 1}
		gtx.Constraints = layout.Exact(image.Pt(320, 120))
		root.Layout(gtx)
		if err := gpu.Frame(gtx.Ops); err != nil {
			t.Fatal(err)
		}
		if err := gpu.Screenshot(img); err != nil {
			t.Fatal(err)
		}
	})
	h.Frame()
	hint := innermost(h, "客户名称")
	caret := func() int {
		dark := 0
		for y := hint.Min.Y; y < hint.Max.Y; y++ {
			for x := hint.Min.X - 4; x < hint.Min.X; x++ {
				if c := img.RGBAAt(x, y); (int(c.R)+int(c.G)+int(c.B))/3 < 0x90 {
					dark++
				}
			}
		}
		return dark
	}
	if n := caret(); n != 0 {
		t.Fatalf("%d caret pixels before focus", n)
	}
	h.Click(float32(hint.Min.X+30), float32(hint.Min.Y+hint.Max.Y)/2)
	h.Frame()
	h.Frame()
	if n := caret(); n < hint.Dy()/2 {
		t.Fatalf("%d caret pixels before the placeholder, want a caret", n)
	}
}
