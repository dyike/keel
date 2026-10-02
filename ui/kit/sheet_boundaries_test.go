package kit

import (
	"math"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
)

func TestSheetNestedPopoverAndOwnerDisable(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	p := Popover(text("Nested content"))
	p.Trigger(Button("Popover", p.Toggle))
	calls := 0
	s := Sheet(el.Right, "Details").Body(p).OnClose(func() { calls++ })
	disabled := false
	h := c.harness(func(cx *el.Context) el.Element {
		return el.Div().Disabled(disabled).Child(Button("Open", func() { s.SetValue(true) }).Render(cx), s.Render(cx))
	})
	click(t, h, "Open")
	c.advance(h, SheetSlide)
	click(t, h, "Popover")
	h.Frame()
	if !p.Value() || !shown(h, "Nested content") {
		t.Fatal("nested overlay missing")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if p.Value() || !s.Value() || calls != 0 {
		t.Fatal("Esc dismissed wrong layer")
	}
	disabled = true
	h.Frame()
	h.Frame()
	if s.Value() || calls != 1 {
		t.Fatal("disabled owner retained sheet")
	}
	disabled = false
	s.SetDisabled(true)
	s.SetValue(true)
	h.Frame()
	if s.Value() || calls != 1 {
		t.Fatal("disabled sheet reopened or notified")
	}
}
func TestSheetAnimationUsesFittedSizeAndRestarts(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	s := Sheet(el.Right, "Details").Size(1000).Body(text("Body"))
	s.Size(float32(math.NaN()))
	s.Size(float32(math.Inf(1)))
	if s.size != 1000 {
		t.Fatal("invalid size")
	}
	s.SetValue(true)
	h := c.harness(func(cx *el.Context) el.Element { return s.Render(cx) })
	c.advance(h, SheetSlide/2)
	n, ok := semanticNode(h, "dialog")
	// A 400dp-wide viewport fits a 400dp sheet; halfway through ease-out the
	// remaining shift is 100dp, not 250dp from the requested 1000dp width.
	if !ok || n.Desc.Bounds.Min.X != 100 {
		t.Fatalf("fitted animation %v", n.Desc.Bounds)
	}
	c.advance(h, SheetSlide)
	s.SetValue(false)
	s.SetValue(true)
	h.Frame()
	if s.openedAt != c.now {
		t.Fatal("reopen without closed frame reused old animation")
	}
	c.advance(h, SheetSlide)
	n, _ = semanticNode(h, "dialog")
	if n.Desc.Bounds.Dx() != 400 || n.Desc.Bounds.Min.X != 0 {
		t.Fatal("fitted sheet size")
	}
}
