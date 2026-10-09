package window

import (
	"testing"

	"github.com/dyike/keel/third_party/gio/io/key"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/kit"
)

func TestShortcutFires(t *testing.T) {
	n := 0
	w := newWindow(Options{Content: views(text("x")), Shortcuts: map[string]func(){"mod+,": func() { n++ }}})
	h := uitest.NewFunc(w.layout)
	h.Key(",", key.ModShortcut)
	if n != 1 {
		t.Fatalf("shortcut fired %d times", n)
	}
}

func TestClickInsideRootInset(t *testing.T) {
	n := 0
	w := newWindow(Options{Content: views(kit.Button("go", func() { n++ }))})
	h := uitest.NewFunc(w.layout)
	h.Click(40, 40) // content starts 24dp in
	if n != 1 {
		t.Fatalf("callback ran %d times", n)
	}
}

func TestUpdateAppliesOnNextFrame(t *testing.T) {
	label := "old"
	h := uitest.New(el.Embed(el.ViewFunc(func(*el.Context) el.Element { return el.Text(label) })))
	core.Update(func() { label = "new" })
	h.Frame()
	if label != "new" {
		t.Fatal("update did not run")
	}
}

func TestParseShortcut(t *testing.T) {
	s, err := parseShortcut("Ctrl+Shift+s")
	if err != nil || s.name != "S" || s.mods != key.ModCtrl|key.ModShift {
		t.Fatalf("got %+v, %v", s, err)
	}
	for _, bad := range []string{"", "ctrl+", "a+b", "ctrl"} {
		if _, err := parseShortcut(bad); err == nil {
			t.Errorf("%q: expected error", bad)
		}
	}
}
