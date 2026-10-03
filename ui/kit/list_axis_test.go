package kit

import (
	"fmt"
	"github.com/dyike/keel/ui/el"
	"strconv"
	"testing"
)

func TestVirtualListHorizontalRevealSwitchAndShrink(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			built := 0
			v := VirtualList(100000, 64, func(cx *el.Context, i int) el.Element {
				built++
				return el.Div().Name(fmt.Sprintf("cell-%d", i)).Child(el.Text(strconv.Itoa(i)))
			}).Horizontal(true).Width(256).Height(96)
			var cx *el.Context
			h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return v.Render(ctx) }), 300, scale)
			settle(h)
			if built > 180 {
				t.Fatal("eager horizontal construction", built)
			}
			v.ScrollToEnd(cx)
			settle(h)
			b := bounds(h, "cell-99999")
			if b.Empty() || b.Min.X < 0 || b.Max.X > 256*scale {
				t.Fatal("end outside viewport", b)
			}
			off, view, total := cx.ScrollStateX(v.ID())
			if view != 256 || total != 6400000 || off <= 0 {
				t.Fatal("horizontal extent", off, view, total)
			}
			anchor := int(off / 64)
			v.Horizontal(false)
			settle(h)
			if !shown(h, fmt.Sprintf("cell-%d", anchor)) {
				t.Fatal("axis switch lost leading item", anchor)
			}
			v.Horizontal(true)
			settle(h)
			v.SetCount(2)
			settle(h)
			if off, _, total := cx.ScrollStateX(v.ID()); off != 0 || total != 128 {
				t.Fatal("shrink retained offset", off, total)
			}
			v.SetCount(0)
			settle(h)
			v.SetCount(4)
			settle(h)
			if !shown(h, "cell-0") {
				t.Fatal("refill missing")
			}
		})
	}
}

func TestVariableListHorizontalMeasurementsAnchorAndSwitch(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			keys := variableKeys(100000)
			built := 0
			v := VariableList(keys, 80, func(cx *el.Context, i int) el.Element {
				built++
				n, _ := strconv.Atoi(keys[i])
				return el.Div().Name("cell-" + keys[i]).W(el.Dp(float32(60 + n%3*30))).H(el.Dp(40)).Child(el.Text(keys[i]))
			}).Horizontal(true).Width(240).Height(80)
			var cx *el.Context
			h := renderView(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return v.Render(ctx) }), 300, scale)
			settle(h)
			if v.measured["0"] != 60 || v.measured["2"] != 120 || built > 250 {
				t.Fatal("horizontal measurements", v.measured, built)
			}
			v.ScrollTo(cx, 99999)
			settle(h)
			if b := bounds(h, "cell-99999"); b.Empty() || b.Min.X < 0 || b.Max.X > 240*scale || v.reveal != "" {
				t.Fatal("variable end reveal", b, v.reveal)
			}
			v.ScrollTo(cx, 20)
			settle(h)
			anchor := v.anchor
			before := bounds(h, "cell-"+anchor).Min.X
			keys = append([]string{"new-a", "new-b"}, keys...)
			v.SetKeys(keys)
			settle(h)
			if after := bounds(h, "cell-"+anchor).Min.X; after != before {
				t.Fatal("prepend moved reading position", before, after)
			}
			v.Horizontal(false)
			settle(h)
			if !shown(h, "cell-"+anchor) || v.measured[anchor] != 40 {
				t.Fatal("old width used as height", anchor, v.measured[anchor])
			}
			v.Horizontal(true)
			settle(h)
			if v.measured[anchor] == 40 {
				t.Fatal("old height used as width")
			}
			keys = keys[:2]
			v.SetKeys(keys)
			settle(h)
			if off, _, _ := cx.ScrollStateX(v.ID()); off != 0 {
				t.Fatal("shrink offset", off)
			}
			if built > 1800 {
				t.Fatal("not virtual", built)
			}
		})
	}
}

func TestHorizontalListKeepsChildStateOnAxisSwitch(t *testing.T) {
	horizontal := true
	v := VirtualList(4, 100, func(cx *el.Context, i int) el.Element { return el.Input().Name(fmt.Sprintf("edit-%d", i)) }).Horizontal(true).Width(300).Height(120)
	h := page(v)
	clickClass(t, h, "Editor", "edit-1")
	h.Type("draft")
	h.Frame()
	horizontal = !horizontal
	v.Horizontal(horizontal)
	settle(h)
	if desc(h, "edit-1") != "draft" {
		t.Fatal("editor state lost on axis switch")
	}
}

func TestHorizontalFillAndPendingReveal(t *testing.T) {
	v := VirtualList(10000, 80, func(_ *el.Context, i int) el.Element { return el.Text(fmt.Sprintf("item-%d", i)) }).Horizontal(true).Height(60).Fill()
	once := true
	h := renderView(el.ViewFunc(func(cx *el.Context) el.Element {
		if once {
			v.ScrollTo(cx, 9999)
			once = false
		}
		return el.Div().W(el.Dp(320)).H(el.Dp(80)).Row().Child(v.Render(cx))
	}), 320, 1)
	settle(h)
	if !shown(h, "item-9999") {
		t.Fatal("pending reveal before first mount")
	}
}
