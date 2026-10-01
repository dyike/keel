package window

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func TestKitDialogSnapshot(t *testing.T) {
	dlg, removed := kit.Dialog(""), 0
	page := el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(kit.Button("删除所选", func() { dlg.ConfirmDanger("删除订单", "确定删除？", "删除", func() { removed++ }) }).Render(cx), dlg.Render(cx))
	})
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(page)})
	w.click(element(t, w, "删除所选").center())
	if e := element(t, w, "删除订单"); e.Role != "alertdialog" {
		t.Fatalf("dialog: %+v", e)
	}
	for _, e := range w.snapshot() {
		if e.Name == "删除所选" {
			t.Fatal("modal dialog exposed the page")
		}
	}
	if err := w.press("escape"); err != nil {
		t.Fatal(err)
	}
	if dlg.Value() || removed != 0 {
		t.Fatal("Esc did not cancel")
	}
	w.click(element(t, w, "删除所选").center())
	w.click(element(t, w, "删除").center())
	if removed != 1 {
		t.Fatal("confirm did not run")
	}
}

func TestKitSheetSnapshot(t *testing.T) {
	theme.SetReducedMotion(true) // as KEEL_AUTOMATION does
	defer theme.SetReducedMotion(false)
	s := kit.Sheet(el.Right, "订单详情")
	s.SetValue(true)
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(s.Render))})
	if e := element(t, w, "订单详情"); e.Role != "dialog" {
		t.Fatalf("sheet: %+v", e)
	}
	w.click(element(t, w, "关闭").center())
	if s.Value() {
		t.Fatal("close button did not close")
	}
}

func TestKitNotifierSnapshot(t *testing.T) {
	n := kit.Notifier()
	n.Notify(kit.Notice{Title: "保存成功", Body: "订单已更新", Tone: kit.ToneSuccess})
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(n.Render(cx))
	}))})
	if e := element(t, w, "保存成功"); e.Role != "status" || e.Value != "success" {
		t.Fatalf("notice: %+v", e)
	}
	w.click(element(t, w, "关闭 保存成功").center())
	if n.Len() != 0 {
		t.Fatal("close did not dismiss")
	}
}

func TestKitCopyButtonSnapshot(t *testing.T) {
	cb := kit.CopyButton(func() string { return "SO-1001" })
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Embed(el.ViewFunc(cb.Render))})
	w.click(element(t, w, "复制").center())
	if e := element(t, w, "已复制"); e.Role != "button" {
		t.Fatalf("copy feedback: %+v", e)
	}
}

func TestKitDropdownButtonSnapshot(t *testing.T) {
	m := kit.Menu().Item("另存为", "", nil)
	d := kit.DropdownButton("保存", m).Split(func() {})
	w := openTest(t, Options{Width: 400, Height: 300, Content: el.Root(el.ViewFunc(d.Render))})
	w.click(element(t, w, "保存 更多选项").center())
	if e := element(t, w, "另存为"); e.Role != "menuitem" {
		t.Fatalf("dropdown menu: %+v", e)
	}
}
