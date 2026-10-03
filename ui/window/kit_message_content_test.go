package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMessageContentReorderPreservesFocus(t *testing.T) {
	input := kit.Input("Draft")
	first := kit.Bubble(input)
	calls := 0
	action := kit.Button("Download", func() { calls++ })
	second := kit.Bubble(kit.Label("Second"))
	mixed := kit.MessageContent(first, action, second)
	msg := kit.Message("Me", mixed).User()
	w := openTest(t, kitPage(msg))
	w.click(element(t, w, "Draft").center())
	w.typeText("saved")
	mixed.SetItems(second, action, first)
	w.snapshot()
	w.typeText("!")
	if input.Value() != "!saved" {
		t.Fatal("reorder lost focus", input.Value())
	}
	mixed.Style(func(e *el.DivEl) { e.ID("ignored").Gap(12) })
	msg.Alignment(el.Start)
	w.snapshot()
	w.typeText("?")
	if input.Value() != "?!saved" {
		t.Fatal("style/alignment lost focus")
	}
	w.click(element(t, w, "Download").center())
	if calls != 1 {
		t.Fatal("mixed action")
	}
	mixed.SetDisabled(true)
	w.click(element(t, w, "Download").center())
	if calls != 1 || !element(t, w, "Draft").Disabled {
		t.Fatal("disabled inheritance")
	}
	mixed.SetDisabled(false)
	mixed.Style(nil)
	mixed.SetItems(first)
	for _, e := range w.snapshot() {
		if e.Name == "Download" || e.Name == "Second" {
			t.Fatal("removed element")
		}
	}
	if input.Value() != "?!saved" {
		t.Fatal("retained input")
	}
}
