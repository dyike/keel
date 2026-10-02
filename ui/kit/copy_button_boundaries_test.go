package kit

import (
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
