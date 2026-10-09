package kit

import (
	"fmt"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
	"time"
)

func TestNotifierEightPlacements(t *testing.T) {
	for _, scale := range []int{1, 2} {
		n := Notifier()
		n.Notify(Notice{Title: "Notice", Timeout: -1})
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) }))
		h := uitest.NewFunc(func(gtx core.C) {
			gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			gtx.Constraints.Max = image.Pt(900*scale, 600*scale)
			root.Layout(gtx)
		})
		for p := NoticeTopRight; p <= NoticeRightCenter; p++ {
			n.Placement(p)
			h.Frame()
			h.Frame()
			b := bounds(h, "Notice")
			x, y := 16*scale, 16*scale
			switch p {
			case NoticeTopRight, NoticeBottomRight, NoticeRightCenter:
				x = 900*scale - 16*scale - b.Dx()
			case NoticeTopCenter, NoticeBottomCenter:
				x = (900*scale - b.Dx()) / 2
			}
			switch p {
			case NoticeBottomLeft, NoticeBottomRight, NoticeBottomCenter:
				y = 600*scale - 16*scale - b.Dy()
			case NoticeLeftCenter, NoticeRightCenter:
				y = (600*scale - b.Dy()) / 2
			}
			if b.Min != image.Pt(x, y) || b.Empty() {
				t.Fatalf("scale=%d placement=%d got=%v want=%v", scale, p, b, image.Pt(x, y))
			}
		}
		n.Placement(NoticePlacement(255))
		if n.position(NoticeDefault) != NoticeRightCenter {
			t.Fatal("invalid default")
		}
		n.Placement(NoticeDefault)
		if n.position(NoticePlacement(255)) != NoticeTopRight {
			t.Fatal("default reset")
		}
	}
}

func TestNotifierIndependentPlacementQueues(t *testing.T) {
	n := Notifier()
	var firstBottom int
	for i := 0; i < MaxNotifications+1; i++ {
		n.Notify(Notice{Title: fmt.Sprintf("Top %d", i), Timeout: -1})
		id := n.Notify(Notice{Title: fmt.Sprintf("Bottom %d", i), Placement: NoticeBottomLeft, Timeout: -1})
		if i == 0 {
			firstBottom = id
		}
	}
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) }))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(900, 900); root.Layout(gtx) })
	h.Frame()
	if !shown(h, "Bottom 4") || !shown(h, "Top 4") || shown(h, "Bottom 5") || shown(h, "Top 5") {
		t.Fatal("per-placement capacity")
	}
	n.Dismiss(firstBottom)
	h.Frame()
	if !shown(h, "Bottom 5") || shown(h, "Top 5") {
		t.Fatal("queue release crossed placement")
	}
	n.Placement(NoticeTopLeft)
	h.Frame()
	if bounds(h, "Top 0").Min.X != 16 || bounds(h, "Bottom 1").Min.X != 16 || bounds(h, "Bottom 1").Min.Y < 400 {
		t.Fatal("override followed default")
	}
	click(t, h, "关闭 Bottom 1")
	h.Frame()
	if shown(h, "Bottom 1") {
		t.Fatal("positioned close")
	}
}

func TestNotifierMovingDoesNotRestartTimeout(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	n.Notify(Notice{Title: "Moving", Timeout: 5 * time.Second})
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(n.Render(cx)) })
	c.advance(h, 2*time.Second)
	n.Placement(NoticeBottomLeft)
	h.Frame()
	h.Frame()
	c.advance(h, 3*time.Second)
	h.Frame()
	if n.Len() != 0 {
		t.Fatal("moving restarted timeout")
	}
}
