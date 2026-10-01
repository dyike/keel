package widget

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestTabsDisabledAndSilentSetter(t *testing.T) {
	changes := 0
	tabs := Tabs().Add("甲", Text("第一页")).Add("乙", Text("第二页")).OnChange(func(int) { changes++ })
	focus := false
	h := uitest.NewFunc(func(gtx core.C) {
		if focus {
			gtx.Execute(key.FocusCmd{Tag: &tabs.clicks[1]})
		}
		tabs.Layout(gtx)
	})
	tabs.SetCurrent(1)
	h.Frame()
	if tabs.Current() != 1 || changes != 0 {
		t.Fatal("setter did not stay silent")
	}
	tabs.SetCurrent(0)
	tabs.SetDisabled(true)
	focus = true
	h.Frame()
	clickNamed(t, h, "乙")
	h.Key(key.NameSpace, 0)
	h.Key(key.NameReturn, 0)
	if tabs.Current() != 0 || changes != 0 || h.Router.Source().Focused(&tabs.clicks[1]) {
		t.Fatal("disabled tab activated or focused")
	}
	focus = false
	tabs.SetDisabled(false)
	h.Frame()
	clickNamed(t, h, "乙")
	if tabs.Current() != 1 || changes != 1 {
		t.Fatal("tab did not recover")
	}
}
