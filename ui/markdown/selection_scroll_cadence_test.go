package markdown

import (
	"fmt"
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestSelectionGroupUpwardScrollHasSteadyCadence(t *testing.T) {
	for _, hz := range []int{60, 120} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("hz=%d/scale=%d", hz, scale), func(t *testing.T) {
				var group SelectionGroup
				docs := make([]*Doc, 40)
				roots := make([]*el.RootWidget, 40)
				for i := range docs {
					docs[i] = New(fmt.Sprintf("第%02d条消息正文", i))
					roots[i] = el.Embed(docs[i])
				}
				group.SetDocuments(docs...)
				var cxSaved *el.Context
				var offset float32
				root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
					cxSaved = cx
					list := el.Div().ID("scroll").H(el.Dp(200)).ScrollY().Gap(18)
					for _, r := range roots {
						list.Child(el.Widget(core.ViewportFunc(func(gtx core.C, visible image.Rectangle) core.D { return r.LayoutIn(gtx, visible, cx) })))
					}
					list.Decorate(func(gtx core.C, draw func()) { draw(); offset, _, _ = cx.ScrollState("scroll") })
					return group.Decorate(cx, el.Div().P(20).Child(list))
				}))
				now := time.Now()
				h := uitest.NewFunc(func(gtx core.C) {
					now = now.Add(time.Second / time.Duration(hz))
					gtx.Now = now
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
					root.Layout(gtx)
				})
				cxSaved.ScrollTo("scroll", 400)
				h.Frame()
				h.Frame()
				x, y := textPoint(t, h, docs[12], "第12条消息正文", 1)
				h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
				h.Frame()
				h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, float32(21*scale))})
				for range 4 {
					h.Frame()
				}
				minDelta, maxDelta := float32(1000), float32(-1000)
				before := offset
				for range 30 {
					last := offset
					h.Frame()
					delta := last - offset
					minDelta = min(minDelta, delta)
					maxDelta = max(maxDelta, delta)
				}
				expected := float32(396*30) / float32(hz)
				if distance := before - offset; distance < expected-1.1 || distance > expected+1.1 {
					t.Fatalf("scroll rate depends on refresh rate: %v dp, want %v", distance, expected)
				}
				if minDelta <= 0 || maxDelta-minDelta > 1.1 {
					t.Fatalf("upward drag stalls/jumps at %d Hz: steps %.2f..%.2f dp, offset %.2f -> %.2f", hz, minDelta, maxDelta, before, offset)
				}
			})
		}
	}
}

func TestSelectionGroupCanStartAnotherDrag(t *testing.T) {
	a, b := New("第一条消息"), New("第二条消息")
	var group SelectionGroup
	group.SetDocuments(a, b)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return group.Decorate(cx, el.Div().P(20).Gap(20).Child(a.Render(cx), b.Render(cx)))
	}))
	h := uitest.New(root)
	x0, y0 := textPoint(t, h, a, "第一条消息", 0)
	x1, y1 := textPoint(t, h, b, "第二条消息", 5)
	h.Drag(x0, y0, x1-1, y1)
	h.Frame()
	want := group.selection.selectedText()
	if want == "" {
		t.Fatal("first drag did not select text")
	}
	h.Drag(x1-1, y1, x0, y0)
	h.Frame()
	if got := group.selection.selectedText(); got != want {
		t.Fatalf("queued blur cancelled the next drag: got %q, want %q", got, want)
	}
}
