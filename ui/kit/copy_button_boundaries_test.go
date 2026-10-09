package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
	"time"
)

func TestCopyButtonRepeatedCopyRestartsFeedbackAndReadsLatest(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	value := "first"
	cb := CopyButton(func() string { return value })
	h := c.harness(func(cx *el.Context) el.Element { return cb.Render(cx) })
	click(t, h, "复制")
	h.Frame()
	_, _, _ = h.Router.WriteClipboard()
	c.advance(h, time.Second)
	value = "latest"
	click(t, h, "已复制")
	h.Frame()
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != "latest" {
		t.Fatalf("clipboard latest=%q %v", data, ok)
	}
	c.advance(h, time.Second)
	h.Frame()
	if !shown(h, "已复制") {
		t.Fatal("second click reused old feedback timer")
	}
	c.advance(h, 500*time.Millisecond)
	h.Frame()
	if !shown(h, "复制") {
		t.Fatal("feedback did not expire")
	}
}
func TestCopyButtonDisabledAndNilDoNotCopy(t *testing.T) {
	calls := 0
	cb := CopyButton(func() string { calls++; return "text" })
	disabled := false
	h := page(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(cb.Render(cx)) }))
	disabled = true
	h.Frame()
	click(t, h, "复制")
	if calls != 0 {
		t.Fatal("disabled ancestor copied")
	}
	disabled = false
	cb.SetDisabled(true)
	h.Frame()
	click(t, h, "复制")
	if calls != 0 {
		t.Fatal("disabled copied")
	}
	nilButton := CopyButton(nil)
	h = page(nilButton)
	click(t, h, "复制")
	if nilButton.copied {
		t.Fatal("nil getter reported copied")
	}
}

func TestCopyButtonContentAndCallback(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	value, reads := "first", 0
	var copied []string
	cb := CopyButton(func() string { reads++; return value })
	cb.OnCopied(func(s string) {
		if !cb.Copied() {
			t.Fatal("callback before feedback state")
		}
		copied = append(copied, s)
		value = "next"
	}).Content(el.ViewFunc(func(cx *el.Context) el.Element {
		if cb.Copied() {
			return el.Text("Done")
		}
		return el.Div().Row().Child(Icon(IconCopy).Render(cx), el.Text("Copy order"))
	}))
	disabled := false
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(cb.Render(cx)) })
	if reads != 0 || !shown(h, "Copy order") {
		t.Fatal("render read value or missed content")
	}
	h.Router.MoveFocus(key.FocusForward)
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if reads != 1 || len(copied) != 1 || copied[0] != "first" || !shown(h, "Done") {
		t.Fatalf("keyboard copy: reads=%d values=%v", reads, copied)
	}
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != "first" {
		t.Fatalf("clipboard=%q %v", data, ok)
	}
	c.advance(h, time.Second)
	h.Key(key.NameSpace, 0)
	h.Frame()
	if reads != 2 || len(copied) != 2 || copied[1] != "next" {
		t.Fatalf("repeat: %v reads=%d", copied, reads)
	}
	c.advance(h, time.Second)
	h.Frame()
	if !cb.Copied() {
		t.Fatal("repeat failed to restart feedback")
	}
	c.advance(h, 500*time.Millisecond)
	h.Frame()
	if cb.Copied() || !shown(h, "Copy order") {
		t.Fatal("custom feedback failed to expire")
	}
	disabled = true
	h.Frame()
	click(t, h, "复制")
	h.Key(key.NameReturn, 0)
	if reads != 2 || len(copied) != 2 {
		t.Fatal("disabled ancestor copied")
	}
	disabled = false
	cb.SetDisabled(true)
	h.Frame()
	click(t, h, "复制")
	if reads != 2 {
		t.Fatal("disabled control copied")
	}
	cb.SetDisabled(false)
	cb.Content(nil).OnCopied(nil)
	h.Frame()
	if shown(h, "Copy order") {
		t.Fatal("nil did not restore default")
	}
	click(t, h, "复制")
	if reads != 3 || len(copied) != 2 {
		t.Fatal("callback removal failed")
	}
}

func TestCopyButtonNilGetterSkipsCallback(t *testing.T) {
	cb := CopyButton(nil).OnCopied(func(string) { t.Fatal("nil getter called callback") })
	h := page(cb)
	click(t, h, "复制")
	if cb.Copied() {
		t.Fatal("nil getter showed feedback")
	}
}
