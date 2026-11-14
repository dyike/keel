package kit

import "testing"

func TestDiffHunkSnapshotAndTargetedReview(t *testing.T) {
	initial := DiffValue{"a\nb\nc\nd\n", "A\nb\nc\nD\n"}
	v := Diff(initial.Before, initial.After)
	h, ok := v.Hunk(1)
	want := DiffHunk{Index: 1, BeforeStart: 3, BeforeEnd: 4, AfterStart: 3, AfterEnd: 4, Before: "d\n", After: "D\n"}
	if !ok || h != want {
		t.Fatal(h, ok)
	}
	calls := 0
	v.OnChange(func(value DiffValue) {
		calls++
		if value != v.Value() {
			t.Fatal("stale value")
		}
	})
	if !v.RevertAt(1) || v.Value().After != "A\nb\nc\nd\n" || v.Current() != 0 || calls != 1 {
		t.Fatal(v.Value(), v.Current(), calls)
	}
	if h != want {
		t.Fatal("snapshot mutated")
	}
	if _, ok := v.Hunk(1); ok {
		t.Fatal("obsolete index accepted")
	}
	if !v.KeepAt(0) || v.Value().Before != v.Value().After || calls != 2 {
		t.Fatal("keep specific region")
	}
	v.SetValue(initial)
	v.SetCurrent(1)
	v.SetDisabled(true)
	if v.KeepAt(0) || v.RevertAt(0) || v.Current() != 1 || v.Value() != initial {
		t.Fatal("disabled review mutated")
	}
	v.SetDisabled(false)
	if v.KeepAt(-1) || v.RevertAt(2) || v.Current() != 1 || calls != 2 {
		t.Fatal("invalid index mutated")
	}
}

func TestDiffHunkInsertionDeletionAndEndings(t *testing.T) {
	for _, tc := range []struct {
		before, after string
		want          DiffHunk
	}{
		{"", "added\n", DiffHunk{AfterEnd: 1, After: "added\n"}},
		{"removed\n", "", DiffHunk{BeforeEnd: 1, Before: "removed\n"}},
		{"old\r\n", "new\r\n", DiffHunk{BeforeEnd: 1, AfterEnd: 1, Before: "old\r\n", After: "new\r\n"}},
		{"old", "new", DiffHunk{BeforeEnd: 1, AfterEnd: 1, Before: "old", After: "new"}},
	} {
		v := Diff(tc.before, tc.after)
		got, ok := v.Hunk(0)
		if !ok || got != tc.want {
			t.Fatal(got, tc.want)
		}
		patch := v.Patch("file", "file")
		if err := v.SetPatch(patch); err != nil {
			t.Fatal(err)
		}
		got, ok = v.Hunk(0)
		want := tc.want
		want.PatchOnly = true
		if !ok || got != want {
			t.Fatal("import changed known source", got, want)
		}
		if v.KeepAt(0) || v.RevertAt(0) {
			t.Fatal("patch inspection mutated")
		}
	}
}

func TestDiffHunkImportedSparseSourceRanges(t *testing.T) {
	v := Diff("", "")
	raw := "--- a/file\n+++ b/file\n@@ -900000000,2 +800000000,2 @@\n context\n-old\n+new\n@@ -900000010,0 +800000010,1 @@\n+inserted\n"
	if err := v.SetPatch(raw); err != nil {
		t.Fatal(err)
	}
	want := []DiffHunk{
		{Index: 0, BeforeStart: 899999999, BeforeEnd: 900000001, AfterStart: 799999999, AfterEnd: 800000001, Before: "context\nold\n", After: "context\nnew\n", PatchOnly: true},
		{Index: 1, BeforeStart: 900000010, BeforeEnd: 900000010, AfterStart: 800000009, AfterEnd: 800000010, After: "inserted\n", PatchOnly: true},
	}
	for i, w := range want {
		got, ok := v.Hunk(i)
		if !ok || got != w {
			t.Fatal(got, w)
		}
	}
	if err := v.SetPatch("invalid"); err == nil {
		t.Fatal("bad patch accepted")
	}
	if got, _ := v.Hunk(1); got != want[1] {
		t.Fatal("failed import changed snapshot")
	}
	v.SetValue(DiffValue{"x", "y"})
	if got, _ := v.Hunk(0); got.PatchOnly || got.Before != "x" || got.BeforeStart != 0 {
		t.Fatal("patch mode leaked", got)
	}
}

func TestDiffHunkMissingNewlineContext(t *testing.T) {
	v := Diff("", "")
	if err := v.SetPatch("--- a/f\n+++ b/f\n@@ -1,2 +1,2 @@\n-old\n+new\n tail\n\\ No newline at end of file\n"); err != nil {
		t.Fatal(err)
	}
	got, _ := v.Hunk(0)
	if got.Before != "old\ntail" || got.After != "new\ntail" {
		t.Fatal(got)
	}
}

func TestDiffHunkRejectsSourceAfterEndMarker(t *testing.T) {
	for _, raw := range []string{
		"--- a/f\n+++ b/f\n@@ -1,2 +1,1 @@\n-old\n\\ No newline at end of file\n-more\n+new\n",
		"--- a/f\n+++ b/f\n@@ -1,1 +1,1 @@\n-old\n+new\n\\ No newline at end of file\n@@ -3,1 +3,1 @@\n-before\n+after\n",
	} {
		v := Diff("original", "proposed")
		if err := v.SetPatch(raw); err == nil {
			t.Fatal("accepted impossible source", raw)
		}
		if got, _ := v.Hunk(0); got.Before != "original" {
			t.Fatal("invalid patch mutated view")
		}
	}
}
