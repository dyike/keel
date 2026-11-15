package el

import (
	"github.com/dyike/keel/gio/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestModalRetainedRootFocusRestoreAndTab(t *testing.T) {
	open, mounted := false, true
	parent := Root(ViewFunc(func(cx *Context) Element { return Input().ID("draft").Name("Draft") }))
	child := Root(ViewFunc(func(cx *Context) Element { return Input().ID("notes").Name("Notes") }))
	root := Root(ViewFunc(func(cx *Context) Element {
		if open {
			cx.Overlay("modal", Modal(Div().W(Dp(200)).H(Dp(150)).FocusTrap(true).Child(Div().ID("close").Name("Close").Focusable(true).H(Dp(32)), Widget(child).Grow())).OnDismiss(func() { open = false }))
		}
		box := Div()
		if mounted {
			box.Child(Widget(parent).H(Dp(80)))
		}
		return box
	}))
	h := uitest.New(root)
	b := nodeBounds(h, "Draft")
	h.Click(float32(b.Min.X+5), float32(b.Min.Y+5))
	h.Frame()
	if !(&Context{root: parent}).Focused("draft") {
		t.Fatal("parent focus missing")
	}
	open = true
	h.Frame()
	h.Frame()
	if !(&Context{root: root}).Focused("close") {
		t.Fatal("modal did not focus first control")
	}
	h.Key(key.NameTab, 0)
	h.Frame()
	if !(&Context{root: child}).Focused("notes") {
		t.Fatal("Tab skipped embedded input")
	}
	h.Key(key.NameTab, 0)
	h.Frame()
	if !(&Context{root: root}).Focused("close") {
		t.Fatal("Tab escaped modal")
	}
	h.Key(key.NameTab, key.ModShift)
	h.Frame()
	if !(&Context{root: child}).Focused("notes") {
		t.Fatal("reverse Tab skipped embedded input")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if open || !(&Context{root: parent}).Focused("draft") {
		t.Fatal("nested parent focus not restored")
	}
	open = true
	h.Frame()
	mounted = false
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if root.focusedTag() != nil {
		t.Fatal("removed page regained focus")
	}
}
