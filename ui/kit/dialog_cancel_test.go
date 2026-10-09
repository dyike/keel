package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestDialogCancelGuardAllPathsAndRetry(t *testing.T) {
	for _, path := range []string{"escape", "outside", "cancel", "close"} {
		t.Run(path, func(t *testing.T) {
			allowed := false
			checks, closes := 0, 0
			d := Dialog("").CloseButton(true).BeforeCancel(func() bool { checks++; return allowed }).OnClose(func() { closes++ })
			h := page(d)
			d.Confirm("Question", "Keep waiting?", nil)
			h.Frame()
			h.Frame()
			dismiss := func() {
				switch path {
				case "escape":
					h.Key(key.NameEscape, 0)
				case "outside":
					h.Click(5, 295)
				case "cancel":
					click(t, h, locale.Current().Cancel)
				case "close":
					click(t, h, locale.Current().Close)
				}
				h.Frame()
			}
			dismiss()
			if !d.Value() || checks != 1 || closes != 0 {
				t.Fatal("veto ignored", checks, closes)
			}
			dismiss()
			if !d.Value() || checks != 2 {
				t.Fatal("veto marked layer dismissed", checks)
			}
			allowed = true
			dismiss()
			if d.Value() || checks != 3 || closes != 1 {
				t.Fatal("retry failed", checks, closes)
			}
		})
	}
}

func TestDialogCancelGuardOwnerCleanupAndReplacement(t *testing.T) {
	disabled := false
	checks, closes := 0, 0
	d := Dialog("Original").BeforeCancel(func() bool { checks++; return false }).OnClose(func() { closes++ })
	d.SetValue(true)
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(d.Render(cx)) })
	disabled = true
	h.Frame()
	h.Frame()
	if d.Value() || checks != 0 || closes != 1 {
		t.Fatal("owner cleanup was vetoed")
	}
	disabled = false
	d.BeforeCancel(func() bool { d.Alert("New title", "Replacement", nil); return true })
	d.SetValue(true)
	h.Frame()
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !d.Value() || !shown(h, "Replacement") {
		t.Fatal("replacement closed")
	}
	d.BeforeCancel(nil)
	h.Key(key.NameEscape, 0)
	h.Frame()
	if d.Value() {
		t.Fatal("cleared guard still blocked")
	}
}
