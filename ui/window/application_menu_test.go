package window

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/internal/uitest"
	"testing"
)

func TestMenuModelCopiesAndAtomicUpdates(t *testing.T) {
	calls := 0
	original := []MenuItem{{ID: "edit", Title: "Edit", Children: []MenuItem{{ID: "copy", Title: "Copy", Action: MenuCopy, OnSelect: func() { calls++ }}}}}
	bar, err := NewMenuBar(original...)
	if err != nil {
		t.Fatal(err)
	}
	original[0].Children[0].Title = "changed"
	snapshot := bar.Items()
	snapshot[0].Children[0].Disabled = true
	if bar.Items()[0].Children[0].Title != "Copy" || !bar.Invoke("copy") || calls != 1 {
		t.Fatal("model aliases input or edit override failed")
	}
	if err := bar.UpdateItem("copy", func(item *MenuItem) { item.ID = "edit" }); err == nil {
		t.Fatal("accepted duplicate ID")
	}
	if bar.Items()[0].Children[0].ID != "copy" {
		t.Fatal("invalid update modified model")
	}
	if err := bar.SetItems(MenuItem{Title: ""}); err == nil {
		t.Fatal("accepted blank title")
	}
	if err := bar.UpdateItem("edit", func(item *MenuItem) { item.Disabled = true }); err != nil {
		t.Fatal(err)
	}
	if bar.Invoke("copy") || calls != 1 {
		t.Fatal("disabled ancestor allowed action")
	}
	if bar.Invoke("edit") || bar.Invoke("missing") {
		t.Fatal("invoked non-action")
	}
	if err := bar.UpdateItem("edit", func(item *MenuItem) { item.Disabled = false }); err != nil {
		t.Fatal(err)
	}
	if err := bar.UpdateItem("copy", func(item *MenuItem) { item.Checked = true; item.Title = "Copy custom" }); err != nil {
		t.Fatal(err)
	}
	wire := wireMenuItems(bar.Items())[0].Children[0]
	if !wire.Checked || wire.Title != "Copy custom" || wire.Action != "" {
		t.Fatal("custom callback must override standard edit action", wire)
	}
}

func TestMenuValidation(t *testing.T) {
	for _, item := range []MenuItem{
		{Title: "Action", Shortcut: "ctrl+"},
		{Title: "Action", Role: "bogus"},
		{Title: "Action", Action: "bogus"},
		{Separator: true, OnSelect: func() {}},
		{Title: "Submenu", OnSelect: func() {}, Children: []MenuItem{{Title: "Child"}}},
	} {
		if _, err := NewMenuBar(item); err == nil {
			t.Errorf("accepted invalid item %+v", item)
		}
	}
}

func TestPortableMenuKeyboardAndDisabledState(t *testing.T) {
	calls := 0
	bar, err := NewMenuBar(MenuItem{ID: "file", Title: "File", Children: []MenuItem{{ID: "new", Title: "New", Shortcut: "mod+t", OnSelect: func() { calls++ }}}})
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(Options{Content: views(text("menu"))})
	w.shortcuts = menuShortcuts(bar)
	h := uitest.NewFunc(w.layout)
	h.Key("T", key.ModShortcut)
	if calls != 1 {
		t.Fatalf("keyboard shortcut called %d times", calls)
	}
	if err := bar.UpdateItem("file", func(item *MenuItem) { item.Disabled = true }); err != nil {
		t.Fatal(err)
	}
	// Already registered shortcut closures must honor later changes, too.
	h.Key("T", key.ModShortcut)
	if calls != 1 || len(menuShortcuts(bar)) != 0 {
		t.Fatal("disabled shortcut still active")
	}
}

func TestMenuWireKeys(t *testing.T) {
	for _, tt := range []struct {
		input, name string
		mods        uint32
	}{
		{"cmd+shift+t", "t", 9}, {"ctrl+alt+=", "=", 6}, {"cmd+return", "⏎", 1}, {"alt+left", "←", 4}, {"ctrl+F1", "f1", 2}, {"shift+tab", "tab", 8},
	} {
		bar, err := NewMenuBar(MenuItem{Title: "Key", Shortcut: tt.input})
		if err != nil {
			t.Fatal(err)
		}
		wire := wireMenuItems(bar.Items())[0]
		if wire.Key != tt.name || wire.Modifiers != tt.mods {
			t.Errorf("%s: %+v", tt.input, wire)
		}
	}
}

func TestMenuInstallationRejectsStaleNativeActions(t *testing.T) {
	t.Setenv("KEEL_AUTOMATION", "1")
	t.Setenv("KEEL_HEADLESS", "1")
	applicationMenu.Lock()
	oldBar, oldGeneration := applicationMenu.current, applicationMenu.generation
	applicationMenu.Unlock()
	defer func() {
		applicationMenu.Lock()
		applicationMenu.current = oldBar
		applicationMenu.generation = oldGeneration
		applicationMenu.Unlock()
	}()
	calls := 0
	bar, err := NewMenuBar(MenuItem{ID: "action", Title: "Action", OnSelect: func() { calls++ }})
	if err != nil {
		t.Fatal(err)
	}
	if err := bar.Install(); err != nil {
		t.Fatal(err)
	}
	applicationMenu.Lock()
	generation := applicationMenu.generation
	applicationMenu.Unlock()
	invokeNativeMenu(generation, "action")
	if err := bar.UpdateItem("action", func(item *MenuItem) { item.OnSelect = func() { calls += 10 } }); err != nil {
		t.Fatal(err)
	}
	invokeNativeMenu(generation, "action")
	if calls != 1 {
		t.Fatal("stale action dispatched after menu update")
	}
	applicationMenu.Lock()
	generation = applicationMenu.generation
	applicationMenu.Unlock()
	invokeNativeMenu(generation, "action")
	if calls != 11 {
		t.Fatal("updated callback did not dispatch")
	}
	replacement, _ := NewMenuBar(MenuItem{ID: "action", Title: "Replacement", OnSelect: func() { calls += 100 }})
	if err := replacement.Install(); err != nil {
		t.Fatal(err)
	}
	invokeNativeMenu(generation, "action")
	if calls != 11 {
		t.Fatal("old menu action dispatched to replacement")
	}
}
