package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestDialogIndependentDismissControls(t *testing.T) {
	calls, background := 0, 0
	d := Dialog("Options").Keyboard(false).Overlay(false).OverlayClosable(false).CloseButton(true).OnClose(func() { calls++ })
	h := page(Button("Background", func() { background++ }), d)
	d.SetValue(true)
	h.Frame()
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Click(5, 295)
	h.Frame()
	if !d.Value() || calls != 0 || background != 0 {
		t.Fatal("disabled dismissal or modal blocking")
	}
	click(t, h, locale.Current().Close)
	h.Frame()
	if d.Value() || calls != 1 {
		t.Fatal("explicit close")
	}
	d.CloseButton(false).Keyboard(true)
	d.SetValue(true)
	h.Frame()
	h.Frame()
	if shown(h, locale.Current().Close) {
		t.Fatal("hidden close button")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if d.Value() || calls != 2 {
		t.Fatal("Esc restored")
	}
	d.Keyboard(false).OverlayClosable(true).Persistent()
	d.SetValue(true)
	h.Frame()
	h.Frame()
	h.Click(5, 295)
	h.Frame()
	if d.Value() || calls != 3 {
		t.Fatal("explicit outside override")
	}
}

func TestDialogDisabledEscapeDoesNotCloseParent(t *testing.T) {
	child := Dialog("Child").Keyboard(false).CloseButton(true)
	parent := Dialog("Parent").Body(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Child(child.Render(cx)) }))
	parent.SetValue(true)
	h := page(parent)
	child.SetValue(true)
	h.Frame()
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !parent.Value() || !child.Value() {
		t.Fatal("Esc escaped top modal")
	}
	click(t, h, locale.Current().Close)
	h.Frame()
	h.Frame()
	if child.Value() || !parent.Value() {
		t.Fatal("child close affected parent")
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if parent.Value() {
		t.Fatal("parent Esc not restored")
	}
}
