package kit

import (
	"fmt"
	"image"
	"strconv"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestAlignedOffset(t *testing.T) {
	// 300 tall viewport over 3000 of content, a row at [1500, 1530).
	for align, want := range map[ScrollAlign]float32{ScrollStart: 1500, ScrollCenter: 1365, ScrollEnd: 1230, ScrollNearest: 1230} {
		if got := alignedOffset(align, 0, 300, 3000, 1500, 1530); got != want {
			t.Errorf("align %d: %v, want %v", align, got, want)
		}
	}
	if got := alignedOffset(ScrollCenter, 0, 300, 3000, 0, 30); got != 0 {
		t.Fatal("clamped at the start", got)
	}
	if got := alignedOffset(ScrollStart, 0, 300, 3000, 2990, 3000); got != 2700 {
		t.Fatal("clamped at the end", got)
	}
}

func TestVirtualListsScrollToAlign(t *testing.T) {
	var cx *el.Context
	vl := VirtualList(100, 30, func(cx *el.Context, i int) el.Element { return el.Text("v" + strconv.Itoa(i)) }).Height(300)
	vroot := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(vl.Render(c)) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(300, 400); vroot.Layout(gtx) })
	settle(h)
	for align, want := range map[ScrollAlign]int{ScrollStart: 15, ScrollCenter: 150, ScrollEnd: 285} {
		vl.ScrollToAlign(cx, 50, align)
		settle(h)
		if r := bounds(h, "v50"); (r.Min.Y+r.Max.Y)/2 != want {
			t.Fatalf("align %d: row at %v, want center %d", align, r, want)
		}
	}

	keys := variableKeys(200)
	var vcx *el.Context
	vlist := VariableList(keys, 40, func(cx *el.Context, i int) el.Element {
		return el.Div().H(el.Dp(float32(20 + i%3*20))).Name(fmt.Sprintf("row-%d", i))
	}).Height(200)
	root := el.Root(viewFunc(func(c *el.Context) el.Element { vcx = c; return el.Div().Child(vlist.Render(c)) }))
	vh := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(300, 400); root.Layout(gtx) })
	settle(vh)
	vlist.ScrollToAlign(vcx, 120, ScrollCenter)
	settle(vh)
	r := bounds(vh, "row-120")
	if mid := (r.Min.Y + r.Max.Y) / 2; mid < 95 || mid > 105 {
		t.Fatalf("variable list center: row at %v", r)
	}
	vlist.ScrollToAlign(vcx, 120, ScrollStart)
	settle(vh)
	if r := bounds(vh, "row-120"); r.Min.Y != 0 {
		t.Fatalf("variable list start: row at %v", r)
	}
	vlist.ScrollToAlign(vcx, 120, ScrollEnd)
	settle(vh)
	if r := bounds(vh, "row-120"); r.Max.Y != 200 {
		t.Fatalf("variable list end: row at %v", r)
	}
}
