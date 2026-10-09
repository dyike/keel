package el

import (
	"image"
	"image/color"
	"strings"
	"testing"
	"unicode/utf8"

	"gioui.org/io/key"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/inputcontent"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestInputObjectsWrapAndUseCustomGeometry(t *testing.T) {
	for _, scale := range []int{1, 2} {
		var doc InputDocument
		prefix := "这段说明后面的文档引用会整体换行： "
		content, _ := NewInputContent(prefix+"ref", InputTokenSpan{Range: InputRange{Start: len(prefix), End: len(prefix) + 3}, Token: InputToken{ID: "r", Text: "ref", Label: "引用"}})
		doc.SetContent(content)
		width := 160
		measured, painted := 0, 0
		renderer := func(g core.C, tok InputToken) core.D {
			if g.Enabled() {
				painted++
			} else {
				measured++
			}
			sz := image.Pt(min(g.Constraints.Max.X, g.Dp(96)), g.Dp(26))
			paint.FillShape(g.Ops, color.NRGBA{R: 220, A: 255}, clip.Rect(image.Rectangle{Max: sz}).Op())
			return core.D{Size: sz, Baseline: g.Dp(5)}
		}
		root := Root(ViewFunc(func(cx *Context) Element {
			return Div().Child(TextArea().ID("object").Document(&doc).TokenRenderer(renderer).W(Dp(float32(width))).H(Dp(150)))
		}))
		h := uitest.NewFunc(func(g core.C) {
			g.Metric = unit.Metric{PxPerDp: float32(scale), PxPerSp: float32(scale)}
			g.Constraints.Max = image.Pt(400*scale, 300*scale)
			root.Layout(g)
		})
		var st *elemState
		for _, s := range root.store.states {
			if s.id == "object" {
				st = s
			}
		}
		if st == nil || st.inputObjects.shaper == nil || len(st.inputTokenHits[0].rects) != 1 {
			t.Fatal("object not rendered")
		}
		r := st.inputTokenHits[0].rects[0]
		before := st.editor.Regions(0, 2, nil)
		if r.Dx() != 96*scale || r.Min.Y <= before[0].Bounds.Min.Y || measured == 0 || painted == 0 {
			t.Fatal("custom object geometry", r, before, measured, painted)
		}
		shaper := st.inputObjects.shaper
		h.Frame()
		if shaper != st.inputObjects.shaper {
			t.Fatal("unchanged frame rebuilt font")
		}
		width = 80
		h.Frame()
		if len(st.inputTokenHits[0].rects) != 1 || st.inputTokenHits[0].rects[0].Dx() > 80*scale {
			t.Fatal("narrow object split or overflow")
		}
		if doc.Content().Text() != content.Text() {
			t.Fatal("layout changed source")
		}
	}
}

func TestInputObjectsKeepPlatformLabelsAndMapCompositionAfterToken(t *testing.T) {
	var doc InputDocument
	prefix := "中\U000f0000 "
	token := InputToken{ID: "r", Text: "ref", Label: "引用\U000f0001"}
	content, _ := NewInputContent(prefix+"ref tail", InputTokenSpan{Range: InputRange{Start: len(prefix), End: len(prefix) + 3}, Token: token})
	doc.SetContent(content)
	focus := true
	root := Root(ViewFunc(func(cx *Context) Element {
		if focus {
			cx.Focus("object")
			focus = false
		}
		return TextArea().ID("object").Document(&doc).W(Dp(250)).H(Dp(100))
	}))
	h := uitest.New(root)
	h.Frame()
	var st *elemState
	for _, s := range root.store.states {
		if s.id == "object" {
			st = s
		}
	}
	if st == nil || !strings.Contains(st.editor.Text(), "\U000f0002") {
		t.Fatal("layout rune collided with source or label")
	}
	if h.Router.EditorState().Snippet.Text != content.Presentation().Text {
		t.Fatal("platform received object placeholders", h.Router.EditorState().Snippet.Text)
	}
	normal := content.Presentation()
	pos := utf8.RuneCountInString(normal.Text[:normal.DisplayOffset(len(prefix)+4, 0)])
	h.Router.Queue(key.SelectionEvent{Start: pos, End: pos}, key.CompositionEvent{Start: pos, End: pos + 2}, key.EditEvent{Range: key.Range{Start: pos, End: pos}, Text: "ni"})
	h.Frame()
	h.Frame()
	if doc.Content().Text() != prefix+"ref nitail" || len(doc.Content().Tokens()) != 1 {
		t.Fatal("composition used layout offsets", doc.Content().Text())
	}
	if h.Router.EditorState().Selection.CompositionBounds.Empty() {
		t.Fatal("missing composition geometry after token")
	}
	h.Router.Queue(key.EditEvent{Range: key.Range{Start: pos, End: pos + 2}, Text: "你"}, key.CompositionEvent{Start: -1, End: -1})
	h.Frame()
	h.Key("Z", key.ModShortcut)
	h.Frame()
	if !inputcontent.Equal(doc.Content(), content) {
		t.Fatal("composition undo lost source metadata")
	}
}
