package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMessageAlignmentPreservesFocus(t *testing.T) {
	input := kit.Input("Draft")
	msg := kit.Message("Me", input).User().Avatar(kit.Avatar("Me"))
	w := openTest(t, kitPage(msg))
	w.click(element(t, w, "Draft").center())
	w.typeText("saved")
	msg.Alignment(el.Start)
	w.snapshot()
	w.typeText("!")
	if input.Value() != "!saved" {
		t.Fatal("alignment lost focus", input.Value())
	}
	msg.ResetAlignment()
	w.snapshot()
	w.typeText("?")
	if input.Value() != "?!saved" {
		t.Fatal("reset lost focus", input.Value())
	}
	msg.SetDisabled(true)
	w.snapshot()
	w.typeText("no")
	if input.Value() != "?!saved" {
		t.Fatal("disabled accepted input")
	}
}
