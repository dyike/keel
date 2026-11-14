package kit

import (
	"reflect"
	"testing"
)

func TestDiffIntralineGraphemeBoundaries(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want []diffRange
	}{
		{"cat\n", "cut\n", []diffRange{{1, 2, 1, 2}}},
		{"你好", "你们", []diffRange{{1, 2, 1, 2}}},
		{"e\u0301", "e\u0300", []diffRange{{0, 2, 0, 2}}},
		{"👨‍👩‍👧", "👨‍👩‍👦", []diffRange{{0, 5, 0, 5}}},
		{"👍🏻", "👍🏽", []diffRange{{0, 2, 0, 2}}},
		{"alpha beta gamma", "Alpha beta Gamma", []diffRange{{0, 1, 0, 1}, {11, 12, 11, 12}}},
		{"hello", "hello!", []diffRange{{5, 5, 5, 6}}},
		{"a\r", "a", []diffRange{{1, 2, 1, 1}}},
	} {
		budget := 1000000
		if got := diffIntraline(tc.a, tc.b, &budget); !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("%q -> %q: %v want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestDiffIntralineUnifiedRowsAndBudget(t *testing.T) {
	v := Diff("old", "new")
	if len(v.inline) != 1 || v.inline[0].oldUnified != 3 || v.inline[0].newUnified != 5 {
		t.Fatal("newline marker shifted highlight", v.inline)
	}
	budget := 0
	got := diffIntraline("abcd", "wxyz", &budget)
	if !reflect.DeepEqual(got, []diffRange{{0, 4, 0, 4}}) || budget != 0 {
		t.Fatal("budget fallback", got, budget)
	}
	v.decorate()
	if len(v.decorations[2].entries) < 4 {
		t.Fatal("missing whole-line and inline decorations")
	}
	before := v.Value()
	patch := v.Patch("file", "file")
	v.decorate()
	if v.Value() != before || v.Patch("file", "file") != patch {
		t.Fatal("highlight changed text")
	}
}
