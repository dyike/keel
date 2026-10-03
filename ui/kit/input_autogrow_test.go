package kit

import (
	"strings"
	"testing"
)

func TestTextAreaAutoGrowWrappedContentAndLimits(t *testing.T) {
	for _, scale := range []int{1, 2} {
		v := TextArea("").Placeholder("Draft").AutoGrow(2, 4)
		calls := 0
		v.OnChange(func(string) { calls++ })
		h := renderView(v, 180, scale)
		height := func() int { return bounds(h, "Draft").Dy() }
		empty := height()
		v.SetValue("one\ntwo\nthree")
		h.Frame()
		three := height()
		if three <= empty {
			t.Fatalf("no growth: %d -> %d", empty, three)
		}
		v.SetValue(strings.Repeat("line\n", 30))
		h.Frame()
		capped := height()
		if capped <= three || capped > empty*2+4*scale {
			t.Fatalf("invalid cap: %d %d %d", empty, three, capped)
		}
		v.SetValue(strings.Repeat("wrapped text ", 120))
		h.Frame()
		if height() != capped {
			t.Fatalf("soft wrapping not capped: %d != %d", height(), capped)
		}
		v.SetValue("short")
		h.Frame()
		if height() != empty || calls != 0 {
			t.Fatalf("shrink or program callback: %d %d", height(), calls)
		}
		clickClass(t, h, "Editor", "Draft")
		h.Type(strings.Repeat("\nmore", 20))
		h.Frame()
		if height() != capped || calls == 0 {
			t.Fatalf("typing growth: %d calls %d", height(), calls)
		}
		// Mode changes retain editor identity and focus.
		v.Rows(3)
		h.Frame()
		before := v.Value()
		h.Type("!")
		h.Frame()
		if len(v.Value()) != len(before)+1 {
			t.Fatal("lost focus after Rows")
		}
		v.AutoGrow(2, 4).AutoGrow(5, 2)
		h.Frame()
		if height() != capped {
			t.Fatal("invalid range changed configuration")
		}
	}
}

func TestTextAreaAutoGrowFixedRowsAndSingleLine(t *testing.T) {
	v := TextArea("").Placeholder("Fixed").AutoGrow(3, 3)
	h := renderView(v, 180, 1)
	empty := bounds(h, "Fixed").Dy()
	v.SetValue(strings.Repeat("line\n", 30))
	h.Frame()
	if bounds(h, "Fixed").Dy() != empty {
		t.Fatal("fixed rows grew")
	}
	v.SetReadOnly(true)
	clickClass(t, h, "Editor", "Fixed")
	before := v.Value()
	h.Type("ignored")
	h.Frame()
	if v.Value() != before {
		t.Fatal("read-only changed")
	}
	single := Input("").Placeholder("Single")
	sh := renderView(single, 180, 1)
	normal := bounds(sh, "Single").Dy()
	single.AutoGrow(8, 10)
	sh.Frame()
	if bounds(sh, "Single").Dy() != normal {
		t.Fatal("single-line changed height")
	}
}
