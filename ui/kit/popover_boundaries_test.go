package kit

import (
	"image"
	"math"
	"testing"

	"gioui.org/io/key"
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
