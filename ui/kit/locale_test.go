package kit

import (
	"testing"

	"github.com/dyike/keel/ui/locale"
)

// Framework text follows locale.Apply on the next frame, without rebuilding views.
func TestFrameworkTextFollowsLocale(t *testing.T) {
	defer locale.Apply(locale.Chinese())
	dlg := Dialog("")
	cb := CopyButton(func() string { return "x" })
	h := page(cb, Spinner(), dlg)
	if !shown(h, "复制") || !shown(h, "加载中") {
		t.Fatal("default Chinese text missing")
	}
	locale.Apply(locale.English())
	dlg.Confirm("Save", "Save changes?", nil)
	h.Frame()
	h.Frame()
	for _, s := range []string{"OK", "Cancel"} {
		if !shown(h, s) {
			t.Errorf("dialog button %q missing", s)
		}
	}
	dlg.SetValue(false)
	h.Frame()
	if !shown(h, "Copy") || !shown(h, "Loading") {
		t.Fatal("English text missing after Apply")
	}
}
