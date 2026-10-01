package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestSelectDisabledClosesDropdownAndRecovers(t *testing.T) {
	changes := 0
	s := Select("", "甲", "乙").Hint("选择").OnChange(func(string) { changes++ })
	s.SetValue("甲")
	h := uitest.New(s)
	clickNamed(t, h, "选择")
	if !s.open || changes != 0 {
		t.Fatal("opening or silent setter failed")
	}
	s.SetDisabled(true)
	h.Frame()
	if s.open {
		t.Fatal("disabled dropdown remained open")
	}
	clickNamed(t, h, "选择")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if s.open || s.Value() != "甲" || changes != 0 {
		t.Fatal("disabled select activated")
	}
	if h.Router.Source().Focused(&s.box) {
		t.Fatal("disabled select retained focus")
	}
	s.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "选择")
	clickNamed(t, h, "乙")
	if s.Value() != "乙" || changes != 1 {
		t.Fatal("select did not recover")
	}
}
