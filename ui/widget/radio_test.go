package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestRadioDisabledAndSilentSetter(t *testing.T) {
	changes := 0
	r := RadioGroup("", "甲", "乙").OnChange(func(string) { changes++ })
	r.SetValue("甲")
	h := uitest.New(r)
	if changes != 0 {
		t.Fatal("setter notified")
	}
	clickNamed(t, h, "乙")
	if r.Value() != "乙" || changes != 1 {
		t.Fatal("radio did not activate")
	}
	r.SetDisabled(true)
	r.SetValue("甲")
	h.Frame()
	clickNamed(t, h, "乙")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if r.Value() != "甲" || changes != 1 {
		t.Fatal("disabled radio activated")
	}
	if _, focused := r.enum.Focused(); focused {
		t.Fatal("disabled radio retained focus")
	}
	r.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "乙")
	if r.Value() != "乙" || changes != 2 {
		t.Fatal("re-enabled radio did not recover")
	}
}
