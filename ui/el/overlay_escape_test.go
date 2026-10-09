package el

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestOverlayEscapeHandlerBeforeDismiss(t *testing.T) {
	calls, dismissed := 0, 0
	open := true
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		if open {
			cx.Overlay("dialog", Modal(Div().Name("Dialog").Size(Dp(100))).OnEscape(func() bool { calls++; return calls == 1 }).OnDismiss(func() { dismissed++; open = false }))
		}
		return Div()
	})))
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if calls != 1 || dismissed != 0 || !open {
		t.Fatal(calls, dismissed, open)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if calls != 2 || dismissed != 1 || open {
		t.Fatal(calls, dismissed, open)
	}
}

func TestOverlayKeepOnEscapePrecedesHandler(t *testing.T) {
	calls := 0
	h := uitest.New(Root(ViewFunc(func(cx *Context) Element {
		cx.Overlay("dialog", Modal(Div().Size(Dp(100))).KeepOnEscape().OnEscape(func() bool { calls++; return false }).OnDismiss(func() { calls++ }))
		return Div()
	})))
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if calls != 0 {
		t.Fatal(calls)
	}
}
