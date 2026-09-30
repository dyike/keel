package widget

import (
	"testing"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestButtonClickRunsCallbackOnce(t *testing.T) {
	n := 0
	h := uitest.New(Button("go", func() { n++ }))
	h.Click(10, 10)
	h.Frame()
	if n != 1 {
		t.Fatalf("callback ran %d times", n)
	}
}

func TestDisabledButtonIgnoresClicks(t *testing.T) {
	n := 0
	b := Button("go", func() { n++ })
	b.SetDisabled(true)
	h := uitest.New(b)
	h.Click(10, 10)
	if n != 0 {
		t.Fatalf("disabled button fired %d times", n)
	}
}

func TestCheckboxToggles(t *testing.T) {
	var got []bool
	c := Checkbox("x", false).OnChange(func(v bool) { got = append(got, v) })
	h := uitest.New(c)
	h.Click(5, 5)
	h.Click(5, 5)
	if c.Value() || len(got) != 2 || !got[0] || got[1] {
		t.Fatalf("value %t, changes %v", c.Value(), got)
	}
}

func TestInputTypingAndSubmit(t *testing.T) {
	var changed, submitted string
	f := Input("").OnChange(func(s string) { changed = s }).OnSubmit(func(s string) { submitted = s })
	h := uitest.New(f)
	h.Click(20, 10) // focus the editor
	h.Type("你好")
	if f.Value() != "你好" || changed != "你好" {
		t.Fatalf("value %q, OnChange saw %q", f.Value(), changed)
	}
	h.Key("⏎", 0)
	if submitted != "你好" {
		t.Fatalf("OnSubmit saw %q", submitted)
	}
}

func TestSetValueDoesNotFireOnChange(t *testing.T) {
	n := 0
	f := Input("").OnChange(func(string) { n++ })
	h := uitest.New(f)
	f.SetValue("abc")
	h.Frame()
	if n != 0 {
		t.Fatalf("OnChange fired %d times for SetValue", n)
	}
	h.Click(20, 10)
	h.Type("d")
	if n != 1 || f.Value() == "abc" {
		t.Fatalf("after typing: %d calls, value %q", n, f.Value())
	}
}
