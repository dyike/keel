package window

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestKitPopoverSnapshot(t *testing.T) {
	p := kit.Popover(el.ViewFunc(func(*el.Context) el.Element { return el.Text("筛选内容") }))
	p.Trigger(kit.Button("筛选", p.Toggle))
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(p.Render))})
	w.click(element(t, w, "筛选").center())
	element(t, w, "筛选内容")
	element(t, w, "筛选") // non-modal: the page stays visible to agents
	if err := w.press("escape"); err != nil {
		t.Fatal(err)
	}
	if p.Value() {
		t.Fatal("Esc did not close the popover")
	}
}

func TestKitTooltipSnapshot(t *testing.T) {
	tip := kit.WithTooltip(kit.Button("复制", nil), "复制到剪贴板")
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().P(60).Child(tip.Render(cx))
	}))})
	if err := w.press("tab"); err != nil { // keyboard focus shows the hint at once
		t.Fatal(err)
	}
	if e := element(t, w, "复制到剪贴板"); e.Role != "tooltip" {
		t.Fatalf("tooltip: %+v", e)
	}
}

func TestKitHoverCardSnapshot(t *testing.T) {
	card := kit.HoverCard(el.ViewFunc(func(*el.Context) el.Element { return el.Text("张三") }),
		el.ViewFunc(func(*el.Context) el.Element { return el.Text("华东区销售") }))
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(card.Render))})
	if e := element(t, w, "张三"); e.Role != "text" {
		t.Fatalf("anchor: %+v", e)
	}
}

func TestKitMenuSnapshotAndKeyboard(t *testing.T) {
	ran := ""
	m := kit.Menu().Item("复制", "mod+c", func() { ran = "copy" }).Item("删除", "", func() { ran = "delete" }).
		Sub("导出", kit.Menu().Item("PDF", "", nil))
	m.Trigger(kit.Button("更多", m.Toggle))
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(m.Render))})
	w.click(element(t, w, "更多").center())
	if e := element(t, w, "菜单"); e.Role != "menu" {
		t.Fatalf("menu: %+v", e)
	}
	if e := element(t, w, "导出"); e.Role != "menuitem" || e.Value != "submenu" {
		t.Fatalf("submenu item: %+v", e)
	}
	for _, e := range w.snapshot() {
		if e.Name == "更多" {
			t.Fatal("modal menu exposed the page behind it")
		}
	}
	for _, k := range []string{"down", "enter"} {
		if err := w.press(k); err != nil {
			t.Fatal(err)
		}
	}
	if ran != "delete" || m.Value() {
		t.Fatalf("keyboard ran %q, open=%v", ran, m.Value())
	}
}
