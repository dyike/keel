package kit

import (
	"fmt"
	"github.com/dyike/keel/third_party/gio/io/key"
	"github.com/dyike/keel/ui/el"
	"reflect"
	"testing"
)

func TestCommandKeywordIndicesAndEventOrder(t *testing.T) {
	var events []string
	keywords := []string{"account"}
	items := []CommandItem{{Title: "Disabled", Disabled: true}, {Title: "Profile", Group: "Settings", Keywords: keywords, Action: func() { events = append(events, "action") }}, {Title: "Billing", Group: "Settings"}}
	c := Command(items...).OnSelect(func(i int) { events = append(events, fmt.Sprintf("select:%d", i)) }).
		OnQuery(func(q string) { events = append(events, "query:"+q) }).
		OnConfirm(func(i int) { events = append(events, fmt.Sprintf("confirm:%d", i)) })
	keywords[0] = "external"
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	if !reflect.DeepEqual(events, []string{"select:1"}) {
		t.Fatal(events)
	}
	h.Key(key.NameDownArrow, 0)
	h.Frame()
	if events[len(events)-1] != "select:2" {
		t.Fatal(events)
	}
	events = nil
	h.Type("account")
	h.Frame()
	h.Frame()
	if !reflect.DeepEqual(events, []string{"select:1", "query:account"}) || !shown(h, "Profile") || shown(h, "Billing") {
		t.Fatal("keyword filter/order", events)
	}
	h.Key(key.NameReturn, 0)
	h.Frame()
	if !reflect.DeepEqual(events, []string{"select:1", "query:account", "action", "confirm:1"}) || c.Value() {
		t.Fatal("confirm index/order", events)
	}
}

func TestCommandCancelAndEscapeClear(t *testing.T) {
	cancelled := 0
	c := Command(CommandItem{Title: "Alpha"}).OnCancel(func() { cancelled++ })
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	h.Type("none")
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if !c.Value() || c.query != "" || cancelled != 0 {
		t.Fatal("clear query cancelled", c.Value(), c.query, cancelled)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if c.Value() || cancelled != 1 {
		t.Fatal("cancel count", cancelled)
	}
	c.SetValue(true)
	h.Frame()
	c.SetValue(false)
	h.Frame()
	c.SetValue(true)
	h.Frame()
	c.SetDisabled(true)
	h.Frame()
	if cancelled != 1 {
		t.Fatal("programmatic cancellation", cancelled)
	}
}

func TestCommandConfirmSnapshotAndReentrantSelection(t *testing.T) {
	confirmed := -1
	var c *CommandView
	c = Command(CommandItem{Title: "Original", Action: func() {
		c.SetItems(CommandItem{Title: "Replacement"})
		c.SetValue(true)
		c.OnConfirm(func(int) { t.Fatal("changed callback used for old event") })
	}}).
		OnConfirm(func(i int) { confirmed = i })
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if confirmed != 0 || !c.Value() || !shown(h, "Replacement") {
		t.Fatal("action changes overwritten", confirmed, c.Value())
	}
	ran := false
	c.SetItems(CommandItem{Title: "First"}, CommandItem{Title: "Second", Action: func() { ran = true }})
	c.OnSelect(func(i int) {
		if i == 1 {
			c.SetItems(CommandItem{Title: "New model"})
		}
	})
	h.Frame()
	h.Frame()
	click(t, h, "Second")
	h.Frame()
	if ran || !shown(h, "New model") {
		t.Fatal("selected callback executed stale command")
	}
}

func TestCommandSelectNoneAndInlineCancel(t *testing.T) {
	var selections []int
	cancel := 0
	c := Command(CommandItem{Title: "Alpha"}).Inline(true).OnSelect(func(i int) { selections = append(selections, i) }).OnCancel(func() { cancel++ })
	var cx *el.Context
	h := render(func(ctx *el.Context) el.Element { cx = ctx; return c.Render(ctx) })
	h.Frame()
	c.Focus(cx)
	h.Frame()
	h.Frame()
	h.Type("zzz")
	h.Frame()
	h.Frame()
	if !reflect.DeepEqual(selections, []int{0, -1}) {
		t.Fatal(selections)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	h.Frame()
	if !c.Value() || cancel != 0 || selections[len(selections)-1] != 0 {
		t.Fatal("inline clear", selections, cancel)
	}
	h.Key(key.NameEscape, 0)
	h.Frame()
	if c.Value() || cancel != 1 {
		t.Fatal("inline cancel", cancel)
	}
}

func TestCommandOutsideCancelAndActionlessConfirm(t *testing.T) {
	cancelled, confirmed := 0, -1
	c := Command(CommandItem{Title: "No action"}).OnCancel(func() { cancelled++ }).OnConfirm(func(i int) { confirmed = i })
	h := page(c)
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	h.Type("query")
	h.Frame()
	h.Click(399, 299)
	h.Frame()
	if c.Value() || cancelled != 1 {
		t.Fatal("outside should cancel rather than clear", c.Value(), cancelled)
	}
	c.SetValue(true)
	h.Frame()
	h.Frame()
	h.Frame()
	h.Key(key.NameReturn, 0)
	h.Frame()
	if confirmed != 0 || cancelled != 1 {
		t.Fatal("actionless confirmation", confirmed, cancelled)
	}
}

func TestCommandSearchCallbackMayCloseWithoutRestoringSelection(t *testing.T) {
	c := Command(CommandItem{Title: "First"}).OnSearch(func(string, uint64) {})
	c.SetValue(true)
	h := page(c)
	h.Frame()
	h.Frame()
	c.OnSearch(func(query string, _ uint64) {
		if query != "" {
			c.SetValue(false)
		}
	})
	h.Frame()
	h.Type("close")
	h.Frame()
	if c.Value() || c.selectionPending {
		t.Fatal("query processing restored closed state")
	}
}
