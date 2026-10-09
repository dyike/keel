package base

import (
	"slices"
	"testing"
	"time"

	"github.com/dyike/keel/third_party/gio/io/key"
)

func TestListKeysSkipDisabled(t *testing.T) {
	off := map[int]bool{0: true, 2: true, 3: true, 7: true}
	l := List{Count: 8, Disabled: func(i int) bool { return off[i] }, Page: 5}
	for _, c := range []struct {
		name     key.Name
		from, to int
	}{
		{key.NameDownArrow, -1, 1},
		{key.NameUpArrow, -1, 6},
		{key.NameDownArrow, 1, 4},
		{key.NameUpArrow, 4, 1},
		{key.NameUpArrow, 1, 1}, // no wrap: stays
		{key.NameDownArrow, 6, 6},
		{key.NameHome, 5, 1},
		{key.NameEnd, 1, 6},
		{key.NamePageDown, 1, 6},
		{key.NamePageDown, 4, 6}, // 7 is disabled: back toward 4
		{key.NamePageUp, 6, 1},
	} {
		if got, ok := l.Key(string(c.name), c.from); !ok || got != c.to {
			t.Errorf("%s from %d = %d, %v; want %d", c.name, c.from, got, ok, c.to)
		}
	}
	if _, ok := l.Key("A", 1); ok {
		t.Error("a letter is not a navigation key")
	}
	if got, _ := (List{}).Key(string(key.NameDownArrow), -1); got != -1 {
		t.Errorf("empty list moved to %d", got)
	}
}

func TestListWraps(t *testing.T) {
	l := List{Count: 4, Wrap: true, Disabled: func(i int) bool { return i == 0 }}
	if got := l.Next(3, 1); got != 1 {
		t.Errorf("wrap down from 3 = %d, want 1", got)
	}
	if got := l.Next(1, -1); got != 3 {
		t.Errorf("wrap up from 1 = %d, want 3", got)
	}
	all := List{Count: 2, Disabled: func(int) bool { return true }}
	if all.First() != -1 || all.Last() != -1 {
		t.Error("no enabled item should give -1")
	}
}

func TestTypeahead(t *testing.T) {
	labels := []string{"Apple", "Banana", "Blueberry", "Cherry", "berry"}
	l := List{Count: len(labels), Disabled: func(i int) bool { return i == 3 }}
	label := func(i int) string { return labels[i] }
	var ta Typeahead
	now := time.Unix(0, 0)
	find := func(s string, cur int) int {
		i, _ := ta.Find(now, s, cur, l, label)
		now = now.Add(100 * time.Millisecond)
		return i
	}
	if got := find("B", 0); got != 1 {
		t.Fatalf("b = %d, want Banana", got)
	}
	if got := find("L", 1); got != 2 {
		t.Fatalf("bl = %d, want Blueberry", got)
	}
	now = now.Add(2 * time.Second)
	if got := find("B", 2); got != 4 {
		t.Fatalf("b after a pause from Blueberry = %d, want berry", got)
	}
	if got := find("B", 4); got != 1 {
		t.Fatalf("b again = %d, want to cycle to Banana", got)
	}
	now = now.Add(2 * time.Second)
	if got, ok := ta.Find(now, "C", 1, l, label); ok || got != 1 {
		t.Fatalf("disabled Cherry matched: %d, %v", got, ok)
	}
	if _, ok := Text("A", false); !ok {
		t.Error("a letter is typeahead text")
	}
	for _, name := range []string{string(key.NameUpArrow), string(key.NameReturn), string(key.NameEscape)} {
		if _, ok := Text(name, false); ok {
			t.Errorf("%q is a named key, not text", name)
		}
	}
	if _, ok := Text("Space", false); ok {
		t.Error("a named key is not typeahead text")
	}
	if _, ok := Text("A", true); ok {
		t.Error("a shortcut is not typeahead text")
	}
}

func TestSelectionClicks(t *testing.T) {
	order := []string{"a", "b", "c", "d", "e"}
	off := func(i int) bool { return i == 2 }
	var s Selection[string]
	s.Click(order, 1, false, false, off)
	s.Click(order, 4, true, false, off)
	if got := s.In(order); !slices.Equal(got, []string{"b", "d", "e"}) {
		t.Fatalf("shift range = %v", got)
	}
	s.Click(order, 0, false, true, off)
	s.Click(order, 3, false, true, off)
	if got := s.In(order); !slices.Equal(got, []string{"a", "b", "e"}) {
		t.Fatalf("toggles = %v", got)
	}
	s.Click(order, 2, false, false, off)
	if s.Len() != 3 {
		t.Fatal("a disabled click changed the selection")
	}
	s.Click(order, 3, false, false, off)
	if got := s.Indexes(order); !slices.Equal(got, []int{3}) {
		t.Fatalf("plain click = %v", got)
	}
	s.Keep(func(k string) bool { return k != "d" })
	if _, ok := s.Anchor(); ok || s.Len() != 0 {
		t.Fatal("Keep left a removed key")
	}
}

func TestDisclosure(t *testing.T) {
	var calls []bool
	d := Disclosure{OnChange: func(open bool) { calls = append(calls, open) }}
	d.Set(true)
	if !d.Open() || len(calls) != 0 {
		t.Fatal("Set should open silently")
	}
	if d.Change(true) {
		t.Fatal("no change reported a change")
	}
	d.Toggle()
	d.SetDisabled(true)
	if d.Change(true) || d.Open() {
		t.Fatal("a disabled disclosure opened")
	}
	if !slices.Equal(calls, []bool{false}) {
		t.Fatalf("OnChange calls = %v", calls)
	}
}
