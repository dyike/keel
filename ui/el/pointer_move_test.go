package el

import (
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestPointerMoveLocalDPAndClick(t *testing.T) {
	for _, scale := range []float32{1, 1.25, 2} {
		var x, y float32
		calls, clicks := 0, 0
		disabled := false
		root := Root(ViewFunc(func(cx *Context) Element {
			return Div().P(20).Child(Div().Translate(4, 8).Size(Dp(80)).Disabled(disabled).OnPointerMove(func(a, b float32) { x, y = a, b; calls++ }).Child(Div().Size(Dp(40)).OnClick(func() { clicks++ })))
		}))
		h := uitest.NewFunc(func(gtx core.C) { gtx.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}; root.Layout(gtx) })
		h.Move(34*scale, 38*scale)
		if calls == 0 || x != 10 || y != 10 {
			t.Fatalf("scale %v: calls=%d at %v,%v", scale, calls, x, y)
		}
		h.Click(34*scale, 38*scale)
		if clicks != 1 {
			t.Fatal("observer consumed child click")
		}
		disabled = true
		h.Frame()
		h.Frame()
		before := calls
		h.Move(35*scale, 39*scale)
		if calls != before {
			t.Fatal("disabled callback")
		}
	}
}

func TestPointerMoveModalIsolation(t *testing.T) {
	open := false
	calls := 0
	root := Root(ViewFunc(func(cx *Context) Element {
		if open {
			cx.Overlay("modal", Modal(Div().Size(Dp(60))))
		}
		return Div().Size(Dp(100)).OnPointerMove(func(x, y float32) { calls++ })
	}))
	h := uitest.New(root)
	h.Move(20, 20)
	if calls == 0 {
		t.Fatal("missing motion")
	}
	open = true
	h.Frame()
	h.Frame()
	before := calls
	h.Move(21, 21)
	if calls != before {
		t.Fatal("modal leaked motion")
	}
}
