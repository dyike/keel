package kit

import (
	"fmt"
	"strings"
	"testing"
)

func TestDiffPatchInputSparseAndReviewGuards(t *testing.T) {
	patch := "diff --git a/file b/file\n--- a/file\n+++ b/file\n@@ -12,2 +12,3 @@ first\n same\n-old\n+new\n+extra\n@@ -900000000,1 +900000001,1 @@ last\n-last\n+LAST\n\\ No newline at end of file\n"
	v := Diff("before", "after")
	calls := 0
	v.OnChange(func(DiffValue) { calls++ })
	if err := v.SetPatch(patch); err != nil {
		t.Fatal(err)
	}
	if v.Stats() != (DiffStats{Added: 3, Deleted: 2, Changes: 2}) {
		t.Fatal(v.Stats())
	}
	if v.old.Lines() > 10 || v.new.Lines() > 10 {
		t.Fatal("allocated omitted source lines")
	}
	if !strings.Contains(v.old.Value(), "900000000 last") {
		t.Fatal("missing source coordinates")
	}
	if !v.SetCurrent(1) || v.Keep() || v.Revert() || calls != 0 {
		t.Fatal("patch review guards")
	}
	v.Expanded(true)
	if v.Patch("ignored", "ignored") != patch || v.unified.Value() != patch {
		t.Fatal("modified imported patch")
	}
	if got, on := v.PatchInput(); !on || got != patch {
		t.Fatal("patch mode query")
	}
	v.SetValue(DiffValue{"left", "right"})
	if _, on := v.PatchInput(); on || !v.Revert() {
		t.Fatal("did not leave patch mode")
	}
}

func TestDiffPatchInputRejectsInvalidAtomically(t *testing.T) {
	header := "--- a/file\n+++ b/file\n"
	for _, bad := range []string{"garbage", header, header + "@@ -1,2 +1,1 @@\n-a\n+b\n", header + "@@ -0,1 +1,1 @@\n-a\n+b\n", header + "@@ -1 +1 @@\n-a\n+b\n+extra\n", header + "@@ -1 +1 @@\n\\ No newline at end of file\n-a\n+b\n", header + "@@ -999999999999999999999 +1 @@\n-a\n+b\n", header + "@@ -1 +1 @@\n-a\n+b\n" + header + "@@ -1 +1 @@\n-c\n+d\n"} {
		v := Diff("a", "b")
		before := v.Value()
		if err := v.SetPatch(bad); err == nil {
			t.Fatalf("accepted malformed patch %q", bad)
		}
		if v.Value() != before {
			t.Fatal("invalid patch changed comparison")
		}
	}
}

func TestDiffExportCanBeImported(t *testing.T) {
	for _, tc := range [][2]string{{"", "new\n"}, {"old\n", ""}, {"old", "new"}, {"a\r\nb\r\n", "a\r\nB\r\n"}} {
		original := Diff(tc[0], tc[1])
		patch := original.Patch("file", "file")
		v := Diff("", "")
		if err := v.SetPatch(patch); err != nil {
			t.Fatalf("%v\n%s", err, patch)
		}
		if v.Stats().Added != original.Stats().Added || v.Stats().Deleted != original.Stats().Deleted {
			t.Fatal("import counts")
		}
	}
}

func FuzzDiffPatchParser(f *testing.F) {
	f.Add("--- a/file\n+++ b/file\n@@ -1 +1 @@\n-a\n+b\n")
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 100000 {
			t.Skip()
		}
		p, err := parseDiffPatch(s)
		if err == nil && p.raw != s {
			t.Fatal(fmt.Sprintf("patch altered: %q", s))
		}
	})
}
