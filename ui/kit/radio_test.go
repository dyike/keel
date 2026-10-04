package kit

import (
	"testing"

	"github.com/dyike/keel/ui/el"
)

// Standalone radios check on click and never uncheck themselves; the app
// keeps them exclusive in OnChange.
func TestStandaloneRadio(t *testing.T) {
	a, b := Radio("快递"), Radio("自提")
	a.SetValue(true)
	calls := 0
	a.OnChange(func(bool) { calls++; b.SetValue(false) })
	b.OnChange(func(bool) { calls++; a.SetValue(false) })
	h := page(viewFunc(func(cx *el.Context) el.Element { return el.Div().Gap(8).Child(a.Render(cx), b.Render(cx)) }))
	h.Frame()
	if n, ok := node(h, "快递"); !ok || !n.Desc.Selected {
		t.Fatal("checked radio not reported")
	}
	click(t, h, "自提")
	h.Frame()
	if !b.Value() || a.Value() || calls != 1 {
		t.Fatalf("click: a=%v b=%v calls=%d", a.Value(), b.Value(), calls)
	}
	click(t, h, "自提")
	if !b.Value() || calls != 1 {
		t.Fatal("clicking a checked radio unchecked it or called OnChange")
	}
	b.SetDisabled(true)
	b.SetValue(false)
	h.Frame()
	click(t, h, "自提")
	if b.Value() || calls != 1 {
		t.Fatal("a disabled radio checked itself")
	}
}
