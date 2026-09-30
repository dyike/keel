package markdown

import (
	"image"
	"strings"
	"testing"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func codeButton(t *testing.T, h *uitest.Harness, label string) image.Rectangle {
	t.Helper()
	var bounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == label && n.Desc.Class.String() == "Button" {
			bounds = n.Desc.Bounds
		}
	})
	if bounds.Empty() {
		t.Fatalf("missing code action %q", label)
	}
	return bounds
}

func clickCodeButton(t *testing.T, h *uitest.Harness, label string) {
	t.Helper()
	b := codeButton(t, h, label)
	h.Click(float32(b.Min.X+b.Max.X)/2, float32(b.Min.Y+b.Max.Y)/2)
}

func codeLines(r *richBlock) int {
	lines, y := 0, -1
	for _, p := range r.rt.pieces {
		if p.rect.Min.Y != y {
			y = p.rect.Min.Y
			lines++
		}
	}
	return lines
}

func codeTextBounds(t *testing.T, h *uitest.Harness, text string) image.Rectangle {
	t.Helper()
	var bounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == text {
			bounds = n.Desc.Bounds
		}
	})
	if bounds.Empty() {
		t.Fatal("missing visible code text")
	}
	return bounds
}

func TestCodeActionsAndWrap(t *testing.T) {
	code := strings.Repeat("long word ", 20) + "END\n\tshort"
	d := New("```text\n" + code + "\n```")
	h := uitest.New(el.Root(docView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	if got := codeLines(cv.rich); got != 2 {
		t.Fatalf("code wrapped by default: %d visual lines", got)
	}
	if cv.body.full.X <= cv.body.viewport.X {
		t.Fatal("long code must overflow its text viewport")
	}
	copyBefore := codeButton(t, h, "复制")
	clickCodeButton(t, h, "自动换行")
	if !cv.wrap || codeLines(cv.rich) <= 2 || cv.body.scrollX != 0 {
		t.Fatal("wrap action did not reflow the code")
	}
	if got := codeButton(t, h, "复制"); got != copyBefore {
		t.Fatalf("code header moved on reflow: %v -> %v", copyBefore, got)
	}
	clickCodeButton(t, h, "复制")
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != code {
		t.Fatalf("copy modified the original code: %q", data)
	}
	codeButton(t, h, "已复制")
	clickCodeButton(t, h, "取消自动换行")
	if cv.wrap || codeLines(cv.rich) != 2 {
		t.Fatal("disabling wrap did not restore source lines")
	}
}

func TestCodeHorizontalScrollAndSelection(t *testing.T) {
	code := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 8)
	d := New("```text\n" + code + "\n```\n\n正文")
	h := uitest.New(el.Root(docView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	bounds := codeTextBounds(t, h, code)
	y := float32(bounds.Min.Y + cv.rich.rt.size.Y/2)
	copyBounds := codeButton(t, h, "复制")
	h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse,
		Position: f32.Pt(float32(bounds.Min.X+20), y), Scroll: f32.Pt(200, 0)})
	h.Frame()
	if cv.body.scrollX != 200 {
		t.Fatalf("horizontal scroll = %d, want 200", cv.body.scrollX)
	}
	if codeButton(t, h, "复制") != copyBounds {
		t.Fatal("code actions scrolled along with the code")
	}
	p := cv.rich.rt.pieces[0]
	x0 := float32(bounds.Min.X + p.xAt(30) - cv.body.scrollX)
	x1 := float32(bounds.Min.X + p.xAt(35) - cv.body.scrollX)
	h.Drag(x0, y, x1, y)
	assertSelectionCopy(t, h, d, code[30:35])
	// Selecting out of a scrolled code block still uses document rune indexes.
	x1, y1 := textPoint(t, h, d, "正文", 1)
	h.Drag(x0, y, x1, y1)
	assertSelectionCopy(t, h, d, code[30:]+"\n\n正")
}

func TestCodeScrollbarDrag(t *testing.T) {
	code := strings.Repeat("abcdefghij", 40)
	d := New("```text\n" + code + "\n```")
	h := uitest.New(el.Root(docView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	bounds := codeTextBounds(t, h, code)
	y := float32(bounds.Min.Y + cv.body.viewport.Y - 5)
	x := float32(bounds.Min.X + 5)
	h.Drag(x, y, float32(bounds.Min.X+cv.body.viewport.X-1), y)
	if got, want := cv.body.scrollX, cv.body.full.X-cv.body.viewport.X; got != want {
		t.Fatalf("scrollbar drag = %d, want %d", got, want)
	}
	if d.selection.selectedText() != "" {
		t.Fatal("dragging the scrollbar selected code")
	}
}

func TestIndependentCodeBlocksAndNestedWrapCache(t *testing.T) {
	line := strings.Repeat("word ", 25)
	d := New("> ```text\n> " + line + "\n> ```\n>\n> ```text\n> " + line + "\n> ```")
	h := uitest.New(el.Root(docView{d}))
	children := d.chunks[0].blocks[0].children
	a, b := children[0].view.(*codeView), children[1].view.(*codeView)
	clickCodeButton(t, h, "自动换行") // last matching action: second code block
	if a.wrap || !b.wrap || codeLines(a.rich) != 1 || codeLines(b.rich) <= 1 {
		t.Fatal("wrap setting leaked between code blocks")
	}
	clickCodeButton(t, h, "自动换行") // first code block
	if !a.wrap || !b.wrap {
		t.Fatal("the first nested code block could not be wrapped")
	}
	clickCodeButton(t, h, "取消自动换行") // second code block
	if !a.wrap || b.wrap || codeLines(b.rich) != 1 {
		t.Fatal("nested cached code did not update independently")
	}
}

func TestStreamingCodeRetainsControls(t *testing.T) {
	line := strings.Repeat("word ", 25)
	d := New("```text\n" + line)
	d.SetStreaming(true)
	h := uitest.New(el.Root(docView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	clickCodeButton(t, h, "自动换行")
	d.Append("\nnext line\n```")
	d.SetStreaming(false)
	h.Frame()
	if d.chunks[0].blocks[0].view != cv || !cv.wrap || codeLines(cv.rich) <= 2 {
		t.Fatal("streaming reset the code controls")
	}
	clickCodeButton(t, h, "复制")
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != line+"\nnext line" {
		t.Fatalf("copy kept stale streaming code: %q", data)
	}
}

func TestCodeScrollInsideVerticalDocument(t *testing.T) {
	code := strings.Repeat("abcdefghij", 40)
	d := New("```text\n" + code + "\n```\n\n" + strings.Repeat("正文\n\n", 20))
	h := uitest.New(el.Root(scrollingDocView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	bounds := codeTextBounds(t, h, code)
	p := f32.Pt(float32(bounds.Min.X+40), float32(bounds.Min.Y+10))
	for _, want := range []int{100, 200} {
		h.Router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: p, Scroll: f32.Pt(100, 0)})
		h.Frame()
		if cv.body.scrollX != want {
			t.Fatalf("nested horizontal scroll = %d, want %d", cv.body.scrollX, want)
		}
	}
	h.Scroll(p.X, p.Y, 40)
	if cv.body.scrollX != 200 {
		t.Fatal("vertical wheel changed horizontal position")
	}
	if got := codeTextBounds(t, h, code).Min.Y; got >= bounds.Min.Y {
		t.Fatal("vertical wheel over code did not scroll document")
	}
}

func TestCodeScrollbarWheelAndTrackInsideChat(t *testing.T) {
	code := strings.Repeat("abcdefghij", 40)
	d := New("```text\n" + code + "\n```\n\n" + strings.Repeat("正文\n\n", 20))
	h := uitest.New(el.Root(scrollingDocView{d}))
	cv := d.chunks[0].blocks[0].view.(*codeView)
	bounds := codeTextBounds(t, h, code)
	y := float32(bounds.Min.Y + cv.body.viewport.Y - 5)
	x := float32(bounds.Min.X + cv.body.viewport.X/2)
	h.Scroll(x, y, 100)
	if cv.body.scrollX != 100 {
		t.Fatalf("wheel on scrollbar = %d, want 100", cv.body.scrollX)
	}
	if codeTextBounds(t, h, code).Min.Y != bounds.Min.Y {
		t.Fatal("wheel on scrollbar moved the chat")
	}
	h.Click(float32(bounds.Min.X+cv.body.viewport.X-2), y)
	if cv.body.scrollX != cv.body.full.X-cv.body.viewport.X {
		t.Fatal("track click did not reach right end")
	}
	if d.selection.selectedText() != "" {
		t.Fatal("scrollbar selected text")
	}
}
