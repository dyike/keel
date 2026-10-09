package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestRadioDisabledOptionsAndKeyboard(t *testing.T) {
	calls := 0
	v := RadioGroup("Plan", "A", "B", "C").OnChange(func(string) { calls++ })
	v.SetOptionDisabled("B", true)
	h := page(v)
	click(t, h, "A")
	h.Key(key.NameRightArrow, 0)
	if v.Value() != "C" {
		t.Fatal("did not skip disabled option")
	}
	click(t, h, "B")
	if v.Value() != "C" {
		t.Fatal("disabled pointer changed selection")
	}
	click(t, h, "C")
	h.Key(key.NameHome, 0)
	if v.Value() != "A" {
		t.Fatal("Home")
	}
	h.Key(key.NameEnd, 0)
	if v.Value() != "C" {
		t.Fatal("End")
	}
	v.SetOptionDisabled("C", true)
	h.Frame()
	if v.Value() != "C" || v.FocusID() != v.itemID("A") {
		t.Fatal("disabled selected value/fallback focus")
	}
	v.SetOptionDisabled("A", true)
	h.Frame()
	before := calls
	h.Key(key.NameRightArrow, 0)
	if calls != before || v.FocusID() != "" {
		t.Fatal("all disabled")
	}
}

func TestRadioIndependentItemsShareState(t *testing.T) {
	v := RadioGroup("Plan", "A", "B", "C")
	disabledB := true
	h := render(func(cx *el.Context) el.Element {
		return el.Div().Role("radiogroup").Name("Plan").Child(
			el.Div().P(8).Child(v.Item("A").Render(cx), el.Text("Basic plan")),
			el.Div().P(8).Disabled(disabledB).Child(v.Item("B").Render(cx)),
			el.Div().P(8).Child(v.Item("C").Render(cx)))
	})
	click(t, h, "A")
	h.Key(key.NameRightArrow, 0)
	if v.Value() != "C" {
		t.Fatal("separate items did not skip disabled ancestor")
	}
	disabledB = false
	h.Frame()
	h.Key(key.NameLeftArrow, 0)
	if v.Value() != "B" {
		t.Fatal("separate keyboard state")
	}
	v.SetDisabled(true)
	h.Frame()
	click(t, h, "A")
	if v.Value() != "B" {
		t.Fatal("group disabled failed for independent item")
	}
}

func TestRadioOptionsOwnershipAndIdentity(t *testing.T) {
	options := []string{"A", "B", "C"}
	v := RadioGroup("Plan", options...)
	options[0] = "Changed"
	v.SetValue("B")
	id := v.FocusID()
	result := v.Options()
	result[1] = "Changed"
	if v.Options()[0] != "A" || v.Options()[1] != "B" {
		t.Fatal("aliased options")
	}
	v.SetOptions("C", "B", "B", "", "A")
	if v.FocusID() != id || len(v.Options()) != 3 {
		t.Fatal("identity or duplicate options")
	}
	v.SetOptions("A")
	if v.Value() != "" {
		t.Fatal("removed selection retained")
	}
	v.SetValue("unknown")
	if v.Value() != "" {
		t.Fatal("unknown value")
	}
}
