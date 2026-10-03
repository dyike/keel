package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestMessageSlotsPreserveContent(t *testing.T) {
	input := kit.Input("Message draft")
	calls := 0
	footer := kit.Button("Footer action", func() { calls++ })
	msg := kit.Message("Assistant", input)
	w := openTest(t, kitPage(msg))
	w.click(element(t, w, "Message draft").center())
	w.typeText("saved")
	msg.Header(kit.Label("Sender metadata")).Avatar(kit.Avatar("Custom").Size(40)).Footer(footer)
	w.snapshot()
	w.typeText("!")
	if input.Value() != "!saved" {
		t.Fatal("inserting slots lost focus", input.Value())
	}
	head, body, foot := element(t, w, "Sender metadata"), element(t, w, "Message draft"), element(t, w, "Footer action")
	if head.Y >= body.Y || foot.Y <= body.Y {
		t.Fatal("slot order", head, body, foot)
	}
	msg.Header(nil).Avatar(nil).Footer(nil)
	w.snapshot()
	w.typeText("?")
	if input.Value() != "?!saved" {
		t.Fatal("removing slots lost focus", input.Value())
	}
	msg.DefaultAvatar().Footer(footer)
	w.click(element(t, w, "Footer action").center())
	if calls != 1 {
		t.Fatal("footer action")
	}
	msg.SetDisabled(true)
	w.click(element(t, w, "Footer action").center())
	if calls != 1 || !element(t, w, "Message draft").Disabled {
		t.Fatal("disabled inheritance")
	}
	found := false
	for _, e := range w.snapshot() {
		if e.Name == "Assistant" && e.Role == "article" {
			found = true
		}
	}
	if !found {
		t.Fatal("article semantics")
	}
}

func TestMessageUserAvatarPlacement(t *testing.T) {
	content := el.ViewFunc(func(*el.Context) el.Element { return el.Text("Body text") })
	avatar := el.ViewFunc(func(*el.Context) el.Element { return el.Div().W(el.Dp(30)).H(el.Dp(30)).Name("Identity") })
	msg := kit.Message("Me", content).User().Avatar(avatar).Header(kit.Label("Header")).Footer(kit.Label("Footer"))
	w := openTest(t, kitPage(msg))
	a, b := element(t, w, "Identity"), element(t, w, "Body text")
	if a.X <= b.X {
		t.Fatal("user avatar must trail content", a, b)
	}
	msg.Avatar(nil)
	for _, e := range w.snapshot() {
		if e.Name == "Identity" {
			t.Fatal("avatar not removed")
		}
	}
	msg.Content(nil)
	for _, e := range w.snapshot() {
		if e.Name == "Body text" {
			t.Fatal("body not removed")
		}
	}
	if roleOfName(w, "Header") == "" || roleOfName(w, "Footer") == "" {
		t.Fatal("metadata-only message")
	}
}
