package markdown

import (
	"fmt"
	"image"
	"strings"
	"testing"

	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestSelectionGroupAcrossEmbeddedDocuments(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		a, b := New("第一块正文"), New("第二块正文\n\n> 引用内容")
		var group SelectionGroup
		group.SetDocuments(a, b)
		ar, br := el.Embed(a), el.Embed(b)
		root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			embed := func(r *el.RootWidget) el.Element {
				return el.Widget(core.ViewportFunc(func(gtx core.C, visible image.Rectangle) core.D { return r.LayoutIn(gtx, visible, cx) }))
			}
			box := el.Div().P(20).Gap(16).Child(embed(ar), el.Text("运行了 4 个工具"), embed(br))
			return group.Decorate(cx, box)
		}))
		h := uitest.New(root)
		h.Frame()
		x0, y0 := textPoint(t, h, a, "第一块正文", 0)
		x1, y1 := textPoint(t, h, b, "引用内容", 4)
		if reverse {
			h.Drag(x1-1, y1, x0, y0)
		} else {
			h.Drag(x0, y0, x1-1, y1)
		}
		h.Frame()
		h.Key("C", key.ModShortcut)
		_, data, ok := h.Router.WriteClipboard()
		if !ok || !strings.Contains(string(data), "第一块正文\n\n第二块正文\n\n引用内容") {
			t.Fatalf("reverse=%v copy=%q", reverse, data)
		}
		if a.selection.selectedText() == "" || b.selection.selectedText() == "" {
			t.Fatal("both documents must highlight")
		}
	}
}

func TestSelectionGroupSelectAllIncludesUnmountedDocuments(t *testing.T) {
	a, b, c := New("第一块"), New("**未挂载的中间块**"), New("最后一块")
	var group SelectionGroup
	group.SetDocuments(a, b, c)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return group.Decorate(cx, el.Div().P(10).Child(a.Render(cx))) }))
	h := uitest.New(root)
	x, y := textPoint(t, h, a, "第一块", 1)
	h.Click(x, y)
	h.Frame()
	h.Key("A", key.ModShortcut)
	h.Frame()
	h.Key("C", key.ModShortcut)
	_, data, ok := h.Router.WriteClipboard()
	if !ok || string(data) != "第一块\n\n未挂载的中间块\n\n最后一块" {
		t.Fatalf("copy=%q", data)
	}
}

func TestSelectionGroupPreservesEndpointsWhenHistoryChanges(t *testing.T) {
	a, b := New("第一块"), New("第二块")
	var group SelectionGroup
	group.SetDocuments(a, b)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return group.Decorate(cx, el.Div().P(10).Child(a.Render(cx), b.Render(cx)))
	}))
	h := uitest.New(root)
	x0, y0 := textPoint(t, h, a, "第一块", 0)
	x1, y1 := textPoint(t, h, b, "第二块", 3)
	h.Drag(x0, y0, x1-1, y1)
	h.Frame()
	want := group.selection.selectedText()
	if want != "第一块\n\n第二块" {
		t.Fatalf("initial=%q", want)
	}
	before, after := New("更早的历史"), New("新消息")
	group.SetDocuments(before, a, b, after)
	h.Frame()
	h.Frame()
	if got := group.selection.selectedText(); got != want {
		t.Fatalf("history insertion moved selection: %q", got)
	}
	a.Append("新增文本")
	h.Frame()
	h.Frame()
	if got := group.selection.selectedText(); got != "第一块新增文本\n\n第二块" {
		t.Fatalf("streaming lost document-relative endpoint: %q", got)
	}
	b.SetSource("替换内容")
	h.Frame()
	if got := group.selection.selectedText(); got != "" {
		t.Fatalf("replacement retained stale selection: %q", got)
	}
}

func BenchmarkSelectionGroupVirtualHistory(b *testing.B) {
	for _, grouped := range []bool{false, true} {
		b.Run(fmt.Sprint(grouped), func(b *testing.B) {
			docs := make([]*Doc, 1000)
			for i := range docs {
				docs[i] = New("历史消息，只有末尾五条参与布局。")
			}
			var group SelectionGroup
			if grouped {
				group.SetDocuments(docs...)
			}
			root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
				box := el.Div()
				for _, d := range docs[len(docs)-5:] {
					box.Child(d.Render(cx))
				}
				if grouped {
					group.Decorate(cx, box)
				}
				return box
			}))
			h := uitest.New(root)
			h.Frame()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				h.Frame()
			}
		})
	}
}
