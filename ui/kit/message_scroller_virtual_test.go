package kit

import (
	"fmt"
	"image"
	"testing"

	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestMessageScrollerVirtualFollowAndReadingAnchor(t *testing.T) {
	for _, scale := range []int{1, 2} {
		keys := variableKeys(10000)
		heights := map[string]float32{}
		built := 0
		sc := MessageScroller(keys, 70, func(cx *el.Context, i int) el.Element {
			built++
			h := float32(50)
			if n, ok := heights[keys[i]]; ok {
				h = n
			}
			return el.Div().H(el.Dp(h)).Name("message-" + keys[i])
		})
		var cx *el.Context
		root := el.Root(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().Child(sc.Render(ctx)) }))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
			root.Layout(gtx)
		})
		settle(h)
		off, view, content := cx.ScrollState(sc.list.ID())
		if content-off-view > 1 || !shown(h, "message-9999") || built > 300 {
			t.Fatalf("initial end %v %v %v built=%d", off, view, content, built)
		}
		heights["9999"] = 350
		settle(h)
		off, view, content = cx.ScrollState(sc.list.ID())
		if content-off-view > 1 {
			t.Fatal("streaming height lost bottom")
		}
		keys = append(keys, "new")
		sc.SetKeys(keys)
		settle(h)
		if !shown(h, "message-new") {
			t.Fatal("new message not followed")
		}
		h.Scroll(float32(200*scale), float32(150*scale), -800*float32(scale))
		settle(h)
		anchor := sc.list.anchor
		before := bounds(h, "message-"+anchor).Min.Y
		if anchor == "" || !shown(h, "回到最新") {
			t.Fatal("scroll up did not stop follow")
		}
		heights["new"] = 500
		keys = append(keys, "newer")
		sc.SetKeys(keys)
		settle(h)
		if after := bounds(h, "message-"+anchor).Min.Y; after != before {
			t.Fatalf("append moved reader %s %d -> %d", anchor, before, after)
		}
		// Prepending unknown heights retains the stable visible key and offset.
		keys = append([]string{"old-1", "old-2"}, keys...)
		heights["old-1"] = 220
		sc.SetKeys(keys)
		settle(h)
		if after := bounds(h, "message-"+anchor).Min.Y; after != before {
			t.Fatalf("history moved reader %d -> %d", before, after)
		}
		index := sc.list.indices[anchor]
		above := keys[index-1]
		heights[above] = 250
		settle(h)
		if after := bounds(h, "message-"+anchor).Min.Y; after != before {
			t.Fatalf("image above reader moved anchor %d -> %d", before, after)
		}
		click(t, h, "回到最新")
		settle(h)
		off, view, content = cx.ScrollState(sc.list.ID())
		if content-off-view > 1 || !shown(h, "message-newer") {
			t.Fatal("explicit latest")
		}
	}
}
func TestMessageScrollerTopCallbackAndDisabled(t *testing.T) {
	keys := variableKeys(100)
	calls := 0
	sc := MessageScroller(keys, 60, func(_ *el.Context, i int) el.Element { return el.Div().H(el.Dp(40)).Name(fmt.Sprintf("message-%d", i)) }).OnReachTop(func() { calls++ })
	sc.SetFollow(false)
	h := uitest.New(el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return sc.Render(cx) })))
	settle(h)
	if calls != 1 {
		t.Fatalf("initial top callbacks %d", calls)
	}
	settle(h)
	if calls != 1 {
		t.Fatal("repeated top callback while idle")
	}
	sc.SetDisabled(true)
	sc.ScrollToEnd()
	settle(h)
	h.Scroll(200, 150, -10000)
	settle(h)
	if calls != 1 {
		t.Fatal("disabled top callback")
	}
}

func TestMessageScrollerFirstFrameFillsTallViewport(t *testing.T) {
	keys := variableKeys(10000)
	sc := MessageScroller(keys, 80, func(_ *el.Context, i int) el.Element {
		return el.Div().H(el.Dp(60)).Name("message-" + keys[i])
	})
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(sc.Render(cx)) }))
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Constraints.Max = image.Pt(680, 1040)
		root.Layout(gtx)
	})
	h.Frame()
	// A 320dp fallback would leave most of this viewport filled by the leading
	// virtual spacer until another frame. The initial range must cover it.
	if !shown(h, "message-9988") || !shown(h, "message-9999") {
		t.Fatal("initial range left blank space in a tall conversation")
	}
}
