package el

import (
	"gioui.org/f32"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestInputDocumentTokenActivation(t *testing.T) {
	var doc InputDocument
	c, err := NewInputContent("before reference after", InputTokenSpan{Range: InputRange{Start: 7, End: 16}, Token: InputToken{ID: "ref", Text: "reference", Label: "引用名称"}})
	if err != nil {
		t.Fatal(err)
	}
	doc.SetContent(c)
	calls := 0
	readonly, disabled := false, false
	root := Root(ViewFunc(func(cx *Context) Element {
		return Input().ID("tokens").Document(&doc).P(0).MinH(Auto).ReadOnly(readonly).Disabled(disabled).OnTokenActivate(func(token InputToken) {
			if token.ID != "ref" {
				t.Fatal(token)
			}
			calls++
		})
	}))
	h := uitest.New(root)
	var st *elemState
	for _, s := range root.store.states {
		if s.id == "tokens" {
			st = s
		}
	}
	if st == nil || len(st.inputTokenHits) != 1 || len(st.inputTokenHits[0].rects) == 0 {
		t.Fatal("token not laid out")
	}
	r := st.inputTokenHits[0].rects[0]
	for _, node := range h.Router.AppendSemantics(nil) {
		if node.Desc.Label == "引用名称" {
			r = node.Desc.Bounds
			break
		}
	}
	x, y := float32(r.Min.X+r.Max.X)/2, float32(r.Min.Y+r.Max.Y)/2
	h.Click(x, y)
	h.Frame()
	if calls != 1 {
		t.Fatal("click", calls, r)
	}
	if tok, ok := doc.SelectedToken(); !ok || tok.ID != "ref" {
		t.Fatal("click did not select reference", doc.Selection())
	}
	h.Drag(x, y, x+50, y)
	h.Frame()
	if calls != 1 {
		t.Fatal("drag activated token", calls)
	}
	p := f32.Pt(x, y)
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: p, Modifiers: key.ModShift}, pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: p, Modifiers: key.ModShift})
	h.Frame()
	if calls != 1 {
		t.Fatal("shift click activated")
	}
	readonly = true
	h.Frame()
	h.Click(x, y)
	h.Frame()
	if calls != 2 {
		t.Fatal("read-only activation", calls)
	}
	disabled = true
	h.Frame()
	h.Click(x, y)
	h.Frame()
	if calls != 2 {
		t.Fatal("disabled activation", calls)
	}
}
