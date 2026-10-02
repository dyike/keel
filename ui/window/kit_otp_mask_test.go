package window

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestOtpMaskEditingAndSnapshot(t *testing.T) {
	changes, completed := 0, ""
	otp := kit.OtpInput("PIN", 4).Masked(true).Groups(2).Size(40).
		OnChange(func(string) { changes++ }).OnComplete(func(s string) { completed = s })
	w := openTest(t, kitPage(otp))
	w.click(element(t, w, "PIN").center())
	if err := w.typeText("1234"); err != nil {
		t.Fatal(err)
	}
	if otp.Value() != "1234" || completed != "1234" || changes == 0 {
		t.Fatalf("value %q complete %q changes %d", otp.Value(), completed, changes)
	}
	assertMasked := func() {
		t.Helper()
		if e := element(t, w, "PIN"); e.Role != "textbox" || e.Value != "••••" {
			t.Fatalf("masked input: %+v", e)
		}
		for _, e := range w.snapshot() {
			if e.Name == "1" || e.Name == "2" || e.Name == "3" || e.Name == "4" || e.Value == "1234" {
				t.Fatalf("unmasked semantics: %+v", e)
			}
		}
	}
	assertMasked()
	before := changes
	core.Update(func() { otp.Groups(1).Size(56).Masked(false) })
	if e := element(t, w, "PIN"); e.Value != "1234" {
		t.Fatalf("unmask: %+v", e)
	}
	core.Update(func() { otp.Masked(true); otp.SetDisabled(true) })
	assertMasked()
	core.Update(func() { otp.SetValue("4321") })
	if changes != before {
		t.Fatal("programmatic change called OnChange")
	}
}
