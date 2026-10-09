package el

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestScrollbarDragTrackKeysAndDisable(t *testing.T) {
	for _, horizontal := range []bool{true, false} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("horizontal=%v/scale=%d", horizontal, scale), func(t *testing.T) {
				var cx *Context
				disabled, readonly := false, false
				length := float32(400)
				clicks := 0
				root := Root(viewFunc(func(c *Context) Element {
					cx = c
					child := Div().W(Dp(100)).H(Dp(length)).OnClick(func() { clicks++ })
					box := Div().ID("scroll").W(Dp(100)).H(Dp(100)).Focusable(true).Disabled(disabled)
					if horizontal {
						box.ScrollX()
						child.W(Dp(length)).H(Dp(100))
					} else {
						box.ScrollY()
					}
					return Div().Items(Start).Child(box.Child(child))
				}))
				h := uitest.NewFunc(func(gtx core.C) {
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints.Max = image.Pt(200*scale, 200*scale)
					if readonly {
						gtx = gtx.Disabled()
					}
					root.Layout(gtx)
				})
				state := func() float32 {
					if horizontal {
						x, _, _ := cx.ScrollStateX("scroll")
						return x
					}
					y, _, _ := cx.ScrollState("scroll")
					return y
				}
				point := func(main float32) (float32, float32) {
					if horizontal {
						return main * float32(scale), 95 * float32(scale)
					}
					return 95 * float32(scale), main * float32(scale)
				}
				a, b := point(12)
				c, d := point(62)
				h.Drag(a, b, c, d)
				h.Frame()
				if got := state(); got < 195 || got > 205 {
					t.Fatalf("drag offset %g", got)
				}
				if clicks != 0 {
					t.Fatal("scrollbar activated underlying content")
				}
				a, b = point(99)
				h.Click(a, b)
				h.Frame()
				if got := state(); got != 300 {
					t.Fatalf("track end %g", got)
				}
				if !cx.Focused("scroll") {
					t.Fatal("bar did not focus scroll viewport")
				}
				h.Key(key.NameHome, 0)
				h.Frame()
				if got := state(); got != 0 {
					t.Fatalf("Home %g", got)
				}
				arrow := key.NameDownArrow
				if horizontal {
					arrow = key.NameRightArrow
				}
				h.Key(arrow, 0)
				h.Frame()
				if got := state(); got != 40 {
					t.Fatalf("arrow %g", got)
				}
				h.Key(key.NamePageDown, 0)
				h.Frame()
				if got := state(); got != 140 {
					t.Fatalf("page %g", got)
				}
				readonly = true
				h.Frame()
				h.Frame()
				readonly = false
				h.Frame()
				if got := state(); got != 140 {
					t.Fatalf("readonly changed offset %g", got)
				}
				disabled = true
				h.Frame()
				a, b = point(99)
				h.Click(a, b)
				h.Key(key.NameEnd, 0)
				h.Frame()
				if got := state(); got != 140 {
					t.Fatalf("disabled changed offset %g", got)
				}
				disabled = false
				length = 50
				h.Frame()
				h.Frame()
				if got := state(); got != 0 {
					t.Fatalf("shrink offset %g", got)
				}
			})
		}
	}
}

func TestScrollbarShortGeometry(t *testing.T) {
	for _, horizontal := range []bool{false, true} {
		for _, length := range []int{2, 8, 20, 100} {
			track := image.Rect(0, 0, 10, length)
			if horizontal {
				track = image.Rect(0, 0, length, 10)
			}
			start := scrollbarThumb(track, horizontal, 0, length, 1000, 24)
			end := scrollbarThumb(track, horizontal, 1000-length, length, 1000, 24)
			if start.Empty() || end.Empty() || !start.In(track) || !end.In(track) || start == end {
				t.Fatalf("track %v start %v end %v", track, start, end)
			}
		}
	}
}

func TestScrollKeysPreserveGlobalShortcut(t *testing.T) {
	calls := 0
	var cx *Context
	h := uitest.New(Root(viewFunc(func(c *Context) Element {
		cx = c
		cx.Shortcut("mod+n", func() { calls++ })
		return Div().Items(Start).Child(Div().ScrollY().W(Dp(100)).H(Dp(100)).Child(Div().ID("button").Focusable(true).H(Dp(400))))
	})))
	cx.Focus("button")
	h.Frame()
	h.Frame()
	h.Key("N", key.ModShortcut)
	if calls != 1 {
		t.Fatal("scroll key filters swallowed global shortcut")
	}
}
