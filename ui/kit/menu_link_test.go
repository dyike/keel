package kit

import (
	"gioui.org/io/key"
	"testing"
)

func TestMenuLinkDispatchAndDisabled(t *testing.T) {
	got := ""
	sub := Menu().Link("Docs", "https://example.com/docs")
	m := Menu().Sub("Resources", sub)
	m.OnLink(func(url string) {
		if m.Value() {
			t.Fatal("menu remained open")
		}
		got = url
	})
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	click(t, h, "Resources")
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if got != "https://example.com/docs" {
		t.Fatal("ancestor link handler", got)
	}
	got = ""
	sub.SetItemDisabled("Docs", true)
	click(t, h, "Open")
	click(t, h, "Resources")
	h.Frame()
	click(t, h, "Docs")
	if got != "" || !m.Value() {
		t.Fatal("disabled link activated")
	}
}

func TestMenuLinkErrorsAndOverride(t *testing.T) {
	failures := 0
	m := Menu().Link("Bad", "invalid:").OnLinkError(func(error) { failures++ })
	m.Trigger(Button("Open", m.Toggle))
	h := page(m)
	click(t, h, "Open")
	click(t, h, "Bad")
	if failures != 1 || m.Value() {
		t.Fatal("validation error not reported")
	}
	got := ""
	m.OnLink(func(url string) { got = url }).ExternalLinkIcon(false)
	click(t, h, "Open")
	click(t, h, "Bad")
	if got != "invalid:" || failures != 1 {
		t.Fatal("override did not own URL policy")
	}
}
