package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/locale"
	"testing"
)

func TestDiffAgentReview(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	d := kit.Diff("before\n", "after\n").Name("Review")
	w := openTest(t, Options{Width: 390, Height: 700, Content: el.Root(d)})
	group := element(t, w, "Review")
	if group.Role != "group" || group.Value != "1 / 1" {
		t.Fatal(group)
	}
	button := element(t, w, "Revert change")
	if button.Role != "button" || button.Height < 44 || button.X+button.Width > 390 {
		t.Fatal(button)
	}
	d.SetDisabled(true)
	if !element(t, w, "Revert change").Disabled {
		t.Fatal("missing disabled state")
	}
	d.SetDisabled(false)
	w.click(element(t, w, "Revert change").center())
	if d.Value().After != "before\n" {
		t.Fatal("revert did not update")
	}
	element(t, w, "Unified diff")
}

func TestDiffAgentKeyboardNavigation(t *testing.T) {
	d := kit.Diff("a\nb\nc\n", "A\nb\nC\n")
	w := openTest(t, Options{Width: 390, Height: 700, Content: el.Root(d)})
	w.snapshot()
	if err := w.press("tab"); err != nil {
		t.Fatal(err)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if d.Current() != 1 {
		t.Fatal("keyboard did not activate next change", d.Current())
	}
}

func TestDiffAgentPatchInspection(t *testing.T) {
	old := locale.Current()
	locale.Apply(locale.English())
	defer locale.Apply(old)
	d := kit.Diff("", "")
	if err := d.SetPatch("--- a/file\n+++ b/file\n@@ -12 +12 @@\n-old\n+new\n@@ -30 +30 @@\n-before\n+after\n"); err != nil {
		t.Fatal(err)
	}
	w := openTest(t, Options{Width: 390, Height: 700, Content: el.Root(d)})
	if !element(t, w, "Keep change").Disabled || !element(t, w, "Revert change").Disabled {
		t.Fatal("patch-only mutations exposed")
	}
	w.click(element(t, w, "Next change").center())
	if d.Current() != 1 {
		t.Fatal("patch hunk navigation")
	}
	if element(t, w, "Copy").Disabled {
		t.Fatal("patch copy disabled")
	}
}

func TestDiffAgentEmbeddedHunkAction(t *testing.T) {
	d := kit.Diff("old\n", "new\n")
	calls := 0
	d.HunkActions(func(kit.DiffHunk) []kit.DiffAction {
		return []kit.DiffAction{{Label: "Inspect source region", OnSelect: func(h kit.DiffHunk) {
			calls++
			if h.Index != 0 || h.Before != "old\n" || h.After != "new\n" {
				t.Fatal(h)
			}
		}}}
	})
	w := openTest(t, Options{Width: 390, Height: 800, Content: el.Root(d)})
	action := element(t, w, "Inspect source region")
	if action.Role != "button" || action.Height < 44 {
		t.Fatal(action)
	}
	w.click(action.center())
	if calls != 1 {
		t.Fatal("pointer action", calls)
	}
	if err := w.press("enter"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatal("keyboard action", calls)
	}
}
