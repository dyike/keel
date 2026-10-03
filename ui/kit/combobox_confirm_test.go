package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/locale"
	"reflect"
	"testing"
)

func TestComboboxConfirmClosePaths(t *testing.T) {
	for _, mode := range []string{"choose", "escape", "outside", "toggle", "clear"} {
		t.Run(mode, func(t *testing.T) {
			c := Combobox("Choice", "a", "b").Clearable(true)
			c.SetValue("a")
			calls := 0
			var got []string
			c.OnConfirm(func(v []string) {
				calls++
				got = append([]string(nil), v...)
				if len(v) > 0 {
					v[0] = "external"
				}
			})
			h := page(c)
			clickClass(t, h, "Editor", "Choice")
			h.Key(key.NameDownArrow, 0)
			h.Frame()
			switch mode {
			case "choose":
				click(t, h, "b")
			case "escape":
				h.Key(key.NameEscape, 0)
			case "outside":
				h.Click(399, 299)
			case "toggle":
				click(t, h, locale.Current().Name(locale.Current().MoreOptions, "Choice"))
			case "clear":
				click(t, h, locale.Current().Name(locale.Current().Clear, "Choice"))
			}
			h.Frame()
			h.Frame()
			want := []string{"a"}
			if mode == "choose" {
				want = []string{"b"}
			}
			if mode == "clear" {
				want = nil
			}
			if calls != 1 || !reflect.DeepEqual(got, want) || c.open || c.Value() == "external" {
				t.Fatal(calls, got, c.Value(), c.open)
			}
		})
	}
}

func TestComboboxConfirmOrderReentrancyAndSilentSetters(t *testing.T) {
	c := Combobox("Choice", "a", "b").Multiple()
	var events []string
	c.OnValuesChange(func([]string) { events = append(events, "change") })
	c.OnConfirm(func(values []string) {
		events = append(events, "confirm")
		if !reflect.DeepEqual(values, []string{"a"}) {
			t.Fatal(values)
		}
		c.SetValue("replacement")
	})
	c.open = true
	c.choose("a")
	if !reflect.DeepEqual(events, []string{"change"}) || !c.open {
		t.Fatal(events, c.open)
	}
	c.confirmClose()
	if !reflect.DeepEqual(events, []string{"change", "confirm"}) || c.Value() != "replacement" {
		t.Fatal(events, c.Value())
	}
	c.open = true
	c.SetValue("a")
	c.open = true
	c.SetDisabled(true)
	c.SetDisabled(false)
	c.open = true
	c.Searchable(false)
	if len(events) != 2 {
		t.Fatal("program operation confirmed", events)
	}
	c.OnConfirm(nil)
	c.open = true
	c.confirmClose()
	if len(events) != 2 {
		t.Fatal("callback not removed")
	}
}

func TestComboboxConfirmSnapshotAfterChange(t *testing.T) {
	c := Combobox("Choice", "a", "b")
	var got []string
	c.OnChange(func(string) { c.SetValue("replacement") })
	c.OnConfirm(func(values []string) { got = values })
	c.open = true
	c.choose("b")
	if !reflect.DeepEqual(got, []string{"b"}) || c.Value() != "replacement" {
		t.Fatal(got, c.Value())
	}
}
