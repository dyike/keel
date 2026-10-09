package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

func TestSelectGroupsDisabledAndMultipleValues(t *testing.T) {
	s := Select("city").Multiple().Searchable()
	source := []SelectOption{{Value: "bj", Label: "北京", Group: "中国"}, {Value: "sh", Label: "上海", Group: "中国", Disabled: true}, {Value: "ny", Label: "New York", Group: "美国"}}
	s.SetEntries(source...)
	source[0].Label = "external"
	calls := 0
	s.OnValuesChange(func(values []string) {
		calls++
		if len(values) > 0 {
			values[0] = "external"
		}
	})
	h := page(s)
	clickRole(t, h, "select", "city")
	h.Frame()
	if !shown(h, "中国") || !shown(h, "北京") || shown(h, "external") {
		t.Fatal("group/ownership")
	}
	click(t, h, "上海")
	if len(s.Values()) != 0 {
		t.Fatal("disabled option selected")
	}
	click(t, h, "北京")
	if !s.open || !reflect.DeepEqual(s.Values(), []string{"bj"}) {
		t.Fatal("multi closed or lost value")
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if !reflect.DeepEqual(s.Values(), []string{"bj", "ny"}) {
		t.Fatalf("keyboard %v", s.Values())
	}
	click(t, h, "北京")
	if !reflect.DeepEqual(s.Values(), []string{"ny"}) || calls != 3 {
		t.Fatalf("toggle %v %d", s.Values(), calls)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	before := calls
	s.SetValues([]string{"bj", "sh", "missing"})
	if calls != before || !reflect.DeepEqual(s.Values(), []string{"bj", "sh"}) {
		t.Fatal("program values")
	}
	entries := s.Entries()
	entries[0].Label = "external"
	if s.Entries()[0].Label != "北京" {
		t.Fatal("entries alias")
	}
	s.SetEntries(SelectOption{Value: "sh", Label: "上海"})
	if !reflect.DeepEqual(s.Values(), []string{"sh"}) {
		t.Fatal("removed option retained")
	}
}

func TestSelectVirtualizationInitialChoiceAndFocus(t *testing.T) {
	s := Select("large")
	options := make([]string, 10000)
	for i := range options {
		options[i] = fmt.Sprintf("option-%05d", i)
	}
	s.SetOptions(options...)
	s.SetValue(options[9999])
	var cx *el.Context
	h := render(func(c *el.Context) el.Element { cx = c; return el.Div().W(el.Dp(300)).Child(s.Render(c)) })
	clickRole(t, h, "select", "large")
	h.Frame()
	h.Frame()
	first, last := s.virtual.visible(cx)
	if last-first > 40 || !shown(h, options[9999]) {
		t.Fatalf("initial reveal %d %d", first, last)
	}
	h.Key(key.NameHome, 0)
	h.Frame()
	h.Frame()
	if !shown(h, options[0]) {
		t.Fatal("Home did not reveal first option")
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if s.Value() != options[0] || s.open {
		t.Fatal("selection failed")
	}
	h.Key(key.NameSpace, 0)
	h.Frame()
	if !s.open {
		t.Fatal("focus did not return to field")
	}
}

func TestSelectEntriesValidationIsAtomic(t *testing.T) {
	s := Select("test", "a")
	s.SetValue("a")
	for _, options := range [][]SelectOption{{{Value: ""}}, {{Value: "x"}, {Value: "x"}}} {
		func() {
			defer func() {
				if recover() == nil {
					t.Error("invalid values accepted")
				}
			}()
			s.SetEntries(options...)
		}()
		if s.Value() != "a" || s.Entries()[0].Value != "a" {
			t.Fatal("invalid data partly applied")
		}
	}
}

func TestSelectAllDisabledKeyboard(t *testing.T) {
	s := Select("disabled")
	s.SetEntries(SelectOption{Value: "a", Label: "A", Disabled: true})
	h := page(s)
	clickRole(t, h, "select", "disabled")
	h.Frame()
	for _, name := range []key.Name{key.NameUpArrow, key.NameDownArrow, key.NameHome, key.NameEnd, key.NameReturn} {
		h.Key(name, 0)
	}
	if s.Value() != "" {
		t.Fatal("selected disabled value")
	}
}
