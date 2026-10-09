package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/pointer"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func timedClick(h *uitest.Harness, x, y float32, at time.Duration) {
	p := f32.Pt(x, y)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: p, Time: at}, pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: p, Buttons: pointer.ButtonPrimary, Time: at}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p, Time: at + 10*time.Millisecond})
	h.Frame()
}

func TestDoubleClickWordBoundaries(t *testing.T) {
	for _, tc := range []struct {
		src, plain string
		index      int
		want       string
	}{
		{"render**ing** works", "rendering works", 5, "rendering"},
		{"hello_world other", "hello_world other", 8, "hello_world"},
		{"context.Context", "context.Context", 9, "Context"},
		{"can't stop", "can't stop", 3, "can't"},
		{"hello, world", "hello, world", 5, ","},
		{"中文，English", "中文，English", 1, "文"},
		{"café noir", "café noir", 3, "café"},
		{"👩‍💻 coder", "👩‍💻 coder", 0, "👩‍💻"},
		{"$x_i$ end", "$x_i$ end", 0, "$x_i$"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			d := New(tc.src)
			h := uitest.New(el.Root(docView{d}))
			x, y := textPoint(t, h, d, tc.plain, tc.index)
			timedClick(h, x+1, y, time.Second)
			timedClick(h, x+1, y, time.Second+100*time.Millisecond)
			assertSelectionCopy(t, h, d, tc.want)
		})
	}
}

func TestTripleClickParagraphAndCodeLine(t *testing.T) {
	paragraph := strings.Repeat("wrapped paragraph ", 8)
	d := New(paragraph + "\n\nnext paragraph")
	h := uitest.New(el.Root(docView{d}))
	x, y := textPoint(t, h, d, strings.TrimSpace(paragraph), 95)
	for i := range 3 {
		timedClick(h, x+1, y, time.Second+time.Duration(i)*100*time.Millisecond)
	}
	assertSelectionCopy(t, h, d, strings.TrimSpace(paragraph))
	d = New("```text\nfirst line\nsecond_line here\n```\n\nafter")
	h = uitest.New(el.Root(docView{d}))
	x, y = textPoint(t, h, d, "first line\nsecond_line here", 14)
	for i := range 3 {
		timedClick(h, x+1, y, time.Second+time.Duration(i)*100*time.Millisecond)
	}
	assertSelectionCopy(t, h, d, "second_line here")
}

func TestDoubleClickThenDragWholeWords(t *testing.T) {
	d := New("alpha beta gamma delta")
	h := uitest.New(el.Root(docView{d}))
	x, y := textPoint(t, h, d, "alpha beta gamma delta", 8)
	timedClick(h, x, y, time.Second)
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: f32.Pt(x, y), Buttons: pointer.ButtonPrimary, Time: time.Second + 100*time.Millisecond})
	h.Frame()
	x1, _ := textPoint(t, h, d, "alpha beta gamma delta", 14)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x1, y), Buttons: pointer.ButtonPrimary})
	h.Frame()
	if got := d.selection.selectedText(); got != "beta gamma" {
		t.Fatalf("forward word drag: %q", got)
	}
	x1, _ = textPoint(t, h, d, "alpha beta gamma delta", 2)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x1, y), Buttons: pointer.ButtonPrimary})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x1, y)})
	h.Frame()
	assertSelectionCopy(t, h, d, "alpha beta")
}

func TestDragSelectionAutoScroll(t *testing.T) {
	d := New("开头文字\n\n" + strings.Repeat("正文段落\n\n", 30) + "末尾")
	root := el.Root(scrollingDocView{d})
	frame := 0
	base := time.Now()
	h := uitest.NewFunc(func(gtx core.C) {
		gtx.Now = base.Add(time.Duration(frame) * time.Second / 60)
		frame++
		root.Layout(gtx)
	})
	x, y := textPoint(t, h, d, "开头文字", 1)
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: f32.Pt(x, y), Buttons: pointer.ButtonPrimary})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(x, 129), Buttons: pointer.ButtonPrimary})
	h.Frame()
	initial := len(d.selection.selectedText())
	for range 30 {
		h.Frame()
	}
	if got := len(d.selection.selectedText()); got <= initial {
		t.Fatalf("stationary edge drag did not expand selection: %d -> %d", initial, got)
	}
	if d.selection.origin.Y >= 0 {
		t.Fatal("document did not scroll")
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x, 129)})
	h.Frame()
	selected := d.selection.selectedText()
	h.Frame() // apply any scroll already scheduled by the last held frame
	origin := d.selection.origin
	for range 10 {
		h.Frame()
	}
	if d.selection.origin != origin || d.selection.selectedText() != selected {
		t.Fatal("selection kept scrolling after release")
	}
}
