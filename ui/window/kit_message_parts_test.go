package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMessagePartStylesPreserveState(t *testing.T) {
	input := kit.Input("Draft")
	calls := 0
	msg := kit.Message("Author", input).Footer(kit.Button("Footer action", func() { calls++ }))
	w := openTest(t, kitPage(msg))
	w.click(element(t, w, "Draft").center())
	w.typeText("saved")
	msg.PartStyle(kit.MessagePartRoot, func(e *el.DivEl) { e.ID("replacement").Role("other").Name("wrong").Value("wrong").P(4).Disabled(false) })
	msg.PartStyle(kit.MessagePartStack, func(e *el.DivEl) { e.ID("other-stack").Gap(20) })
	msg.PartStyle(kit.MessagePartContent, func(e *el.DivEl) { e.ID("other-content").P(6) })
	w.snapshot()
	w.typeText("!")
	if input.Value() != "!saved" {
		t.Fatal("style changed identity", input.Value())
	}
	msg.PartStyle(kit.MessagePartRoot, nil).PartStyle(kit.MessagePartStack, nil).PartStyle(kit.MessagePartContent, nil)
	w.snapshot()
	w.typeText("?")
	if input.Value() != "?!saved" {
		t.Fatal("reset lost focus")
	}
	msg.PartStyle(kit.MessagePartContent, func(e *el.DivEl) { e.Disabled(true) })
	if !element(t, w, "Draft").Disabled || element(t, w, "Footer action").Disabled {
		t.Fatal("style leaked between parts")
	}
	w.click(element(t, w, "Footer action").center())
	if calls != 1 {
		t.Fatal("footer disabled by body")
	}
	msg.PartStyle(kit.MessagePartRoot, func(e *el.DivEl) { e.Disabled(false).Role("other").Name("wrong").Value("wrong") })
	msg.SetDisabled(true)
	w.click(element(t, w, "Footer action").center())
	if calls != 1 {
		t.Fatal("style bypassed disabled")
	}
	found := false
	for _, e := range w.snapshot() {
		if e.Role == "article" && e.Name == "Author" && e.Value == "" && e.Disabled {
			found = true
		}
	}
	if !found {
		t.Fatal("root semantics overridden")
	}
}
