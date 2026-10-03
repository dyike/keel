package window

import (
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMessageSurfaceChangesPreserveFocus(t *testing.T) {
	input := kit.Input("Draft")
	surface := kit.Bubble(input).Variant(kit.BubbleSecondary)
	msg := kit.Message("Me", nil).User().Bubble(surface).Header(kit.Label("Metadata"))
	w := openTest(t, kitPage(msg))
	w.click(element(t, w, "Draft").center())
	w.typeText("saved")
	surface.Variant(kit.BubbleGhost)
	msg.HeaderInset(true)
	w.snapshot()
	w.typeText("!")
	if input.Value() != "!saved" {
		t.Fatal("ghost lost focus", input.Value())
	}
	msg.ResetContentInsets()
	surface.Variant(kit.BubbleOutline)
	w.snapshot()
	w.typeText("?")
	if input.Value() != "?!saved" {
		t.Fatal("surface restore lost focus", input.Value())
	}
	msg.Content(input)
	w.snapshot()
	w.typeText("+")
	if input.Value() != "+?!saved" {
		t.Fatal("returning to automatic user bubble lost focus", input.Value())
	}
	msg.SetDisabled(true)
	w.snapshot()
	w.typeText("no")
	if input.Value() != "+?!saved" {
		t.Fatal("disabled surface accepted input")
	}
}
