package kit

import (
	"image"
	"math"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestPopoverNestedMenuLayerOrderAndEsc(t *testing.T) {
	m := Menu().Item("Nested action", "", nil)
	m.Trigger(Button("Menu", m.Toggle))
	p := Popover(m)
	p.Trigger(Button("Popover", p.Toggle))
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(10).Items(el.Start).Child(p.Render(cx)) })))
	click(t, h, "Popover")
	click(t, h, "Menu")
	h.Frame()
	if !m.Value() || !p.Value() || !shown(h, "Nested action") {
		t.Fatal("child overlay is underneath or immediately dismissed")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if m.Value() || !p.Value() {
		t.Fatal("Esc should close inner layer only")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.Value() {
		t.Fatal("Esc should close outer layer")
	}
}
func TestPopoverBoundsScrollingAndDisabled(t *testing.T) {
	p := Popover(el.ViewFunc(func(*el.Context) el.Element {
		return el.Div().W(el.Dp(600)).H(el.Dp(800)).Child(el.Text("Long content"))
	})).Width(1000)
	p.Width(float32(math.NaN()))
	p.Width(float32(math.Inf(1)))
	if p.width != 1000 {
		t.Fatal("invalid width accepted")
	}
	p.Trigger(Button("Open", p.Toggle))
	disabled := false
	var cx *el.Context
	root := el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Disabled(disabled).Items(el.Start).Child(p.Render(ctx))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(220, 170); root.Layout(gtx) })
	click(t, h, "Open")
	h.Frame()
	n, ok := semanticNode(h, "dialog")
	if !ok || n.Desc.Bounds.Min.X < 0 || n.Desc.Bounds.Max.X > 220 || n.Desc.Bounds.Max.Y > 170 {
		t.Fatalf("unbounded panel %v", n.Desc.Bounds)
	}
	// The panel remains scrollable instead of making the whole window taller.
	x, y := center(n.Desc.Bounds)
	h.Scroll(x, y, 100)
	if offset, _, _ := cx.ScrollState(autoID("popover", p) + "/panel"); !p.Value() || offset <= 0 {
		t.Fatal("panel did not scroll")
	}
	disabled = true
	h.Frame()
	h.Frame()
	if p.Value() {
		t.Fatal("disabled ancestor retained overlay")
	}
	disabled = false
	p.SetDisabled(true)
	p.SetValue(true)
	p.Toggle()
	h.Frame()
	if p.Value() {
		t.Fatal("disabled popover opened")
	}
}

func TestPopoverOffsetUpdates(t *testing.T) {
	p := Popover(text("Preview")).Width(60)
	p.Trigger(Button("Target", p.Toggle))
	changes := 0
	p.OnChange(func(bool) { changes++ })
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(100).Items(el.Start).Child(p.Render(cx))
	})))
	click(t, h, "Target")
	for _, gap := range []float32{4, 12, 0, -4} {
		if gap != 4 {
			p.Offset(gap)
		}
		for _, side := range []el.Side{el.Top, el.Bottom, el.Left, el.Right} {
			p.Placement(side, el.Start)
			h.Frame()
			n, ok := semanticNode(h, "dialog")
			if !ok {
				t.Fatal("lost panel")
			}
			a, b := bounds(h, "Target"), n.Desc.Bounds
			var distance int
			switch side {
			case el.Top:
				distance = a.Min.Y - b.Max.Y
			case el.Bottom:
				distance = b.Min.Y - a.Max.Y
			case el.Left:
				distance = a.Min.X - b.Max.X
			case el.Right:
				distance = b.Min.X - a.Max.X
			}
			if distance != int(gap) {
				t.Fatalf("side %v gap %v: anchor %v panel %v", side, gap, a, b)
			}
		}
	}
	for _, invalid := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1))} {
		p.Offset(invalid)
		if p.offset != -4 {
			t.Fatal("invalid offset accepted")
		}
	}
	if changes != 1 || !p.Value() {
		t.Fatal("position updates changed open state", changes)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.Value() || changes != 2 {
		t.Fatal("Esc did not dismiss repositioned panel")
	}
}

func TestPopoverAppearanceAndStylePreserveContent(t *testing.T) {
	for _, scale := range []int{1, 2} {
		input := Input("Search")
		p := Popover(input).Width(180)
		p.Trigger(Button("Open", p.Toggle))
		h := renderView(p, 240, scale)
		click(t, h, "Open")
		h.Frame()
		n, _ := semanticNode(h, "dialog")
		original := n.Desc.Bounds
		clickClass(t, h, "Editor", "Search")
		h.Type("a")
		p.Appearance(false)
		h.Frame()
		n, ok := semanticNode(h, "dialog")
		if !ok || n.Desc.Bounds.Dy() >= original.Dy() {
			t.Fatal("default padding retained")
		}
		p.PanelStyle(func(e *el.DivEl) { e.P(24).ID("ignored").Role("ignored").MaxW(el.Dp(900)) })
		h.Frame()
		n, ok = semanticNode(h, "dialog")
		if !ok || n.Desc.Bounds.Dy() <= original.Dy() || n.Desc.Bounds.Max.X > 240*scale {
			t.Fatal("panel refinement or bounds lost", n)
		}
		h.Key(key.NameRightArrow, 0)
		h.Key(key.NameDeleteBackward, 0)
		if input.Value() != "" {
			t.Fatal("style update lost input focus", input.Value())
		}
		p.Appearance(true).PanelStyle(nil)
		h.Frame()
		n, ok = semanticNode(h, "dialog")
		if !ok || n.Desc.Bounds.Size() != original.Size() {
			t.Fatal("default style not restored", n.Desc.Bounds, original)
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		if p.Value() {
			t.Fatal("styled panel did not close")
		}
	}
}

func TestPopoverRightClickTrigger(t *testing.T) {
	primary, changes := 0, 0
	p := Popover(text("Preview")).RightClick(true).OnChange(func(bool) { changes++ })
	p.Trigger(Button("Target", func() { primary++ }))
	disabled := false
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(10).Disabled(disabled).Items(el.Start).Child(p.Render(cx))
	})))
	click(t, h, "Target")
	if primary != 1 || p.Value() {
		t.Fatal("primary behavior changed")
	}
	tableRightClick(t, h, "Target")
	if !p.Value() || !shown(h, "Preview") || changes != 1 || primary != 1 {
		t.Fatal("right-click did not exclusively open", changes, primary)
	}
	tableRightClick(t, h, "Target")
	if p.Value() || changes != 2 {
		t.Fatal("repeat right-click did not close", changes)
	}
	p.RightClick(false)
	h.Frame()
	tableRightClick(t, h, "Target")
	if p.Value() || changes != 2 {
		t.Fatal("disabled right-click handler retained")
	}
	p.RightClick(true)
	h.Frame()
	tableRightClick(t, h, "Target")
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.Value() || changes != 4 {
		t.Fatal("Esc did not close", changes)
	}
	disabled = true
	h.Frame()
	tableRightClick(t, h, "Target")
	if p.Value() || changes != 4 {
		t.Fatal("ancestor disabled bypassed")
	}
	disabled = false
	p.SetDisabled(true)
	h.Frame()
	tableRightClick(t, h, "Target")
	if p.Value() {
		t.Fatal("disabled popover opened")
	}
}

func TestPopoverMouseButtons(t *testing.T) {
	p := Popover(text("Preview"))
	primary, changes := 0, 0
	p.Trigger(Button("Target", func() { primary++ })).OnChange(func(bool) { changes++ })
	h := renderView(p, 300, 1)
	press := func(button pointer.Buttons) {
		x, y := center(bounds(h, "Target"))
		pos := f32.Pt(x, y)
		h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: pos}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: pos, Buttons: button}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: pos})
		h.Frame()
		h.Frame()
	}
	buttons := []pointer.Buttons{pointer.ButtonPrimary, pointer.ButtonSecondary, pointer.ButtonTertiary}
	for _, selected := range buttons {
		p.MouseButton(selected)
		h.Frame()
		for _, actual := range buttons {
			p.SetValue(false)
			h.Frame()
			before := changes
			press(actual)
			want := actual == selected
			if p.Value() != want || changes-before != map[bool]int{true: 1, false: 0}[want] {
				t.Fatalf("selected %v actual %v: open %v changes %v", selected, actual, p.Value(), changes-before)
			}
			if want {
				press(actual)
				if p.Value() || changes-before != 2 {
					t.Fatal("duplicate or missing toggle")
				}
			}
		}
	}
	if primary == 0 {
		t.Fatal("primary handler swallowed")
	}
	p.MouseButton(pointer.ButtonPrimary | pointer.ButtonSecondary)
	if p.mouseButton != pointer.ButtonTertiary {
		t.Fatal("invalid mask accepted")
	}
	p.MouseButton(0)
	h.Frame()
	before := changes
	for _, b := range buttons {
		press(b)
	}
	if p.Value() || changes != before {
		t.Fatal("manual mode retained handler")
	}
	p.MouseButton(pointer.ButtonTertiary)
	h.Frame()
	press(pointer.ButtonTertiary | pointer.ButtonSecondary)
	if p.Value() {
		t.Fatal("chord triggered")
	}
	p.SetDisabled(true)
	h.Frame()
	press(pointer.ButtonTertiary)
	if p.Value() {
		t.Fatal("disabled middle trigger")
	}
}
