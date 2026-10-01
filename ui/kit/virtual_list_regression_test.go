package kit

import (
	"github.com/dyike/keel/ui/el"
	"strconv"
	"testing"
)

func TestVirtualListShrinkingAtEnd(t *testing.T) {
	vl := VirtualList(1000, 20, func(cx *el.Context, i int) el.Element { return el.Text("item " + strconv.Itoa(i)) }).Height(100)
	var cx *el.Context
	h := page(viewFunc(func(c *el.Context) el.Element { cx = c; return vl.Render(c) }))
	vl.ScrollTo(cx, 999)
	h.Frame()
	h.Frame()
	if !shown(h, "item 999") {
		t.Fatal("last row not revealed")
	}
	vl.SetCount(3)
	for range 4 {
		h.Frame()
	}
	if !shown(h, "item 0") || !shown(h, "item 2") {
		t.Fatal("shrinking left a blank scroll viewport")
	}
	vl.SetCount(0)
	h.Frame()
	h.Frame()
	vl.SetCount(2)
	h.Frame()
	h.Frame()
	if !shown(h, "item 0") {
		t.Fatal("empty list did not recover")
	}
}
