package window

import (
	"image"
	"strconv"
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

// roleOfName is the role of the first element named name (a tab before its panel).
func roleOfName(w *Window, name string) string {
	for _, e := range w.snapshot() {
		if e.Name == name {
			return e.Role
		}
	}
	return ""
}

func TestKitDataSnapshot(t *testing.T) {
	rows := make([][]string, 500)
	for i := range rows {
		rows[i] = []string{"SO-" + strconv.Itoa(i), "华东"}
	}
	tb := kit.Table(kit.Col("单号"), kit.Col("客户")).Height(160)
	tb.SetRows(rows)
	l := kit.List("甲", "乙").Height(80)
	tr := kit.Tree(&kit.TreeNode{ID: "ui", Label: "ui", Children: []*kit.TreeNode{{ID: "kit", Label: "kit"}}}).Height(80)
	vl := kit.VirtualList(1000, 20, func(cx *el.Context, i int) el.Element { return el.Text("日志 " + strconv.Itoa(i)) }).Height(60)
	p := kit.Pagination(50, 10)
	o := kitPage(tb, l, tr, vl, p)
	o.Width, o.Height = 640, 900
	w := openTest(t, o)
	if e := element(t, w, "SO-0 | 华东"); e.Role != "row" {
		t.Fatalf("row: %+v", e)
	}
	if roleOfName(w, "SO-400 | 华东") != "" {
		t.Fatal("virtualized rows beyond the viewport were listed")
	}
	for name, role := range map[string]string{"单号": "columnheader", "甲": "option", "ui": "treeitem", "日志 0": "text", "2": "button"} {
		if got := roleOfName(w, name); got != role {
			t.Errorf("%s: %q want %q", name, got, role)
		}
	}
	w.click(element(t, w, "SO-1 | 华东").center())
	if err := w.press("down"); err != nil {
		t.Fatal(err)
	}
	if tb.Value() != 2 {
		t.Fatalf("table keys: %d", tb.Value())
	}
}

func TestKitCommandAndChatSnapshot(t *testing.T) {
	cmd := kit.Command(kit.CommandItem{Title: "新建订单", Shortcut: "mod+n"})
	cmd.Toggle()
	w := openTest(t, kitPage(cmd))
	if e := element(t, w, "新建订单"); e.Role != "option" {
		t.Fatalf("command item: %+v", e)
	}
	say := el.ViewFunc(func(*el.Context) el.Element { return el.Text("你好") })
	sc := kit.MessageScroller([]string{"a", "b"}, 80, func(cx *el.Context, i int) el.Element {
		if i == 0 {
			return kit.Message("AI", say).Render(cx)
		}
		return kit.Bubble(say).Mine().Render(cx)
	})
	a := kit.Attachment("报价.pdf", 2048)
	w = openTest(t, Options{Width: 640, Height: 400, Content: el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Child(sc.Render(cx), a.Render(cx))
	}))})
	for name, role := range map[string]string{"AI": "article", "报价.pdf": "attachment"} {
		if got := roleOfName(w, name); got != role {
			t.Errorf("%s: %q want %q", name, got, role)
		}
	}
}

func TestKitMigratedSnapshot(t *testing.T) {
	txt := el.ViewFunc(func(*el.Context) el.Element { return el.Text("内容") })
	tabs := kit.Tabs().Add("基本", txt).Add("高级", txt)
	acc := kit.Accordion().Add("问题", txt)
	b := kit.Badge(3).Child(kit.Button("通知", nil))
	pr := kit.Progress("导入")
	pr.SetValue(0.5)
	ln := kit.Link("详情", nil)
	img := kit.Image(image.NewNRGBA(image.Rect(0, 0, 40, 20)), "缩略图")
	o := kitPage(tabs, acc, b, pr, ln, img)
	o.Height = 600
	w := openTest(t, o)
	for name, role := range map[string]string{"基本": "tab", "问题": "disclosure", "3": "badge", "导入": "progressbar", "详情": "link", "缩略图": "image"} {
		if got := roleOfName(w, name); got != role {
			t.Errorf("%s: %q want %q", name, got, role)
		}
	}
	w.click(element(t, w, "高级").center())
	if tabs.Value() != 1 {
		t.Fatal("agent click did not switch tabs")
	}
}
