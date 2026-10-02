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

func TestDialogNestedMenuAndDialogFocus(t *testing.T) {
	inner := Dialog("Inner").Body(text("Inner body"))
	m := Menu().Item("Nested command", "", nil)
	m.Trigger(Button("Menu", m.Toggle))
	outer := Dialog("Outer").Body(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Gap(8).Child(m.Render(cx), Button("Inner open", func() { inner.SetValue(true) }).ID("inner-open").Render(cx), inner.Render(cx))
	}))
	var cx *el.Context
	h := uitest.New(el.Root(el.ViewFunc(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().Child(Button("Open", func() { outer.SetValue(true) }).ID("outer-open").Render(ctx), outer.Render(ctx))
	})))
	click(t, h, "Open")
	click(t, h, "Menu")
	h.Frame()
	if !m.Value() || !shown(h, "Nested command") {
		t.Fatal("nested menu missing")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !outer.Value() || m.Value() {
		t.Fatal("nested menu Esc")
	}
	click(t, h, "Inner open")
	h.Frame()
	if !inner.Value() || !shown(h, "Inner body") {
		t.Fatal("nested modal missing")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if inner.Value() || !outer.Value() || !cx.Focused("inner-open") {
		t.Fatal("nested modal focus return")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if outer.Value() || !cx.Focused("outer-open") {
		t.Fatal("outer focus return")
	}
}
func TestDialogOwnerDisableAndHiddenCloseOnce(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		calls := 0
		d := Dialog("Owned").Body(text("Body")).OnClose(func() { calls++ })
		d.SetValue(true)
		blocked := false
		h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Disabled(blocked && !hidden).Hidden(blocked && hidden).Child(d.Render(cx))
		})))
		h.Frame()
		if !shown(h, "Body") {
			t.Fatal("modal missing before disable")
		}
		blocked = true
		h.Frame()
		h.Frame()
		h.Key(key.NameEscape, 0)
		h.Frame()
		if d.Value() || shown(h, "Body") || calls != 1 {
			t.Fatalf("owner change: open=%v calls=%d", d.Value(), calls)
		}
		blocked = false
		d.SetDisabled(true)
		d.SetValue(true)
		d.Confirm("No", "No", nil)
		h.Frame()
		if d.Value() || calls != 1 {
			t.Fatal("disabled dialog reopened or notified")
		}
	}
}
func TestDialogTallBodyBoundsScrollAndFooterOwnership(t *testing.T) {
	d := Dialog("Long").Width(1000).Body(el.ViewFunc(func(*el.Context) el.Element { return el.Div().H(el.Dp(1000)).Child(el.Text("Start")) }))
	footer := []el.View{Button("Save", nil)}
	d.Footer(footer...)
	footer[0] = nil
	d.Width(float32(math.Inf(1)))
	d.Width(float32(math.NaN()))
	if d.width != 1000 || d.footer[0] == nil {
		t.Fatal("invalid width or footer alias")
	}
	d.SetValue(true)
	var cx *el.Context
	root := el.Root(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return d.Render(ctx) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(240, 220); root.Layout(gtx) })
	h.Frame()
	n, ok := semanticNode(h, "dialog")
	if !ok || n.Desc.Bounds.Min.X < 0 || n.Desc.Bounds.Max.X > 240 || n.Desc.Bounds.Max.Y > 220 || !shown(h, "Save") {
		t.Fatalf("dialog bounds %v", n.Desc.Bounds)
	}
	_, height, content := cx.ScrollState(autoID("dialog", d) + "/body")
	if height <= 0 || content <= height {
		t.Fatalf("body not bounded: height=%v content=%v", height, content)
	}
	cx.ScrollTo(autoID("dialog", d)+"/body", 200)
	h.Frame()
	offset, _, _ := cx.ScrollState(autoID("dialog", d) + "/body")
	if offset <= 0 {
		t.Fatal("body did not scroll")
	}
}
