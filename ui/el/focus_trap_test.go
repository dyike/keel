package el

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestContainerFocusTrapScopesAndOrder(t *testing.T) {
	request := "a"
	enabled, inner, disabled := true, true, false
	root := Root(ViewFunc(func(cx *Context) Element {
		if request != "" {
			cx.Focus(request)
			request = ""
		}
		button := func(id string) *DivEl { return Div().ID(id).Name(id).Focusable(true).H(Dp(24)).Child(Text(id)) }
		return Div().Child(
			button("outside").TabIndex(0),
			Div().FocusTrap(enabled).Child(button("a").TabIndex(2), Input().ID("input").Name("input").TabIndex(1), button("skip").TabStop(false), button("hidden").Hidden(true), button("disabled").Disabled(disabled),
				Div().FocusTrap(inner).Child(button("inner1"), button("inner2"))),
			Div().FocusTrap(true).Child(button("b1"), button("b2")), button("end"))
	}))
	h := uitest.New(root)
	focused := func(want string) {
		t.Helper()
		if !(&Context{root: root}).Focused(want) {
			t.Fatalf("want focus %s, got %v", want, root.focusedTag())
		}
	}
	focused("a")
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("disabled")
	request = "input"
	h.Frame()
	focused("input")
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("a")
	h.Key(key.NameTab, key.ModShift)
	h.Frame()
	focused("input")
	request = "inner1"
	h.Frame()
	h.Key(key.NameTab, key.ModShift)
	h.Frame()
	focused("inner2")
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("inner1")
	inner = false
	h.Frame()
	h.Key(key.NameTab, key.ModShift)
	h.Frame()
	focused("disabled")
	disabled = true
	request = "a"
	h.Frame()
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("inner1")
	request = "b2"
	h.Frame()
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("b1")
	// Clicking outside a scope is allowed and doesn't get pulled back next frame.
	r := nodeBounds(h, "outside")
	h.Click(float32(r.Min.X+2), float32(r.Min.Y+2))
	h.Frame()
	focused("outside")
	enabled = false
	request = "a"
	h.Frame()
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("outside")
}

func TestContainerFocusTrapOverlayAndRemoval(t *testing.T) {
	request := "a"
	open, show := false, true
	root := Root(ViewFunc(func(cx *Context) Element {
		if request != "" {
			cx.Focus(request)
			request = ""
		}
		if open {
			cx.Overlay("modal", Modal(Div().W(Dp(100)).Child(Input().ID("modal1"), Input().ID("modal2"))).OnDismiss(func() { open = false }))
		}
		parent := Div().Child(Div().ID("outside").Focusable(true).TabIndex(0).Child(Text("outside")))
		if show {
			parent.Child(Div().FocusTrap(true).Child(Input().ID("a"), Input().ID("b")))
		}
		return parent
	}))
	h := uitest.New(root)
	focused := func(want string) {
		t.Helper()
		if !(&Context{root: root}).Focused(want) {
			t.Fatalf("want %s", want)
		}
	}
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("b")
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("a")
	open = true
	h.Frame()
	h.Frame()
	focused("modal1")
	// Background container traps must not constrain foreground modal focus.
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	focused("modal2")
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	focused("modal1")
	h.Key(key.NameEscape, 0)
	h.Frame()
	focused("a")
	show = false
	h.Frame()
	h.Key(key.NameTab, 0)
	h.Frame()
	focused("outside")
}
