package markdown

import (
	"fmt"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/io/pointer"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestEmbeddedSelectionFollowsOuterScroll(t *testing.T) {
	for _, padding := range []int{0, 8} {
		for _, scale := range []int{1, 2} {
			t.Run(fmt.Sprintf("padding=%d/scale=%d", padding, scale), func(t *testing.T) {
				doc := New("开头可选文字\n\n" + strings.Repeat("正文段落\n\n", 30) + "末尾")
				nested := el.Embed(docView{doc})
				var offset float32
				root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
					return el.Div().P(20).Child(el.Div().ID("outer-scroll").H(el.Dp(120)).ScrollY().Child(
						el.Widget(core.ViewportFunc(func(gtx core.C, visible image.Rectangle) core.D { return nested.LayoutIn(gtx, visible, cx) })).Px(float32(padding)).Py(float32(padding)),
					).Decorate(func(gtx core.C, draw func()) { draw(); offset, _, _ = cx.ScrollState("outer-scroll") }))
				}))
				now := time.Now()
				h := uitest.NewFunc(func(gtx core.C) {
					now = now.Add(time.Second / 60)
					gtx.Now = now
					gtx.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
					gtx.Constraints.Max = image.Pt(400*scale, 300*scale)
					root.Layout(gtx)
				})
				h.Frame()
				x, y := textPoint(t, h, doc, "开头可选文字", 1)
				h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
				h.Frame()
				h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, float32(139*scale))})
				for range 80 {
					h.Frame()
				}
				if offset < 300 {
					t.Fatalf("embedded root did not scroll its enclosing viewport: %v", offset)
				}
				if doc.selection.pointerRoot != image.Pt(int(x), 139*scale) {
					t.Fatalf("measurement or scrolling moved stationary pointer: %v", doc.selection.pointerRoot)
				}
				h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x, float32(139*scale))})
				h.Frame()
				selected := doc.selection.selectedText()
				if strings.Count(selected, "正文段落") < 5 {
					t.Fatalf("selection did not grow with viewport: %q", selected)
				}
				assertSelectionCopy(t, h, doc, selected)
				// Scrolling after releasing the pointer keeps the selected range intact.
				h.Scroll(float32(100*scale), float32(80*scale), float32(-1000*scale))
				h.Frame()
				assertSelectionCopy(t, h, doc, selected)
			})
		}
	}
}

func TestCachedDocumentReflowsWhenFontScaleChanges(t *testing.T) {
	doc := New(strings.Repeat("中文正文 mixed words，验证字号变化后的换行和选区。", 4))
	root := el.Embed(docView{doc})
	scale := float32(1)
	var size image.Point
	h := uitest.NewFunc(func(gtx core.C) { gtx.Metric.PxPerSp = scale; size = root.Layout(gtx).Size })
	h.Frame()
	before := size
	scale = 1.5
	h.Frame()
	if size.Y <= before.Y {
		t.Fatalf("font scaling reused stale cached geometry: before=%v after=%v", before, size)
	}
}
