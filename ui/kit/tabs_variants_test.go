package kit

import (
	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"testing"
)

func TestTabsVariantsDisabledAndRichLabels(t *testing.T) {
	for _, variant := range []TabsVariant{TabsUnderline, TabsPill, TabsOutline, TabsSegmented} {
		for _, scale := range []int{1, 2} {
			calls, closes := 0, 0
			v := Tabs().Variant(variant).Add("A", text("page A")).AddItem(TabItem{Title: "B", Page: text("page B"), Disabled: true, Icon: IconLock}).AddItem(TabItem{Title: "C", Page: text("page C"), Icon: IconStar, Content: text("Custom")}).OnChange(func(int) { calls++ }).Closable(func(int) { closes++ })
			h := renderView(v, 500, scale)
			click(t, h, "B")
			if v.Value() != 0 || calls != 0 {
				t.Fatal("disabled tab activated")
			}
			click(t, h, "A")
			h.Key(key.NameRightArrow, 0)
			if v.Value() != 2 || calls != 1 || !shown(h, "page C") || !shown(h, "Custom") {
				t.Fatal("skip disabled or rich label")
			}
			h.Key(key.NameRightArrow, 0)
			if v.Value() != 0 {
				t.Fatal("wrap")
			}
			h.Key(key.NameEnd, 0)
			if v.Value() != 2 {
				t.Fatal("end")
			}
			h.Key(key.NameHome, 0)
			if v.Value() != 0 {
				t.Fatal("home")
			}
			before := calls
			v.SetValue(1)
			v.SetItemDisabled(0, true)
			h.Frame()
			if v.Value() != 2 || calls != before {
				t.Fatal("programmatic disable callback/selection")
			}
			v.SetItemDisabled(2, true)
			h.Frame()
			h.Key(key.NameDeleteForward, 0)
			if closes != 0 {
				t.Fatal("disabled close")
			}
			for _, label := range []string{"A", "B", "C"} {
				n, ok := node(h, label)
				if !ok || !n.Desc.Disabled {
					t.Fatal("disabled semantics", label)
				}
			}
			v.SetItemDisabled(1, false)
			h.Frame()
			if v.Value() != 1 || calls != before {
				t.Fatal("enable recovery")
			}
			id := v.pages[1].id
			v.SetItem(1, TabItem{Title: "Renamed", Page: text("new page"), Content: el.ViewFunc(func(*el.Context) el.Element { return el.Text("New label") })})
			v.Move(1, 0)
			h.Frame()
			if v.pages[0].id != id || v.Value() != 0 || !shown(h, "new page") || !shown(h, "New label") {
				t.Fatal("update/move state")
			}
			v.Remove(0)
			h.Frame()
			if calls != before {
				t.Fatal("remove callback")
			}
		}
	}
}

func TestTabsDisabledOverflowAndDrag(t *testing.T) {
	v := Tabs().Add("First", text("first page")).AddItem(TabItem{Title: "Blocked", Disabled: true}).Add("Last", text("last page")).Reorderable(func(int, int) { t.Error("disabled drag") })
	h := sized(500, v)
	x, y := center(bounds(h, "Blocked"))
	tx, ty := center(bounds(h, "First"))
	h.Drag(x, y, tx, ty)
	if v.pages[1].Title != "Blocked" {
		t.Fatal("disabled dragged")
	}
	h = sized(120, v)
	for range 4 {
		h.Frame()
	}
	click(t, h, "更多")
	n, ok := node(h, "Blocked")
	if !ok || !n.Desc.Disabled {
		t.Fatal("disabled overflow missing")
	}
	click(t, h, "Blocked")
	if v.Value() != 0 {
		t.Fatal("disabled overflow selected")
	}
	click(t, h, "Last")
	if v.Value() != 2 {
		t.Fatal("enabled overflow not selected")
	}
}
