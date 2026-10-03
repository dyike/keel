package window

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"testing"
)

func TestKitTabsRichLabelsAndDisabledSnapshot(t *testing.T) {
	text := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
	v := kit.Tabs().Variant(kit.TabsSegmented).Add("First", text("First page")).AddItem(kit.TabItem{Title: "Blocked", Disabled: true, Icon: kit.IconLock}).AddItem(kit.TabItem{Title: "Rich", Content: text("Custom label"), Icon: kit.IconStar, Page: text("Rich page")})
	w := openTest(t, kitPage(v))
	if e := element(t, w, "Blocked"); e.Role != "tab" || !e.Disabled {
		t.Fatalf("disabled tab %+v", e)
	}
	w.click(element(t, w, "Blocked").center())
	if v.Value() != 0 {
		t.Fatal("disabled clicked")
	}
	w.click(element(t, w, "Rich").center())
	if v.Value() != 2 || roleOfName(w, "Rich") != "tab" {
		t.Fatal("rich name/activation")
	}
	if err := w.press("left"); err != nil {
		t.Fatal(err)
	}
	if v.Value() != 0 {
		t.Fatal("keyboard did not skip disabled")
	}
	v.SetItemDisabled(0, true)
	if e := element(t, w, "First"); !e.Disabled || v.Value() != 2 {
		t.Fatal("dynamic disable")
	}
	v.Move(2, 0)
	if roleOfName(w, "Rich") != "tab" || v.Value() != 0 {
		t.Fatal("move lost tab semantics")
	}
}
