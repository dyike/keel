package kit

import (
	"fmt"
	"github.com/dyike/keel/gio/io/key"
	"strings"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestDiffReviewAndSilentAssignments(t *testing.T) {
	initial := DiffValue{"one\ntwo\nthree\nfour\nfive\n", "one\nTWO\nthree\nFOUR\nfive\n"}
	v := Diff(initial.Before, initial.After)
	calls := 0
	v.OnChange(func(got DiffValue) {
		calls++
		if got != v.Value() {
			t.Fatal("stale callback")
		}
	})
	if v.Changes() != 2 || v.Current() != 0 {
		t.Fatal("initial changes", v.Changes(), v.Current())
	}
	v.SetCurrent(1)
	if !v.Revert() || v.Value().After != "one\nTWO\nthree\nfour\nfive\n" {
		t.Fatal("revert", v.Value())
	}
	if !v.Keep() || v.Value().Before != v.Value().After || v.Changes() != 0 || calls != 2 {
		t.Fatal("keep", v.Value(), calls)
	}
	v.SetValue(initial)
	v.SetCurrent(1)
	if calls != 2 {
		t.Fatal("assignment notified")
	}
	v.SetDisabled(true)
	if v.Keep() || v.Revert() || v.Value() != initial {
		t.Fatal("disabled mutation")
	}
	v.SetValue(DiffValue{"x", "y"})
	if v.Value().After != "y" {
		t.Fatal("disabled assignment blocked")
	}
}

func TestDiffContextExpansionPreservesReviewAndPatch(t *testing.T) {
	var before, after strings.Builder
	for i := range 40 {
		fmt.Fprintf(&before, "line-%d\n", i)
		if i == 10 || i == 30 {
			fmt.Fprintf(&after, "changed-%d\n", i)
		} else {
			fmt.Fprintf(&after, "line-%d\n", i)
		}
	}
	v := Diff(before.String(), after.String()).Mode(DiffModeUnified)
	calls := 0
	v.OnChange(func(DiffValue) { calls++ })
	v.SetCurrent(1)
	patch := v.Patch("file", "file")
	stats := v.Stats()
	if stats != (DiffStats{Added: 2, Deleted: 2, Changes: 2}) {
		t.Fatal(stats)
	}
	h := renderView(v, 600, 1)
	if strings.Contains(v.unified.Value(), " line-0\n") {
		t.Fatal("collapsed context contains distant line")
	}
	click(t, h, locale.Current().DiffExpand)
	if !strings.Contains(v.unified.Value(), " line-0\n") || !strings.Contains(v.unified.Value(), " line-39\n") {
		t.Fatal("expanded view omitted context")
	}
	if v.Current() != 1 || v.Stats() != stats || v.Patch("file", "file") != patch || calls != 0 {
		t.Fatal("expansion changed review state")
	}
	line, _ := v.unified.Cursor()
	if line != v.positions[1] {
		t.Fatal("expansion lost selected change")
	}
	click(t, h, locale.Current().DiffCollapse)
	if strings.Contains(v.unified.Value(), " line-0\n") || v.Current() != 1 {
		t.Fatal("collapse lost selected change")
	}
	v.Keep()
	if v.Stats() != (DiffStats{Added: 1, Deleted: 1, Changes: 1}) {
		t.Fatal("stats not updated after keep", v.Stats())
	}
}

func TestDiffReadOnlySelectionAndCopy(t *testing.T) {
	v := Diff("old\n", "new\n").Mode(DiffModeUnified)
	h := renderView(v, 600, 1)
	v.unified.Focus()
	h.Frame()
	h.Key("A", key.ModShortcut)
	h.Key("C", key.ModShortcut)
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != v.Patch("before", "after") {
		t.Fatalf("copy selected diff: %q %t", data, ok)
	}
	typeAt(h, "overwrite")
	if v.unified.Value() != v.Patch("before", "after") || v.Value().After != "new\n" {
		t.Fatal("read-only comparison was edited")
	}
	v.Mode(DiffModeSplit)
	h.Frame()
	if v.Value() != (DiffValue{"old\n", "new\n"}) || v.Current() != 0 {
		t.Fatal("presentation change lost state")
	}
}

func TestDiffReviewButtonsAndAncestorDisabled(t *testing.T) {
	v := Diff("a\nb\nc\n", "A\nb\nC\n")
	disabled := false
	h := renderView(viewFunc(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) }), 320, 1)
	click(t, h, locale.Current().DiffNext)
	if v.Current() != 1 {
		t.Fatal("next")
	}
	disabled = true
	h.Frame()
	click(t, h, locale.Current().DiffRevert)
	if v.Changes() != 2 {
		t.Fatal("ancestor disabled ignored")
	}
	disabled = false
	h.Frame()
	click(t, h, locale.Current().DiffRevert)
	if v.Value().After != "A\nb\nc\n" {
		t.Fatal(v.Value())
	}
	click(t, h, locale.Current().DiffKeep)
	if v.Changes() != 0 {
		t.Fatal("keep button")
	}
}

func TestDiffPatchRangesAndTerminators(t *testing.T) {
	for _, tc := range []struct{ a, b, header string }{{"", "new\n", "@@ -0,0 +1,1 @@"}, {"old\n", "", "@@ -1,1 +0,0 @@"}, {"old", "new", "@@ -1,1 +1,1 @@"}} {
		p := Diff(tc.a, tc.b).Patch("file", "file")
		if !strings.Contains(p, tc.header) {
			t.Fatal(p)
		}
		if tc.a == "old" && strings.Count(p, "\\ No newline at end of file") != 2 {
			t.Fatal("missing newline markers", p)
		}
	}
	if p := Diff("same", "same").Patch("file", "file"); p != "" {
		t.Fatal("unchanged patch", p)
	}
	if p := Diff("a", "b").Patch("x\ninjected", "x\ninjected"); strings.Contains(p, "\ninjected") {
		t.Fatal("header injection", p)
	}
}
