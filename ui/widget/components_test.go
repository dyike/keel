package widget

import (
	"testing"

	"gioui.org/io/key"

	"github.com/dyike/keel/ui/internal/uitest"
)

func TestTableSortsNumbersAsNumbers(t *testing.T) {
	tb := Table(Col("name", 1), Col("amount", 1))
	tb.SetRows([][]string{{"a", "900"}, {"b", "1,200"}, {"c", "80"}})
	tb.SortBy(1, false)
	var got []string
	for _, i := range tb.order {
		got = append(got, tb.rows[i][0])
	}
	if want := "c a b"; join(got) != want {
		t.Fatalf("ascending by amount: %q, want %q", join(got), want)
	}
}

func TestTableClickSelectsAndKeysMove(t *testing.T) {
	var selected []int
	activated := -1
	tb := Table(Col("x", 1)).Height(200).OnSelect(func(r int) { selected = append(selected, r) }).OnActivate(func(r int) { activated = r })
	tb.SetRows([][]string{{"r0"}, {"r1"}, {"r2"}})
	h := uitest.New(tb)
	h.Click(50, 60) // header is ~38dp tall, rows ~40dp: this is row 0
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if tb.Selected() != 2 || join(ints(selected)) != "0 1 2" || activated != 2 {
		t.Fatalf("selected %d, OnSelect %v, activated %d", tb.Selected(), selected, activated)
	}
}

func TestSelectChooseOption(t *testing.T) {
	var got string
	s := Select("", "北京", "上海").OnChange(func(v string) { got = v })
	h := uitest.New(s)
	h.Click(50, 20)            // open
	h.Click(50, 20+40+4+36+18) // second option: below the field, 4dp gap, 4dp padding
	if s.Value() != "上海" || got != "上海" || s.open {
		t.Fatalf("value %q, OnChange %q, open %t", s.Value(), got, s.open)
	}
}

func TestSelectClosesOnOutsideClickAndEsc(t *testing.T) {
	s := Select("", "a", "b")
	h := uitest.New(s)
	h.Click(50, 20)
	h.Click(390, 290)
	if s.open {
		t.Fatal("still open after clicking outside")
	}
	h.Click(50, 20)
	h.Key(key.NameEscape, 0)
	if s.open || s.Value() != "" {
		t.Fatalf("open %t value %q after Esc", s.open, s.Value())
	}
}

func TestTabsSwitchPage(t *testing.T) {
	changed := -1
	tabs := Tabs().Add("一", Text("page one")).Add("二", Text("page two")).OnChange(func(i int) { changed = i })
	h := uitest.New(tabs)
	h.Click(80, 20) // second tab header
	if tabs.Current() != 1 || changed != 1 {
		t.Fatalf("current %d, OnChange %d", tabs.Current(), changed)
	}
}

func TestRadioAndSwitch(t *testing.T) {
	var radio string
	var on bool
	r := RadioGroup("", "a", "b").OnChange(func(v string) { radio = v })
	h := uitest.New(r)
	h.Click(10, 45) // second option, below the first (28dp + 6dp gap)
	if r.Value() != "b" || radio != "b" {
		t.Fatalf("radio %q OnChange %q", r.Value(), radio)
	}
	s := Switch("x", false).OnChange(func(v bool) { on = v })
	h = uitest.New(s)
	h.Click(48, 10) // on the label, which is part of the switch
	if !s.Value() || !on {
		t.Fatalf("switch %t OnChange %t", s.Value(), on)
	}
}

func TestDialogConfirmAndCancel(t *testing.T) {
	ok, cancelled := 0, 0
	d := Dialog().OnCancel(func() { cancelled++ })
	h := uitest.New(d)
	// Each show needs a frame before keys: an app redraws after the callback
	// that opened the dialog, and only then does the dialog listen for keys.
	d.Confirm("t", "m", func() { ok++ })
	h.Frame()
	h.Key(key.NameEscape, 0)
	d.Confirm("t", "m", func() { ok++ })
	h.Frame()
	h.Key(key.NameReturn, 0)
	d.Alert("t", "m", nil)
	h.Frame()
	h.Key(key.NameEscape, 0) // an alert has no cancel: Esc does nothing
	if ok != 1 || cancelled != 1 || !d.IsOpen() {
		t.Fatalf("ok %d cancelled %d open %t", ok, cancelled, d.IsOpen())
	}
}

func TestProgressClamps(t *testing.T) {
	p := Progress("x")
	p.SetValue(1.5)
	if p.Value() != 1 {
		t.Fatal(p.Value())
	}
}

func join(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += " "
		}
		out += v
	}
	return out
}

func ints(v []int) []string {
	var s []string
	for _, i := range v {
		s = append(s, string(rune('0'+i)))
	}
	return s
}
