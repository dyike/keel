package window

import (
	"github.com/dyike/keel/ui/core"
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

func TestKitHoverCardConfiguredPlacement(t *testing.T) {
	card := kit.HoverCard(kit.Button("Target", nil), el.ViewFunc(func(*el.Context) el.Element { return el.Text("Preview") })).
		Width(100).OpenDelay(0).CloseDelay(0).Placement(el.Right, el.Start).Offset(10)
	w := openTest(t, Options{Width: 500, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(40).Items(el.Start).Child(card.Render(cx)) }))})
	w.click(element(t, w, "Target").center())
	e := element(t, w, "Preview")
	if e.Role != "text" {
		t.Fatalf("preview semantics: %+v", e)
	}
	if err := w.press("esc"); err != nil {
		t.Fatal(err)
	}
	for _, e := range w.snapshot() {
		if e.Name == "Preview" {
			t.Fatal("Esc left preview visible")
		}
	}
}

func TestKitRichTooltipSnapshot(t *testing.T) {
	const action = "test.tooltip.preview"
	core.Bind(action, "mod+s")
	t.Cleanup(func() { core.Bind(action) })
	tip := kit.WithTooltip(kit.Button("Save", nil), "Save hint").Action(action).Placement(el.Bottom, el.Start).
		Content(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Child(el.Text("Rich preview").Bold(), el.Text("Version history"))
		}))
	w := openTest(t, Options{Width: 400, Height: 240, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(24).Child(tip.Render(cx)) }))})
	w.click(element(t, w, "Save").center())
	if e := element(t, w, "Save hint"); e.Role != "tooltip" {
		t.Fatalf("tooltip: %+v", e)
	}
	if e := element(t, w, "mod+s"); e.Role != "text" {
		t.Fatalf("binding: %+v", e)
	}
	element(t, w, "Rich preview")
	if err := w.press("esc"); err != nil {
		t.Fatal(err)
	}
	for _, e := range w.snapshot() {
		if e.Name == "Save hint" {
			t.Fatal("tooltip remained after Esc")
		}
	}
}
