package kit

import (
	"image"
	"strconv"
	"testing"

	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestMessageScrollerNavigation(t *testing.T) {
	for _, scale := range []int{1, 2} {
		keys := variableKeys(1000)
		sc := MessageScroller(keys, 80, func(_ *el.Context, i int) el.Element {
			n, _ := strconv.Atoi(keys[i])
			return el.Div().H(el.Dp(float32(30 + n%4*20))).Name("message-" + keys[i])
		})
		var cx *el.Context
		root := el.Root(el.ViewFunc(func(ctx *el.Context) el.Element { cx = ctx; return sc.Render(ctx) }))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
			root.Layout(gtx)
		})
		sc.ScrollToEnd()
		if !sc.ScrollToMessage("500") || sc.ScrollToMessage("missing") {
			t.Fatal("ID validation")
		}
		settle(h)
		if !shown(h, "message-500") || !sc.IsScrolledUp(cx) || sc.IsFollowingTail(cx) || sc.list.reveal != "" {
			t.Fatal("initial message jump")
		}
		before := bounds(h, "message-500")
		sc.ScrollToMessage("500")
		settle(h)
		if bounds(h, "message-500") != before {
			t.Fatal("visible message moved")
		}
		keys = append([]string{"older"}, keys...)
		sc.SetKeys(keys)
		settle(h)
		if bounds(h, "message-500") != before {
			t.Fatal("history changed reading position")
		}
		sc.ScrollToMessage("200")
		sc.ScrollToEnd()
		settle(h)
		if !shown(h, "message-999") || sc.IsScrolledUp(cx) || !sc.IsFollowingTail(cx) {
			t.Fatal("end must win")
		}
		sc.ScrollToMessage("700")
		settle(h)
		if !shown(h, "message-700") || sc.IsFollowingTail(cx) {
			t.Fatal("jump from tail")
		}
		h.Scroll(float32(200*scale), float32(150*scale), 1000000*float32(scale))
		settle(h)
		if !sc.IsFollowingTail(cx) {
			t.Fatal("manual bottom resumes follow")
		}
		sc.SetFollow(false)
		sc.ScrollToMessage("300")
		settle(h)
		sc.ScrollToEnd()
		settle(h)
		if sc.IsScrolledUp(cx) || sc.IsFollowingTail(cx) {
			t.Fatal("explicit end with automatic follow disabled")
		}
	}
}
