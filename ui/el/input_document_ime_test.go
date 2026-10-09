package el

import (
	"github.com/dyike/keel/third_party/gio/f32"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/third_party/gio/unit"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"image"
	"testing"
)

func TestInputDocumentCompositionGeometry(t *testing.T) {
	for _, scale := range []float32{1, 2} {
		var doc InputDocument
		c, _ := NewInputContent("ref tail", InputTokenSpan{Range: InputRange{Start: 0, End: 3}, Token: InputToken{ID: "ref", Text: "ref", Label: "引用"}})
		doc.SetContent(c)
		focus := true
		root := Root(ViewFunc(func(cx *Context) Element {
			if focus {
				cx.Focus("input")
				focus = false
			}
			return Div().P(18).Child(TextArea().ID("input").Document(&doc).W(Dp(180)).H(Dp(80)))
		}))
		h := uitest.NewFunc(func(g core.C) {
			g.Metric = unit.Metric{PxPerDp: scale, PxPerSp: scale}
			g.Constraints.Max = image.Pt(int(400*scale), int(300*scale))
			root.Layout(g)
		})
		h.Frame()
		h.Router.Queue(key.SelectionEvent{Start: 0, End: 2}, key.CompositionEvent{Start: 0, End: 2}, key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "ni"})
		h.Frame()
		h.Frame()
		state := h.Router.EditorState()
		if state.Selection.CompositionBounds.Empty() {
			t.Fatalf("scale %v: no composing bounds reported", scale)
		}
		if state.Snippet.Text != "ni tail" {
			t.Fatalf("snippet %q", state.Snippet.Text)
		}
		if state.Selection.Caret.Ascent <= 0 {
			t.Fatal("no caret ascent")
		}
		// Candidate geometry is local to the editor's translated input area.
		// It must land inside the field, not at the window origin or twice offset.
		b := state.Selection.CompositionBounds
		minPoint := state.Selection.Transform.Transform(f32.Pt(float32(b.Min.X), float32(b.Min.Y)))
		maxPoint := state.Selection.Transform.Transform(f32.Pt(float32(b.Max.X), float32(b.Max.Y)))
		if minPoint.X < 18*scale || minPoint.Y < 18*scale || maxPoint.X > 198*scale || maxPoint.Y > 98*scale {
			t.Fatalf("composition transform leaves field: %v..%v", minPoint, maxPoint)
		}

		h.Router.Queue(key.EditEvent{Range: key.Range{Start: 0, End: 2}, Text: "你"}, key.CompositionEvent{Start: 0, End: 1})
		h.Frame()
		h.Frame()
		if h.Router.EditorState().Selection.CompositionBounds.Empty() {
			t.Fatal("replacement lost composition bounds")
		}
		h.Router.Queue(key.CompositionEvent{Start: -1, End: -1})
		h.Frame()
		h.Frame()
		if !h.Router.EditorState().Selection.CompositionBounds.Empty() {
			t.Fatal("committed text retains composing bounds")
		}
		h.Key("Z", key.ModShortcut)
		h.Frame()
		if doc.Content().Text() != "ref tail" || len(doc.Content().Tokens()) != 1 {
			t.Fatal("geometry handling broke composition history")
		}
	}
}

func TestInputDocumentCompositionLifecycle(t *testing.T) {
	for _, mode := range []string{"readonly", "disabled", "reset", "blur"} {
		t.Run(mode, func(t *testing.T) {
			var doc InputDocument
			_ = doc.SetText("draft")
			readOnly, disabled := false, false
			focus := "input"
			root := Root(ViewFunc(func(cx *Context) Element {
				if focus != "" {
					cx.Focus(focus)
					focus = ""
				}
				return Div().Child(Input().ID("input").Document(&doc).ReadOnly(readOnly).Disabled(disabled), Input().ID("other"))
			}))
			h := uitest.New(root)
			h.Frame()
			h.Router.Queue(key.CompositionEvent{Start: 0, End: 2}, key.EditEvent{Range: key.Range{Start: 0, End: 5}, Text: "ni"})
			h.Frame()
			h.Frame()
			if h.Router.EditorState().Selection.CompositionBounds.Empty() || !doc.session.Composing() {
				t.Fatal("composition was not active")
			}
			switch mode {
			case "readonly":
				readOnly = true
			case "disabled":
				disabled = true
			case "reset":
				_ = doc.SetText("new")
			case "blur":
				focus = "other"
			}
			h.Frame()
			h.Frame()
			if doc.session.Composing() {
				t.Fatal("composition survived lifecycle change")
			}
			if !h.Router.EditorState().Selection.CompositionBounds.Empty() {
				t.Fatal("stale composing bounds")
			}
			readOnly, disabled = false, false
			focus = "input"
			h.Frame()
			h.Frame()
			if !h.Router.EditorState().Selection.CompositionBounds.Empty() {
				t.Fatal("refocus restored stale composing bounds")
			}
		})
	}
}
