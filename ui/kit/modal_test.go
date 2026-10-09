package kit

import (
	"image"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
	"github.com/dyike/keel/ui/theme"
)

func page(views ...el.View) *uitest.Harness {
	return uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		box := el.Div().P(20).Gap(12).Items(el.Start)
		for _, v := range views {
			box.Child(v.Render(cx))
		}
		return box
	})))
}

func TestDialogConfirmFocusAndEnter(t *testing.T) {
	dlg, oks := Dialog(""), 0
	h := page(dlg)
	dlg.Confirm("保存修改", "离开前保存吗？", func() { oks++ })
	h.Frame()
	h.Frame()
	if n, ok := semanticNode(h, "dialog"); !ok || n.Desc.Label != "保存修改" {
		t.Fatal("dialog not shown")
	}
	h.Key(key.NameReturn, 0) // focus starts on 确定
	h.Frame()
	if oks != 1 || dlg.Value() {
		t.Fatalf("Enter: oks=%d open=%v", oks, dlg.Value())
	}
}

func TestDialogDangerIsPersistentAndEscCancels(t *testing.T) {
	cancels, oks := 0, 0
	dlg := Dialog("").OnClose(func() { cancels++ })
	h := page(viewFunc(func(*el.Context) el.Element { return el.Text("背景") }), dlg)
	dlg.ConfirmDanger("删除订单", "删除后不能恢复。", "删除", func() { oks++ })
	h.Frame()
	h.Frame()
	if _, ok := semanticNode(h, "alertdialog"); !ok {
		t.Fatal("danger confirm is not an alertdialog")
	}
	h.Click(5, 295) // the scrim
	h.Frame()
	if !dlg.Value() {
		t.Fatal("scrim press closed a persistent dialog")
	}
	h.Key(key.NameReturn, 0) // focus starts on 取消
	h.Frame()
	if oks != 0 || dlg.Value() || cancels != 1 {
		t.Fatalf("Enter on 取消: oks=%d open=%v cancels=%d", oks, dlg.Value(), cancels)
	}
	dlg.ConfirmDanger("删除订单", "删除后不能恢复。", "删除", func() { oks++ })
	h.Frame()
	h.Key(key.NameEscape, 0)
	h.Frame()
	if dlg.Value() || oks != 0 || cancels != 2 {
		t.Fatal("Esc did not cancel")
	}
	dlg.ConfirmDanger("删除订单", "删除后不能恢复。", "删除", func() { oks++ })
	h.Frame()
	click(t, h, "删除")
	h.Frame()
	if oks != 1 || dlg.Value() || cancels != 2 {
		t.Fatalf("删除: oks=%d cancels=%d", oks, cancels)
	}
}

func TestDialogCustomBodyScrimClosesAndRestoresFocus(t *testing.T) {
	dlg := Dialog("编辑").Body(viewFunc(func(*el.Context) el.Element { return el.Text("表单") }))
	open := Button("打开", func() { dlg.SetValue(true) })
	h := page(open, dlg)
	click(t, h, "打开")
	h.Frame()
	if !shown(h, "表单") {
		t.Fatal("custom dialog not shown")
	}
	h.Click(5, 295)
	h.Frame()
	h.Frame()
	if dlg.Value() || shown(h, "表单") {
		t.Fatal("scrim did not close a plain dialog")
	}
}

func TestSheetSlidesAndCloses(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	s := Sheet(el.Right, "设置").Size(200).Body(viewFunc(func(*el.Context) el.Element { return el.Text("内容") }))
	s.SetValue(true)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().Child(s.Render(cx)) })
	h.Frame()
	sheet := func() image.Rectangle { n, _ := semanticNode(h, "dialog"); return n.Desc.Bounds }
	start := sheet()
	c.advance(h, SheetSlide)
	h.Frame()
	end := sheet()
	if end.Min.X != 200 || end.Dx() != 200 || end.Dy() != 300 {
		t.Fatalf("sheet bounds %v, want right edge 200×300", end)
	}
	if start.Min.X <= end.Min.X {
		t.Fatalf("sheet did not slide in: start %v end %v", start, end)
	}
	click(t, h, "关闭")
	h.Frame()
	if s.Value() {
		t.Fatal("close button did not close")
	}
	theme.SetReducedMotion(true)
	defer theme.SetReducedMotion(false)
	s.SetValue(true)
	h.Frame()
	if b := sheet(); b.Min.X != 200 {
		t.Fatalf("reduced motion still slid: %v", b)
	}
}

func TestNotifierTimeoutHoverQueueAndEscape(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	n := Notifier()
	dlgOpen := true
	h := c.harness(func(cx *el.Context) el.Element {
		if dlgOpen {
			cx.Overlay("dlg", el.Modal(el.Div().Name("对话框").Size(el.Dp(40))).OnDismiss(func() { dlgOpen = false }))
		}
		return el.Div().Child(n.Render(cx))
	})
	for i := 0; i < MaxNotifications+1; i++ {
		n.Notify(Notice{Title: "通知 " + string(rune('A'+i)), Timeout: time.Second})
	}
	keep := n.Notify(Notice{Title: "常驻", Timeout: -1})
	h.Frame()
	if !shown(h, "通知 A") || shown(h, "通知 F") {
		t.Fatal("expected only the first notices shown")
	}
	h.Key(key.NameEscape, 0) // reaches the dialog below, not the stack
	h.Frame()
	if dlgOpen {
		t.Fatal("notifications swallowed Esc")
	}
	x, y := center(bounds(h, "通知 A"))
	h.Move(x, y)
	c.advance(h, 1500*time.Millisecond)
	h.Frame()
	if !shown(h, "通知 A") || shown(h, "通知 B") {
		t.Fatal("hover should keep A; B should time out")
	}
	h.Move(1, 299)
	c.advance(h, 1500*time.Millisecond)
	h.Frame()
	if shown(h, "通知 A") || n.Len() != 1 || !shown(h, "常驻") {
		t.Fatalf("after timeouts: len=%d", n.Len())
	}
	n.Dismiss(keep)
	h.Frame()
	if n.Len() != 0 || shown(h, "常驻") {
		t.Fatal("Dismiss failed")
	}
}

func TestCopyButtonWritesAndConfirms(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	cb := CopyButton(func() string { return "SO-1001" })
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().P(10).Child(cb.Render(cx)) })
	click(t, h, "复制")
	h.Frame()
	if _, data, ok := h.Router.WriteClipboard(); !ok || string(data) != "SO-1001" {
		t.Fatalf("clipboard %q %v", data, ok)
	}
	if !shown(h, "已复制") {
		t.Fatal("no feedback")
	}
	c.advance(h, CopiedFeedback)
	h.Frame()
	if !shown(h, "复制") {
		t.Fatal("feedback did not reset")
	}
}

func TestDropdownButtonPlainAndSplit(t *testing.T) {
	saved, exported := 0, ""
	formats := Menu().Item("PDF", "", func() { exported = "pdf" })
	more := Menu().Item("另存为", "", nil)
	h := page(DropdownButton("导出", formats), DropdownButton("保存", more).Split(func() { saved++ }))
	click(t, h, "导出")
	h.Frame()
	click(t, h, "PDF")
	h.Frame()
	if exported != "pdf" {
		t.Fatal("plain dropdown did not run the item")
	}
	click(t, h, "保存")
	h.Frame()
	if saved != 1 || more.Value() {
		t.Fatal("split main action")
	}
	click(t, h, "保存 更多选项")
	h.Frame()
	if !more.Value() || !shown(h, "另存为") {
		t.Fatal("split arrow did not open the menu")
	}
}

func uitest_page(fn viewFunc) *uitest.Harness { return uitest.New(el.Root(fn)) }
