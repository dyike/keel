package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"strings"
	"testing"
	"time"
)

func TestSheetFooterAndDismissOptions(t *testing.T) {
	for _, side := range []el.Side{el.Left, el.Right, el.Top, el.Bottom} {
		calls, closes := 0, 0
		s := Sheet(side, "Panel").Size(220).Keyboard(false).Overlay(false).OverlayClosable(false).CloseButton(false).OnClose(func() { closes++ })
		views := []el.View{Button("Apply", func() { calls++ })}
		s.Footer(views...)
		views[0] = Button("Wrong", nil)
		s.Body(text(strings.Repeat("Long body\n", 100)))
		c := &clock{now: time.Unix(100, 0)}
		s.SetValue(true)
		h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(s.Render(cx)) })
		c.advance(h, SheetSlide)
		h.Frame()
		if shown(h, locale.Current().Close) || shown(h, "Wrong") {
			t.Fatal("configuration ignored")
		}
		b := bounds(h, "Apply")
		if b.Min.Y < 0 || b.Max.Y > 300 || b.Dy() == 0 {
			t.Fatal("footer not visible", side, b)
		}
		click(t, h, "Apply")
		if calls != 1 {
			t.Fatal("footer action")
		}
		h.Key(key.NameEscape, 0)
		h.Frame()
		if !s.Value() || closes != 0 {
			t.Fatal("disabled Esc")
		}
		s.CloseButton(true)
		h.Frame()
		click(t, h, locale.Current().Close)
		h.Frame()
		if s.Value() || closes != 1 {
			t.Fatal("close button")
		}
	}
}
