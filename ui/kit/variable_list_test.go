package kit

import (
	"fmt"
	"image"
	"strconv"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func variableKeys(count int) []string {
	out := make([]string, count)
	for i := range out {
		out[i] = strconv.Itoa(i)
	}
	return out
}
func settle(h *uitest.Harness) {
	for range 8 {
		h.Frame()
	}
}

func TestVariableListMeasurementsRevealAndVirtualization(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			var cx *el.Context
			built := 0
			list := VariableList(variableKeys(100000), 40, func(cx *el.Context, i int) el.Element {
				built++
				return el.Div().H(el.Dp(float32(20 + i%4*20))).Name(fmt.Sprintf("row-%d", i))
			}).Height(200)
			root := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(list.Render(c)) }))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(300*scale, 400*scale)
				root.Layout(gtx)
			})
			settle(h)
			if list.measured["0"] != 20 || list.measured["3"] != 80 {
				t.Fatalf("measurements %v", list.measured)
			}
			if built > 300 {
				t.Fatalf("built %d rows for initial viewport", built)
			}
			list.ScrollTo(cx, 99999)
			settle(h)
			if r := bounds(h, "row-99999"); r.Empty() || r.Min.Y < 0 || r.Max.Y > 200*scale {
				t.Fatalf("last row not revealed: %v", r)
			}
			if list.reveal != "" {
				t.Fatal("reveal did not settle")
			}
			if built > 700 {
				t.Fatalf("built %d rows after reveal", built)
			}
			list.SetKeys(variableKeys(2))
			settle(h)
			if off, _, total := cx.ScrollState(list.ID()); off != 0 || total != 60 {
				t.Fatalf("shrink %g %g", off, total)
			}
			list.SetKeys(nil)
			settle(h)
			if off, _, total := cx.ScrollState(list.ID()); off != 0 || total != 0 {
				t.Fatalf("empty %g %g", off, total)
			}
			list.SetKeys(variableKeys(10))
			settle(h)
			if bounds(h, "row-0").Empty() {
				t.Fatal("refill failed")
			}
		})
	}
}

func TestVariableListKeepsAnchorOnPrependAndHeightChange(t *testing.T) {
	keys := variableKeys(100)
	heights := map[string]float32{}
	var cx *el.Context
	var list *VariableListView
	list = VariableList(keys, 40, func(cx *el.Context, i int) el.Element {
		k := list.keys[i]
		height := float32(40)
		if h, ok := heights[k]; ok {
			height = h
		}
		return el.Div().H(el.Dp(height)).Name("row-" + k)
	}).Height(200)
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(300)).Child(list.Render(c)) })
	settle(h)
	list.ScrollTo(cx, 20)
	settle(h)
	anchor := list.anchor
	before := bounds(h, "row-"+anchor).Min.Y
	newKeys := append([]string{"older-a", "older-b"}, keys...)
	list.SetKeys(newKeys)
	settle(h)
	if got := bounds(h, "row-"+anchor).Min.Y; got != before {
		t.Fatalf("prepend moved anchor %s: %d to %d", anchor, before, got)
	}
	i := list.indices[anchor]
	above := list.keys[i-1]
	heights[above] = 120
	settle(h)
	if got := bounds(h, "row-"+anchor).Min.Y; got != before {
		t.Fatalf("height change moved anchor %s: %d to %d", anchor, before, got)
	}
	if list.measured[above] != 120 {
		t.Fatal("new height not measured")
	}
	// Offscreen invalidation changes the estimate above the anchor too.
	heights["0"] = 200
	list.Invalidate("0")
	settle(h)
	if got := bounds(h, "row-"+anchor).Min.Y; got != before {
		t.Fatalf("invalidate moved anchor: %d to %d", before, got)
	}
	// SetKeys owns a copy and rejects duplicate identities atomically.
	newKeys[0] = "mutated"
	if list.keys[0] != "older-a" {
		t.Fatal("retained caller keys")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("duplicate accepted")
			}
		}()
		list.SetKeys([]string{"x", "x"})
	}()
	if list.Count() != 102 {
		t.Fatal("invalid keys changed list")
	}
}

func TestVariableListWidthAndReadOnly(t *testing.T) {
	var cx *el.Context
	width := 300
	readonly := false
	list := VariableList(variableKeys(200), 40, func(cx *el.Context, i int) el.Element {
		return el.Text(strings.Repeat("中英 wrapped text ", 3)).Name(fmt.Sprint(i))
	}).Height(180)
	root := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(list.Render(c)) }))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints.Max = image.Pt(width, 300)
		if readonly {
			gtx = gtx.Disabled()
		}
		root.Layout(gtx)
	})
	settle(h)
	old := list.measured["0"]
	list.ScrollTo(cx, 30)
	settle(h)
	anchor := list.anchor
	before := bounds(h, anchor).Min.Y
	width = 150
	settle(h)
	if list.measured[anchor] <= old {
		t.Fatal("width change did not remeasure wrapped rows")
	}
	if got := bounds(h, anchor).Min.Y; got != before {
		t.Fatalf("width moved anchor: %d to %d", before, got)
	}
	off, _, _ := cx.ScrollState(list.ID())
	readonly = true
	settle(h)
	readonly = false
	settle(h)
	if got, _, _ := cx.ScrollState(list.ID()); got != off {
		t.Fatalf("readonly %g to %g", off, got)
	}
}

func TestVariableListWheelAndClick(t *testing.T) {
	var cx *el.Context
	selected := -1
	list := VariableList(variableKeys(1000), 40, func(cx *el.Context, i int) el.Element {
		return el.Div().H(el.Dp(float32(30 + i%3*10))).Name(fmt.Sprintf("item-%d", i)).OnClick(func() { selected = i })
	}).Height(180)
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(300)).Child(list.Render(c)) })
	settle(h)
	h.Scroll(50, 80, 95)
	settle(h)
	if off, _, _ := cx.ScrollState(list.ID()); off != 95 {
		t.Fatalf("wheel %g", off)
	}
	if list.anchor == "0" {
		t.Fatal("wheel retained old anchor")
	}
	index := list.indices[list.anchor] + 1
	click(t, h, fmt.Sprintf("item-%d", index))
	if selected != index {
		t.Fatalf("clicked %d instead of %d", selected, index)
	}
}

func BenchmarkVariableListFrame(b *testing.B) {
	for _, count := range []int{1000, 100000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			list := VariableList(variableKeys(count), 40, func(cx *el.Context, i int) el.Element {
				return el.Div().H(el.Dp(float32(30 + i%3*10))).Child(el.Text("row"))
			}).Height(200)
			h := render(func(cx *el.Context) el.Element { return list.Render(cx) })
			settle(h)
			b.ResetTimer()
			for b.Loop() {
				h.Frame()
			}
		})
	}
}

func TestVariableListRevealKeyAfterPrependAndReorder(t *testing.T) {
	for _, scale := range []int{1, 2} {
		t.Run(fmt.Sprint(scale), func(t *testing.T) {
			keys := make([]string, 100000)
			for i := range keys {
				keys[i] = strconv.Itoa(i + 1)
			}
			var cx *el.Context
			list := VariableList(keys, 72, func(_ *el.Context, i int) el.Element {
				n, _ := strconv.Atoi(keys[i])
				return el.Div().H(el.Dp(float32(40 + n%3*30))).Name("message-" + keys[i])
			}).Height(200)
			root := el.Root(viewFunc(func(c *el.Context) el.Element { cx = c; return el.Div().Child(list.Render(c)) }))
			h := uitest.NewFunc(func(gtx core.C) {
				gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
				gtx.Constraints.Max = image.Pt(300*scale, 400*scale)
				root.Layout(gtx)
			})
			check := func() {
				t.Helper()
				settle(h)
				r := bounds(h, "message-50000")
				if r.Empty() || r.Min.Y < 0 || r.Max.Y > 200*scale {
					t.Fatalf("message 50000 not visible after key reveal: %v", r)
				}
				if list.reveal != "" {
					t.Fatal("reveal did not settle")
				}
			}
			list.ScrollToKey(cx, "50000")
			check()
			for batch := range 2 {
				add := make([]string, 10)
				for i := range add {
					add[i] = fmt.Sprintf("older-%d-%d", batch, i)
				}
				keys = append(add, keys...)
				list.SetKeys(keys)
				list.ScrollTo(cx, 0)
				settle(h)
				list.ScrollToKey(cx, "50000")
				list.ScrollToKey(cx, "missing")
				if list.reveal != "50000" {
					t.Fatal("missing key replaced pending reveal")
				}
				check()
			}
			// A pending key request must survive reordering before the next frame.
			list.ScrollToKey(cx, "50000")
			i := list.indices["50000"]
			keys[10], keys[i] = keys[i], keys[10]
			list.SetKeys(keys)
			check()
			off, _, _ := cx.ScrollState(list.ID())
			list.ScrollToKey(cx, "")
			list.ScrollToKey(cx, "missing")
			settle(h)
			if got, _, _ := cx.ScrollState(list.ID()); got != off {
				t.Fatalf("missing key moved viewport: %g to %g", off, got)
			}
		})
	}
}
