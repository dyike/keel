package kit

import (
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

// inputLongPress holds a finger on the named editor past the hold delay.
func inputLongPress(t *testing.T, h *uitest.Harness, name string) {
	t.Helper()
	var at *f32.Point
	var walk func(input.SemanticNode)
	walk = func(n input.SemanticNode) {
		if n.Desc.Class.String() == "Editor" && n.Desc.Label == name && at == nil {
			p := f32.Pt(float32(n.Desc.Bounds.Min.X+4), float32(n.Desc.Bounds.Min.Y+4))
			at = &p
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	for _, node := range h.Router.AppendSemantics(nil) {
		walk(node)
	}
	if at == nil {
		t.Fatal("no editor", name)
	}
	h.Router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 1, Position: *at})
	h.Frame()
	time.Sleep(600 * time.Millisecond)
	h.Frame()
	h.Router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Touch, PointerID: 1, Position: *at})
	h.Frame()
}

func TestInputTouchHoldOpensEditMenu(t *testing.T) {
	v := Input("note")
	v.SetValue("hello")
	h := render(func(c *el.Context) el.Element { return v.Render(c) })
	inputLongPress(t, h, "note")
	if v.editMenu == nil || !v.editMenu.open {
		t.Fatal("long press should open the edit menu")
	}
	if !shown(h, "全选") {
		t.Fatal("edit menu items")
	}
}

func TestTextAreaTouchHoldOpensEditMenu(t *testing.T) {
	v := TextArea("story")
	v.SetValue("line one\nline two")
	h := render(func(c *el.Context) el.Element { return v.Render(c) })
	inputLongPress(t, h, "story")
	if v.editMenu == nil || !v.editMenu.open {
		t.Fatal("long press should open the text area's edit menu")
	}
}
