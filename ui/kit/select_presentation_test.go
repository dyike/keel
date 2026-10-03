package kit

import (
	"fmt"
	"testing"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/locale"
)

func TestSelectClearPresentationAndFocus(t *testing.T) {
	for _, multiple := range []bool{false, true} {
		v := Select("Choice", "a", "b").Clearable(true).TitlePrefix("Chosen:").Size(48)
		if multiple {
			v.Multiple()
		}
		v.SetValue("a")
		v.SetError("bad")
		changes, values := 0, 0
		v.OnChange(func(string) { changes++ })
		v.OnValuesChange(func([]string) { values++ })
		h := page(v)
		clear := locale.Current().Name(locale.Current().Clear, "Choice")
		click(t, h, clear)
		if v.Value() != "" || v.open || v.Error() != "" || changes != 1 || multiple && values != 1 {
			t.Fatal("clear", v.Value(), v.open, changes, values)
		}
		h.Key(key.NameSpace, 0)
		h.Frame()
		if !v.open {
			t.Fatal("clear did not restore field focus")
		}
		h.Key(key.NameReturn, 0)
		h.Frame()
		if !multiple {
			h.Key(key.NameSpace, 0)
			h.Frame()
			if !v.open {
				t.Fatal("selection close lost field focus")
			}
		}
	}
}
func TestSelectCustomRowsAndSearch(t *testing.T) {
	v := Select("Custom").Searchable().MenuWidth(360).MenuMaxHeight(180).RowHeight(44)
	entries := make([]SelectOption, 10000)
	for i := range entries {
		entries[i] = SelectOption{Value: fmt.Sprint(i), Label: fmt.Sprintf("Label %d", i)}
	}
	v.SetEntries(entries...)
	built := 0
	v.RenderItem(func(cx *el.Context, s SelectItemContext) el.Element {
		built++
		return el.Text("custom " + s.Option.Value)
	})
	v.RenderValue(func(cx *el.Context, s []SelectOption) el.Element {
		s[0].Label = "changed"
		return el.Text("picked " + s[0].Value)
	})
	v.Match(func(s SelectOption, q string) bool { return s.Value == q }).Empty(text("Nothing here"))
	h := page(v)
	clickRole(t, h, "select", "Custom")
	h.Frame()
	if built > 100 || !shown(h, "custom 0") {
		t.Fatal("virtual custom rows", built)
	}
	v.query = "9999"
	v.active = -1
	h.Frame()
	if !shown(h, "custom 9999") || shown(h, "custom 0") {
		t.Fatal("custom match")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if v.Value() != "9999" || !shown(h, "picked 9999") || v.entries[9999].Label != "Label 9999" {
		t.Fatal("display snapshot", v.Value())
	}
	clickRole(t, h, "select", "Custom")
	v.query = "missing"
	h.Frame()
	if !shown(h, "Nothing here") {
		t.Fatal("custom empty")
	}
}
func TestSelectClearInheritedDisabled(t *testing.T) {
	v := Select("Choice", "a").Clearable(true)
	v.SetValue("a")
	disabled := true
	h := render(func(cx *el.Context) el.Element { return el.Div().Disabled(disabled).Child(v.Render(cx)) })
	click(t, h, locale.Current().Name(locale.Current().Clear, "Choice"))
	if v.Value() != "a" {
		t.Fatal("ancestor disabled")
	}
	disabled = false
	v.SetDisabled(true)
	h.Frame()
	click(t, h, locale.Current().Name(locale.Current().Clear, "Choice"))
	if v.Value() != "a" {
		t.Fatal("own disabled")
	}
}
