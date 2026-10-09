package markdown

import (
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/input"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/io/pointer"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// textPoint uses the actual semantic bounds, rather than the selection
// controller's bounds, to exercise offsets from headings, lists and tables.
func textPoint(t *testing.T, h *uitest.Harness, d *Doc, label string, index int) (float32, float32) {
	t.Helper()
	var bounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == label {
			bounds = n.Desc.Bounds
		}
	})
	if bounds.Empty() {
		t.Fatalf("no visible text %q", label)
	}
	for _, part := range d.selection.parts {
		if part.r.plain != label {
			continue
		}
		for _, p := range part.r.rt.pieces {
			if index >= p.start && index <= p.start+p.runes {
				return float32(bounds.Min.X + p.rect.Min.X + p.xAt(index-p.start)),
					float32(bounds.Min.Y + (p.rect.Min.Y+p.rect.Max.Y)/2)
			}
		}
	}
	t.Fatalf("no rune %d in %q", index, label)
	return 0, 0
}

func assertSelectionCopy(t *testing.T, h *uitest.Harness, d *Doc, want string) {
	t.Helper()
	if got := d.selection.selectedText(); got != want {
		t.Fatalf("selected %q, want %q", got, want)
	}
	h.Key("C", key.ModShortcut)
	mime, data, ok := h.Router.WriteClipboard()
	if !ok || mime != "application/text" || string(data) != want {
		t.Fatalf("clipboard (%q, %q, %v), want %q", mime, data, ok, want)
	}
}

func TestSelectAcrossParagraphs(t *testing.T) {
	d := New("第一段**加粗**。\n\n第二段 `code`。\n\n末段结束。")
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "第一段加粗。", 2)
	x1, y1 := textPoint(t, h, d, "末段结束。", 2)
	want := "段加粗。\n\n第二段 code。\n\n末段"
	h.Drag(x0, y0, x1, y1)
	assertSelectionCopy(t, h, d, want)
	// Every block receives its own slice of the shared range.
	for i, want := range []string{"段加粗。", "第二段 code。", "末段"} {
		if got := d.selection.parts[i].r.rt.selectedText(); got != want {
			t.Fatalf("block %d selection %q, want %q", i, got, want)
		}
	}
	h.Drag(x1, y1, x0, y0)
	assertSelectionCopy(t, h, d, want)
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, d, "第一段加粗。\n\n第二段 code。\n\n末段结束。")
	h.Click(350, 280)
	if got := d.selection.selectedText(); got != "" {
		t.Fatalf("selection survived blur: %q", got)
	}
}

func TestSelectAcrossNestedBlocks(t *testing.T) {
	d := New("# 标题\n\n> 引用\n\n- 列表甲\n- 列表乙\n\n```go\nx := 1\n```\n\n结束")
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "标题", 1)
	x1, y1 := textPoint(t, h, d, "结束", 1)
	h.Drag(x0, y0, x1, y1)
	assertSelectionCopy(t, h, d, "题\n\n引用\n\n列表甲\n\n列表乙\n\nx := 1\n\n结")
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, d, "标题\n\n引用\n\n列表甲\n\n列表乙\n\nx := 1\n\n结束")
	// Copying a code block still copies only the code, without selecting it.
	var copyBounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == "复制" {
			copyBounds = n.Desc.Bounds
		}
	})
	if copyBounds.Empty() {
		t.Fatal("missing copy button")
	}
	h.Click(float32(copyBounds.Min.X+copyBounds.Max.X)/2, float32(copyBounds.Min.Y+copyBounds.Max.Y)/2)
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != "x := 1" || d.selection.selectedText() != "" {
		t.Fatalf("code copy %q, selection %q", data, d.selection.selectedText())
	}
}

func TestSelectAcrossTableCells(t *testing.T) {
	d := New("表前\n\n| 名称 | 数值 |\n|:--|--:|\n| 甲乙 | 一二 |\n| 丙丁 | 三四 |\n\n表后")
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "甲乙", 1)
	x1, y1 := textPoint(t, h, d, "三四", 1)
	h.Drag(x0, y0, x1, y1)
	assertSelectionCopy(t, h, d, "乙\t一二\n丙丁\t三")
	h.Drag(x1, y1, x0, y0)
	assertSelectionCopy(t, h, d, "乙\t一二\n丙丁\t三")
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, d, "表前\n\n名称\t数值\n甲乙\t一二\n丙丁\t三四\n\n表后")
}

func TestLinkDragDoesNotActivate(t *testing.T) {
	clicks := 0
	d := New("[官方文档](https://go.dev/doc)\n\n下一段").OnLink(func(string) { clicks++ })
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "官方文档", 0)
	x1, y1 := textPoint(t, h, d, "官方文档", 2)
	h.Drag(x0+1, y0, x1, y1)
	assertSelectionCopy(t, h, d, "官方")
	if clicks != 0 {
		t.Fatalf("drag activated link %d times", clicks)
	}
	x1, y1 = textPoint(t, h, d, "下一段", 1)
	h.Drag(x0+1, y0, x1, y1)
	assertSelectionCopy(t, h, d, "官方文档\n\n下")
	if clicks != 0 {
		t.Fatal("cross-paragraph drag activated link")
	}
	h.Click(x0+1, y0)
	if clicks != 1 {
		t.Fatalf("plain click activated link %d times", clicks)
	}
}

func TestSelectionSurvivesStreamingTailReplacement(t *testing.T) {
	d := New("首段\n\n**尾段")
	d.SetStreaming(true)
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "首段", 1)
	x1, y1 := textPoint(t, h, d, "尾段", 1)
	h.Drag(x0, y0, x1, y1)
	assertSelectionCopy(t, h, d, "段\n\n尾")
	first := d.selection.parts[0].r
	d.Append("继续**\n\n新段")
	h.Frame()
	assertSelectionCopy(t, h, d, "段\n\n尾")
	if first != d.selection.parts[0].r {
		t.Fatal("finished paragraph lost its cached rich text")
	}
	d.SetStreaming(false)
	h.Frame()
	assertSelectionCopy(t, h, d, "段\n\n尾")
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, d, "首段\n\n尾段继续\n\n新段")
	d.SetSource("替换后的内容")
	h.Frame()
	if got := d.selection.selectedText(); got != "" {
		t.Fatalf("old selection survived SetSource: %q", got)
	}
}

type scrollingDocView struct{ d *Doc }

func (v scrollingDocView) Render(cx *el.Context) el.Element {
	return el.Div().P(10).Child(el.Div().H(el.Dp(120)).ScrollY().Child(v.d.Render(cx)))
}

func TestSelectWhileScrolling(t *testing.T) {
	d := New("首段\n\n" + strings.Repeat("中间段落\n\n", 12) + "末段")
	h := uitest.New(el.Root(scrollingDocView{d}))
	x0, y0 := textPoint(t, h, d, "首段", 1)
	p := f32.Pt(x0, y0)
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p})
	h.Frame()
	h.Scroll(100, 60, 1000)
	x1, y1 := textPoint(t, h, d, "末段", 1)
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x1, y1)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(x1, y1)})
	h.Frame()
	assertSelectionCopy(t, h, d, "段\n\n"+strings.Repeat("中间段落\n\n", 12)+"末")
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, d, "首段\n\n"+strings.Repeat("中间段落\n\n", 12)+"末段")
}

func TestDragOutsideDocument(t *testing.T) {
	d := New("开头\n\n| 左 | 右 |\n|--|--|\n| 甲 | 乙 |")
	h := uitest.New(el.Root(docView{d}))
	x0, y0 := textPoint(t, h, d, "开头", 0)
	h.Drag(x0+1, y0, x0, 280)
	assertSelectionCopy(t, h, d, "开头\n\n左\t右\n甲\t乙")
	x1, y1 := textPoint(t, h, d, "乙", 1)
	h.Drag(x1, y1, 0, 0)
	assertSelectionCopy(t, h, d, "开头\n\n左\t右\n甲\t乙")
}

type twoDocView struct{ a, b *Doc }

func (v twoDocView) Render(cx *el.Context) el.Element {
	return el.Div().P(10).Gap(20).Child(v.a.Render(cx), v.b.Render(cx))
}

func TestSelectionIsScopedToOneDocument(t *testing.T) {
	a, b := New("回答甲"), New("回答乙\n\n第二段")
	h := uitest.New(el.Root(twoDocView{a, b}))
	x, y := textPoint(t, h, a, "回答甲", 1)
	h.Click(x, y)
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, a, "回答甲")
	x, y = textPoint(t, h, b, "回答乙", 1)
	h.Click(x, y)
	h.Key("A", key.ModShortcut)
	assertSelectionCopy(t, h, b, "回答乙\n\n第二段")
	if a.selection.selectedText() != "" {
		t.Fatal("the first document retained selection after focus changed")
	}
}

func TestWheelDuringAndAfterSelection(t *testing.T) {
	d := New("开头可选文字\n\n" + strings.Repeat("正文段落\n\n", 20) + "末尾")
	h := uitest.New(el.Root(scrollingDocView{d}))
	x, y := textPoint(t, h, d, "开头可选文字", 1)
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x, y)})
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(x+30, y)})
	h.Frame()
	if !d.selection.moved {
		t.Fatal("test did not grab a selection drag")
	}
	h.Scroll(100, 60, 40)
	if d.selection.pointerRoot != image.Pt(100, 60) {
		t.Fatalf("pointer moved with scrolling content: %v", d.selection.pointerRoot)
	}
	if d.selection.parts[0].r.plain != "开头可选文字" {
		t.Fatal("lost selection document")
	}
	var bounds image.Rectangle
	walk(h.Router.AppendSemantics(nil)[0], func(n input.SemanticNode) {
		if n.Desc.Label == "开头可选文字" {
			bounds = n.Desc.Bounds
		}
	})
	if !bounds.Empty() && bounds.Min.Y >= int(y) {
		t.Fatal("wheel did not scroll during selection")
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(100, 60)})
	h.Frame()
	selected := d.selection.selectedText()
	h.Scroll(100, 60, 1000)
	textPoint(t, h, d, "末尾", 1)
	if d.selection.selectedText() != selected {
		t.Fatal("scroll after release changed selection")
	}
}
