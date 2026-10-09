package kit

import (
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestDialogConfirmGuardRetryAndReuse(t *testing.T) {
	allowed := false
	checks, oks, closes := 0, 0, 0
	d := Dialog("").BeforeConfirm(func() bool { checks++; return allowed }).OnClose(func() { closes++ })
	h := page(d)
	d.Confirm("Validate", "Try again", func() {
		if d.Value() {
			t.Fatal("legacy OK should run after close")
		}
		oks++
	})
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !d.Value() || checks != 1 || oks != 0 || closes != 0 {
		t.Fatal("rejected confirmation changed state")
	}
	allowed = true
	h.Key(key.NameReturn, 0)
	h.Frame()
	if d.Value() || checks != 2 || oks != 1 || closes != 0 {
		t.Fatal("retry lost focus or callback")
	}
	d.SetValue(true)
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if d.Value() || oks != 2 {
		t.Fatal("programmatic reopen lost action")
	}
	oks = 1
	checks = 2
	allowed = false
	d.ConfirmDanger("Delete item", "Sure?", "Delete", func() { oks++ })
	h.Frame()
	h.Frame()
	click(t, h, "Delete")
	h.Frame()
	if !d.Value() || checks != 3 || oks != 1 {
		t.Fatal("danger did not retain guard")
	}
	click(t, h, locale.Current().Cancel)
	h.Frame()
	if d.Value() || closes != 1 || checks != 3 {
		t.Fatal("guard intercepted cancel")
	}
	d.BeforeConfirm(nil)
	d.Alert("Done", "OK", func() { oks++ })
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if d.Value() || oks != 2 || checks != 3 {
		t.Fatal("clearing guard")
	}
}

func TestDialogGuardCanReplaceOrDisableDialog(t *testing.T) {
	for _, replace := range []bool{false, true} {
		oks := 0
		d := Dialog("")
		d.BeforeConfirm(func() bool {
			if replace {
				d.Alert("Replacement", "New message", nil)
			} else {
				d.SetDisabled(true)
			}
			return true
		})
		h := page(d)
		d.Confirm("Original", "Message", func() { oks++ })
		h.Frame()
		h.Frame()
		h.Key(key.NameReturn, 0)
		h.Frame()
		if oks != 0 {
			t.Fatal("stale OK callback executed")
		}
		if replace && (!d.Value() || !shown(h, "New message")) {
			t.Fatal("replacement was closed")
		}
		if !replace && d.Value() {
			t.Fatal("disabled dialog reopened")
		}
	}
}
