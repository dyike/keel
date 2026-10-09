package markdown

import (
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/kit"
)

func TestSelectionAutoScrollInsideVirtualConversation(t *testing.T) {
	d := New("开头文字\n\n" + strings.Repeat("正文段落用于验证聊天中的选择滚动。\n\n", 60) + "末尾")
	sc := kit.MessageScroller([]string{"answer"}, 200, func(cx *el.Context, _ int) el.Element { return kit.Message("AI", d).Render(cx) })
	sc.SetFollow(false)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(el.Div().H(el.Dp(200)).Child(sc.Render(cx))) }))
	frame := 0
	base := time.Unix(100, 0)
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = base.Add(time.Duration(frame) * time.Second / 60)
		frame++
		root.Layout(gtx)
	})
	for range 8 {
		h.Frame()
	}
	x, y := textPoint(t, h, d, "开头文字", 1)
	origin := d.selection.origin
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: f32.Pt(x, y), Buttons: pointer.ButtonPrimary})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, 195), Buttons: pointer.ButtonPrimary})
	h.Frame()
	initial := len(d.selection.selectedText())
	for range 40 {
		h.Frame()
	}
	if len(d.selection.selectedText()) <= initial || d.selection.origin.Y >= origin.Y {
		t.Fatalf("virtual selection initial=%d now=%d origin=%v -> %v dragging=%v root=%v", initial, len(d.selection.selectedText()), origin, d.selection.origin, d.selection.dragging, d.selection.pointerRoot)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x, 195)})
	h.Frame()
	h.Frame()
	selected := d.selection.selectedText()
	origin = d.selection.origin
	for range 10 {
		h.Frame()
	}
	if d.selection.origin != origin || d.selection.selectedText() != selected {
		t.Fatal("selection continued after release")
	}
	assertSelectionCopy(t, h, d, selected)
}
