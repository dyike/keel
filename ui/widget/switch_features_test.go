package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestSwitchLoadingDisabledAndKeyboard(t *testing.T) {
	n := 0
	s := Switch("同步", false).OnChange(func(bool) { n++ })
	h := uitest.New(s)
	clickNamed(t, h, "同步")
	h.Key(key.NameSpace, 0)
	if s.Value() || n != 2 {
		t.Fatal("switch keyboard failed")
	}
	s.SetLoading(true)
	h.Frame()
	clickNamed(t, h, "同步")
	h.Key(key.NameReturn, 0)
	if s.Value() || n != 2 {
		t.Fatal("loading switch changed")
	}
	s.SetLoading(false)
	s.SetDisabled(true)
	h.Frame()
	clickNamed(t, h, "同步")
	if s.Value() || n != 2 {
		t.Fatal("disabled switch changed")
	}
	s.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "同步")
	if !s.Value() || n != 3 {
		t.Fatal("re-enabled switch failed")
	}
}
