//go:build windows

package window

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestWindowsFramelessAutoMenuRemainsAccessible(t *testing.T) {
	calls := 0
	useTestMenu(t, MenuItem{ID: "file", Title: "File", Children: []MenuItem{
		{ID: "run", Title: "Run", OnSelect: func() { calls++ }},
	}})
	w := newWindow(Options{Frameless: true, Content: views(text("content"))})
	h := uitest.NewFunc(w.layout)
	h.Click(12, 12)
	if len(w.menuView.path) != 1 {
		t.Fatal("default menu disappeared in a frameless Windows window")
	}
	h.Key(key.NameReturn, 0)
	if calls != 1 {
		t.Fatal("frameless menu action was not invoked")
	}
	w.opts.MenuDisplay = MenuDisplayHidden
	h.Frame()
	h.Click(12, 12)
	if len(w.menuView.path) != 0 {
		t.Fatal("hidden menu remained clickable")
	}
}
