package kit

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func checkDiffRoundTrip(t testing.TB, before, after string) []diffRange {
	t.Helper()
	a, b, changes := diffText(before, after)
	var out strings.Builder
	old, new := 0, 0
	for _, c := range changes {
		if c.oldStart < old || c.newStart < new || c.oldEnd < c.oldStart || c.newEnd < c.newStart || c.oldEnd > len(a) || c.newEnd > len(b) {
			t.Fatalf("invalid range %+v", c)
		}
		if strings.Join(a[old:c.oldStart], "") != strings.Join(b[new:c.newStart], "") {
			t.Fatalf("unequal omitted region %+v", c)
		}
		out.WriteString(strings.Join(a[old:c.oldStart], ""))
		out.WriteString(strings.Join(b[c.newStart:c.newEnd], ""))
		old, new = c.oldEnd, c.newEnd
	}
	out.WriteString(strings.Join(a[old:], ""))
	if out.String() != after {
		t.Fatalf("reconstruction mismatch: %q -> %q, got %q", before, after, out.String())
	}
	return changes
}

func TestDiffModelExactText(t *testing.T) {
	for _, p := range [][2]string{{"", ""}, {"", "猫\n"}, {"a\n", ""}, {"a", "a\n"}, {"a\r\n", "a\n"}, {"one\ntwo\nthree\n", "one\nchanged\nthree\n"}, {"\xff\n", "\xfe\n"}, {"a\na\nb\n", "a\nb\na\n"}} {
		checkDiffRoundTrip(t, p[0], p[1])
	}
	if changes := checkDiffRoundTrip(t, "same\n", "same\n"); len(changes) != 0 {
		t.Fatal("unchanged text has changes")
	}
	changes := checkDiffRoundTrip(t, "one\ntwo\nthree\n", "one\nchanged\nthree\n")
	if len(changes) != 1 || changes[0] != (diffRange{1, 2, 1, 2}) {
		t.Fatal("isolated replacement", changes)
	}
}

func TestDiffModelRandomRoundTrips(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	for range 300 {
		var a, b strings.Builder
		for range r.Intn(100) {
			fmt.Fprintf(&a, "%d\n", r.Intn(12))
		}
		for range r.Intn(100) {
			fmt.Fprintf(&b, "%d\n", r.Intn(12))
		}
		checkDiffRoundTrip(t, a.String(), b.String())
	}
}

func TestDiffModelLargeAnchorsAndBudgetFallback(t *testing.T) {
	var a, b strings.Builder
	for i := range 5000 {
		fmt.Fprintf(&a, "line-%d\n", i)
		if i%250 == 0 {
			fmt.Fprintf(&b, "changed-%d\n", i)
		} else {
			fmt.Fprintf(&b, "line-%d\n", i)
		}
	}
	if c := checkDiffRoundTrip(t, a.String(), b.String()); len(c) != 20 {
		t.Fatal("unique anchors did not isolate edits", len(c))
	}
	checkDiffRoundTrip(t, strings.Repeat("a\n", 10000), strings.Repeat("b\n", 10000))
}

func BenchmarkDiffModelSparseLarge(b *testing.B) {
	var old, new strings.Builder
	for i := range 10000 {
		fmt.Fprintf(&old, "line-%d\n", i)
		if i%1000 == 0 {
			fmt.Fprintf(&new, "changed-%d\n", i)
		} else {
			fmt.Fprintf(&new, "line-%d\n", i)
		}
	}
	a, z := old.String(), new.String()
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		diffText(a, z)
	}
}

func FuzzDiffModelRoundTrip(f *testing.F) {
	for _, pair := range [][2]string{{"a\nb\n", "a\nc\n"}, {"", "hello"}, {"a\r\n", "a\n"}, {"repeat\nrepeat\n", "repeat\n"}} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, before, after string) {
		if len(before)+len(after) > 100_000 {
			t.Skip()
		}
		checkDiffRoundTrip(t, before, after)
	})
}
