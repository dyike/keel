package kit

import (
	"fmt"
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"math"
	"testing"
)

func TestTabsScrollableRevealAndMaxWidth(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := Tabs().Scrollable(true).MaxWidth(90)
		for i := range 12 {
			v.Add(fmt.Sprintf("Long tab %02d", i), text(fmt.Sprintf("Page %02d", i)))
		}
		v.SetItemDisabled(10, true)
		calls := 0
		v.OnChange(func(int) { calls++ })
		var cx *el.Context
		h := renderView(viewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }), 260, scale)
		for range 4 {
			h.Frame()
		}
		b := bounds(h, "Long tab 00")
		if b.Dx() > 90*scale {
			t.Fatal("max width", b)
		}
		click(t, h, "Long tab 00")
		h.Key(key.NameEnd, 0)
		for range 4 {
			h.Frame()
		}
		off, view, content := v.ScrollState(cx)
		b = bounds(h, "Long tab 11")
		if v.Value() != 11 || calls != 1 || off <= 0 || content <= view || b.Min.X < 0 || b.Max.X > 260*scale {
			t.Fatalf("reveal scale=%d offset=%v view=%v content=%v bounds=%v value=%d", scale, off, view, content, b, v.Value())
		}
		h.Key(key.NameLeftArrow, 0)
		for range 3 {
			h.Frame()
		}
		if v.Value() != 9 {
			t.Fatal("skip disabled")
		}
		before := calls
		v.ScrollTo(0)
		for range 4 {
			h.Frame()
		}
		off, _, _ = v.ScrollState(cx)
		if off != 0 || calls != before || v.Value() != 9 {
			t.Fatal("program scroll selects or fails", off)
		}
		v.SetValue(11)
		for range 4 {
			h.Frame()
		}
		off, _, _ = v.ScrollState(cx)
		if off <= 0 {
			t.Fatal("program selection not revealed")
		}
		v.Scrollable(false)
		for range 4 {
			h.Frame()
		}
		if !shown(h, "更多") || !shown(h, "Long tab 11") {
			t.Fatal("overflow restore")
		}
		v.MaxWidth(float32(math.NaN())).MaxWidth(-1)
		if v.maxWidth != 90 {
			t.Fatal("invalid cap")
		}
		v.MaxWidth(0)
	}
}

func TestTabsScrollRequestBeforeMountAndMove(t *testing.T) {
	v := Tabs().Scrollable(true)
	for i := range 8 {
		v.Add(fmt.Sprintf("Tab %d", i), nil)
	}
	v.ScrollTo(7)
	var cx *el.Context
	h := renderView(viewFunc(func(c *el.Context) el.Element { cx = c; return v.Render(c) }), 150, 1)
	for range 4 {
		h.Frame()
	}
	off, _, _ := v.ScrollState(cx)
	if off <= 0 || v.Value() != 0 {
		t.Fatal("initial scroll request lost", off)
	}
	v.ScrollTo(7)
	v.Move(7, 0)
	for range 4 {
		h.Frame()
	}
	off, _, _ = v.ScrollState(cx)
	if off != 0 {
		t.Fatal("request did not follow stable identity", off)
	}
}
