package kit

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestSidebarLargeTextAdaptation(t *testing.T) {
	for _, density := range []float32{1, 1.25, 2} {
		v := Sidebar().Collapsible(false).Section("", SidebarItem{ID: "one", Label: "One"}, SidebarItem{ID: "two", Label: "Two"})
		var width float32
		viewport := float32(1000)
		scale := float32(1.6364)
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			out := v.Render(cx)
			width = v.adapt(cx)
			return out
		}))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: density, PxPerSp: density * scale}
			gtx.Constraints.Max = image.Pt(int(viewport*density), int(600*density))
			root.Layout(gtx)
		})
		h.Frame()
		if width < 359 || width > 361 || v.Collapsed() {
			t.Fatalf("width=%v collapsed=%v", width, v.Collapsed())
		}
		if got := v.positions["two"]; got < 62 || got > 64 {
			t.Fatalf("scaled row offset=%v", got)
		}
		viewport = 400
		h.Frame()
		if !v.Collapsed() {
			t.Fatal("narrow viewport did not use icon rail")
		}
		viewport = 1000
		h.Frame()
		if v.Collapsed() {
			t.Fatal("automatic collapse persisted after resize")
		}
		v.SetCollapsed(true)
		h.Frame()
		if !v.Collapsed() {
			t.Fatal("explicit collapse lost")
		}
		v.SetCollapsed(false)
		v.Width(230)
		viewport = 400
		h.Frame()
		if width != 230 || v.Collapsed() {
			t.Fatal("explicit width changed")
		}
	}
}

func TestSidebarLargeTextKeyboardReveal(t *testing.T) {
	v := sidebarScrollFixture(100, 0).Height(280)
	root := el.Root(v)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Metric = unit.Metric{PxPerDp: 1.25, PxPerSp: 2.0455}
		gtx.Constraints.Max = image.Pt(900, 400)
		root.Layout(gtx)
	})
	h.Frame()
	click(t, h, "任务 0")
	h.Key(key.NameEnd, 0)
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if v.Value() != "99" || !shown(h, "任务 99") {
		t.Fatal("large-text last item not revealed")
	}
}

func TestSidebarTruncatedLabelTooltip(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	label := "A very long navigation destination that cannot fit"
	v := Sidebar().Width(180).Collapsible(false).Section("", SidebarItem{ID: "long", Label: label})
	h := c.harness(func(cx *el.Context) el.Element { return v.Render(cx) })
	x, y := center(bounds(h, label))
	h.Move(x, y)
	c.advance(h, 600*time.Millisecond)
	h.Frame()
	if n, ok := semanticNode(h, "tooltip"); !ok || n.Desc.Label != label {
		t.Fatal("missing full label tooltip")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if _, ok := semanticNode(h, "tooltip"); ok {
		t.Fatal("Escape did not dismiss label tooltip")
	}
	h.Move(390, 290)
	h.Frame()
	v.Width(900)
	h.Frame()
	x, y = center(bounds(h, label))
	h.Move(x, y)
	c.advance(h, 600*time.Millisecond)
	h.Frame()
	if _, ok := semanticNode(h, "tooltip"); ok {
		t.Fatal("untruncated label has a tooltip")
	}
}
