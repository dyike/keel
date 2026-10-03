package window

import (
	"slices"
	"testing"

	"github.com/dyike/keel/ui/kit"
)

func TestToggleGroupItemPresentation(t *testing.T) {
	calls, sourceCalls := 0, 0
	source := kit.Toggle("", true).Icon(kit.IconStar).OnChange(func(bool) { sourceCalls++ })
	disabled := kit.Toggle("Locked", false)
	disabled.SetDisabled(true)
	group := kit.ToggleGroup("star", "lock", "plain").Multiple().Size(kit.ToggleSizeLarge).Variant(kit.ToggleOutline).
		Item("star", source).Item("lock", disabled).OnChange(func([]string) { calls++ })
	w := openTest(t, kitPage(group))
	e := element(t, w, "star")
	if e.Height != 40 || e.Selected == nil || *e.Selected {
		t.Fatal("item did not inherit group or leaked source selection", e)
	}
	if !element(t, w, "Locked").Disabled {
		t.Fatal("item disabled ignored")
	}
	w.press("Tab")
	w.press("Space")
	if !slices.Equal(group.Value(), []string{"star"}) || calls != 1 || sourceCalls != 0 || !source.Value() {
		t.Fatal("callback/state ownership", group.Value(), calls, sourceCalls)
	}
	// Changing the source does not silently modify the stored snapshot.
	source.Size(kit.ToggleSizeXSmall).Variant(kit.ToggleGhost)
	if element(t, w, "star").Height != 40 {
		t.Fatal("source mutation leaked")
	}
	group.Item("star", source)
	if element(t, w, "star").Height != 24 {
		t.Fatal("explicit size not applied")
	}
	w.press("Space")
	if len(group.Value()) != 0 {
		t.Fatal("replacing item lost focus")
	}
	w.press("Tab")
	w.press("Space")
	if !slices.Equal(group.Value(), []string{"plain"}) {
		t.Fatal("tab did not skip disabled item", group.Value())
	}
	group.Item("star", nil).Item("lock", nil)
	if element(t, w, "star").Height != 40 || element(t, w, "lock").Disabled {
		t.Fatal("override not restored")
	}
	group.SetValue("lock")
	group.Item("lock", disabled)
	if !slices.Equal(group.Value(), []string{"lock"}) || calls != 3 {
		t.Fatal("configuration changed selection or fired callback")
	}
	group.SetDisabled(true)
	w.snapshot()
	w.click(element(t, w, "star").center())
	if calls != 3 {
		t.Fatal("group disabled ignored")
	}
}
