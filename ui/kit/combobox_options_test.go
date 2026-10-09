package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

func TestComboboxMultipleTagsAndOwnedValues(t *testing.T) {
	source := []string{"北京", "上海", "深圳"}
	c := Combobox("cities", source...).Multiple()
	source[0] = "external"
	calls := 0
	c.OnValuesChange(func(values []string) {
		calls++
		if len(values) > 0 {
			values[0] = "external"
		}
	})
	h := page(c)
	clickClass(t, h, "Editor", "cities")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"北京"}) || c.text != "" {
		t.Fatalf("first tag %v %q", c.Values(), c.text)
	}
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"北京", "上海"}) {
		t.Fatalf("second tag %v", c.Values())
	}
	click(t, h, "移除 北京")
	h.Frame()
	if !reflect.DeepEqual(c.Values(), []string{"上海"}) || calls != 3 {
		t.Fatalf("remove %v %d", c.Values(), calls)
	}
	before := calls
	c.SetValues([]string{"北京", "北京", "深圳"})
	values := c.Values()
	values[0] = "external"
	if calls != before || !reflect.DeepEqual(c.Values(), []string{"北京", "深圳"}) {
		t.Fatal("program assignment/ownership")
	}
}

func TestComboboxAsyncAndDisabledDraft(t *testing.T) {
	var tokens []uint64
	c := Combobox("remote").OnSearch(func(query string, token uint64) { tokens = append(tokens, token) })
	disabled := false
	h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(300)).Disabled(disabled).Child(c.Render(cx)) })
	clickClass(t, h, "Editor", "remote")
	h.Type("a")
	h.Frame()
	h.Type("b")
	h.Frame()
	if len(tokens) != 2 || !c.loading {
		t.Fatalf("requests %v", tokens)
	}
	if c.SetResults(tokens[0], "old") {
		t.Fatal("stale accepted")
	}
	c.SetSearchError(tokens[1], "failed")
	h.Frame()
	click(t, h, "重试")
	h.Frame()
	if len(tokens) != 3 {
		t.Fatal("retry missing")
	}
	if !c.SetResults(tokens[2], "semantic result") {
		t.Fatal("current result rejected")
	}
	h.Frame()
	h.Frame()
	click(t, h, "semantic result")
	h.Frame()
	if c.Value() != "semantic result" {
		t.Fatal("remote result filtered locally")
	}
	// Programmatic selection closes the popup and invalidates pending requests.
	clickClass(t, h, "Editor", "remote")
	h.Type("x")
	h.Frame()
	token := tokens[len(tokens)-1]
	c.SetValue("chosen")
	if c.SetResults(token, "late") {
		t.Fatal("setter accepted stale query")
	}
	c.OnSearch(nil).AllowCustom()
	clickClass(t, h, "Editor", "remote")
	h.Type("draft")
	h.Frame()
	disabled = true
	h.Frame()
	h.Frame()
	disabled = false
	h.Frame()
	h.Frame()
	if c.Value() != "chosen" || c.text != "chosen" {
		t.Fatalf("disabled committed draft %q %q", c.Value(), c.text)
	}
}

func TestComboboxVirtualResultsReveal(t *testing.T) {
	options := make([]string, 10000)
	for i := range options {
		options[i] = fmt.Sprintf("item-%05d", i)
	}
	c := Combobox("large", options...)
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element { cx = ctx; return el.Div().W(el.Dp(300)).Child(c.Render(ctx)) })
	clickClass(t, h, "Editor", "large")
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	h.Frame()
	a, b := c.virtual.visible(cx)
	if b-a > 40 {
		t.Fatalf("range %d %d", a, b)
	}
	c.active = 9990
	h.Key(key.NamePageDown, 0)
	h.Key(key.NamePageDown, 0)
	h.Frame()
	h.Frame()
	if c.active != 9999 || !shown(h, options[9999]) {
		t.Fatalf("last option %d", c.active)
	}
	h.Key(key.NameReturn, 0)
	if c.Value() != options[9999] {
		t.Fatal("keyboard choice")
	}
}
