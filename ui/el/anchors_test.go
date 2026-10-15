package el

import (
	"fmt"
	"image"
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestAnchorsReuseAndRelease(t *testing.T) {
	var e engine
	e.beginAnchors()
	old := e.anchors
	old["old"] = image.Rect(1, 2, 3, 4)
	e.beginAnchors()
	if len(old) != 0 {
		t.Fatal("prior string references retained")
	}
	e.anchors["new"] = image.Rectangle{}
	if _, ok := old["new"]; !ok {
		t.Fatal("small map not reused")
	}
	if n := testing.AllocsPerRun(100, func() { e.beginAnchors(); e.anchors["new"] = image.Rectangle{} }); n != 0 {
		t.Fatalf("warm map allocated %g times", n)
	}
	for i := range maxRetainedAnchors + 1 {
		e.anchors[fmt.Sprint(i)] = image.Rectangle{}
	}
	large := e.anchors
	e.beginAnchors()
	e.anchors["fresh"] = image.Rectangle{}
	if _, ok := large["fresh"]; ok {
		t.Fatal("oversized map retained")
	}
}

func TestAnchorsKeepPreviousFrameUntilPaint(t *testing.T) {
	show := true
	var previous [2]float32
	r := Root(ViewFunc(func(cx *Context) Element {
		previous[0], previous[1] = cx.LastSize("target")
		if show {
			return Div().Child(Div().ID("target").W(Dp(23)).H(Dp(17)), Div().ID("target").W(Dp(70)).H(Dp(60)))
		}
		return Div()
	}))
	h := uitest.New(r)
	h.Frame()
	if previous != [2]float32{23, 17} {
		t.Fatalf("previous frame/first duplicate lost: %v", previous)
	}
	show = false
	h.Frame()
	if previous != [2]float32{23, 17} {
		t.Fatal("previous anchors cleared before Render")
	}
	h.Frame()
	if previous != [2]float32{} || len(r.e.anchors) != 0 {
		t.Fatal("stale anchor survived new frame")
	}
}

func BenchmarkAnchorsFrame(b *testing.B) {
	ids := make([]string, 512)
	for i := range ids {
		ids[i] = fmt.Sprint(i)
	}
	for _, reuse := range []bool{false, true} {
		name := "fresh"
		if reuse {
			name = "reuse"
		}
		b.Run(name, func(b *testing.B) {
			var e engine
			frame := func() {
				if reuse {
					e.beginAnchors()
				} else {
					e.anchors = make(map[string]image.Rectangle)
				}
				for _, id := range ids {
					e.anchors[id] = image.Rectangle{}
				}
			}
			for range 3 {
				frame()
			}
			b.ReportAllocs()
			for b.Loop() {
				frame()
			}
		})
	}
}
