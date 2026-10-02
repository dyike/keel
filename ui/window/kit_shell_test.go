package window

import (
	"testing"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func TestKitShellSnapshot(t *testing.T) {
	txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
	ran := ""
	bar := kit.Toolbar(kit.ToolbarItem{Label: "新建", Action: func() { ran = "新建" }}, kit.ToolbarItem{Label: "保存"})
	nav := kit.Sidebar().Section("", kit.SidebarItem{ID: "inbox", Label: "收件箱", Icon: kit.IconInbox})
	res := kit.Resizable(txt("左侧"), txt("右侧"))
	car := kit.Carousel(txt("第一张"), txt("第二张"))
	set := kit.Settings().Section("通用", kit.IconUser, kit.SettingItem{Label: "深色模式", Control: kit.Switch("", false)})
	o := kitPage(bar, nav, res, car)
	o.Height = 700
	w := openTest(t, o)
	w.render() // Measure the toolbar before inspecting its expanded commands.
	for name, role := range map[string]string{"新建": "button", "收件箱": "link", "调整大小": "separator", "下一张": "button"} {
		if got := roleOfName(w, name); got != role {
			t.Errorf("%s: %q want %q", name, got, role)
		}
	}
	w.click(element(t, w, "新建").center())
	w.click(element(t, w, "下一张").center())
	if ran != "新建" || car.Value() != 1 {
		t.Fatalf("toolbar %q carousel %d", ran, car.Value())
	}
	w = openTest(t, Options{Width: 640, Height: 400, Content: el.Root(set)})
	if got := roleOfName(w, "深色模式"); got != "group" {
		t.Fatalf("settings row: %q", got)
	}
	found := false
	for _, e := range w.snapshot() {
		if e.Role == "switch" && e.Name == "深色模式" {
			found = true
		}
	}
	if !found {
		t.Fatal("settings did not name its switch")
	}
}

func TestKitDockSnapshot(t *testing.T) {
	txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
	d := kit.Dock(txt("编辑器")).Panel(kit.DockPanel{ID: "files", Title: "文件", View: txt("文件树")}, kit.DockLeft)
	w := openTest(t, Options{Width: 800, Height: 500, Content: el.Root(d)})
	if got := roleOfName(w, "文件"); got != "region" {
		t.Fatalf("dock region: %q", got)
	}
	w.click(element(t, w, "更多 文件").center())
	w.click(element(t, w, "停靠到右侧").center())
	if l := d.Layout(); len(l.Right) != 1 {
		t.Fatalf("agent move: %+v", l)
	}
}

func TestKitDockNestedSnapshot(t *testing.T) {
	txt := func(s string) el.View { return el.ViewFunc(func(*el.Context) el.Element { return el.Text(s) }) }
	d := kit.Dock(txt("Center")).Panel(kit.DockPanel{ID: "a", Title: "Files", View: txt("File content")}, kit.DockLeft).
		Panel(kit.DockPanel{ID: "b", Title: "Search", View: kit.Input("").Placeholder("Find text")}, kit.DockLeft)
	d.Split("b", "a", kit.DockPlacementBottom)
	w := openTest(t, Options{Width: 800, Height: 500, Content: el.Root(d)})
	if roleOfName(w, "Files") != "region" || roleOfName(w, "Search") != "region" {
		t.Fatal("missing nested regions")
	}
	w.click(element(t, w, "更多 Search").center())
	w.click(element(t, w, "关闭").center())
	if d.Visible("b") {
		t.Fatal("close nested group")
	}
}
