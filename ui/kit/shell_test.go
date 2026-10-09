package kit

import (
	"encoding/json"
	"image"
	"testing"
	"time"

	"gioui.org/io/key"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func sized(w float32, views ...el.View) *uitest.Harness {
	return uitest.New(el.Root(viewFunc(func(cx *el.Context) el.Element {
		box := el.Div().W(el.Dp(w)).Items(el.Stretch).Gap(8)
		for _, v := range views {
			box.Child(v.Render(cx))
		}
		return el.Div().Items(el.Start).Child(box)
	})))
}

func TestResizableDragKeysAndLimits(t *testing.T) {
	r := Resizable(text("左"), text("右")).Min(50, 100)
	r.SetValue(150)
	h := sized(400, r)
	h.Frame()
	n, _ := semanticNode(h, "separator:150")
	b := n.Desc.Bounds
	x, y := float32(b.Min.X+3), float32(b.Min.Y+b.Dy()/2)
	h.Drag(x, y, x+60, y)
	h.Frame()
	if r.Value() != 210 {
		t.Fatalf("drag: %v", r.Value())
	}
	h.Key(key.NameEnd, 0) // limited by the second pane's minimum: 400-6-100
	if r.Value() != 294 {
		t.Fatalf("End: %v", r.Value())
	}
	h.Key(key.NameHome, 0)
	h.Key(key.NameRightArrow, 0)
	if r.Value() != 66 {
		t.Fatalf("Home+→: %v", r.Value())
	}
}

func TestSidebarSelectCollapseBadge(t *testing.T) {
	var picked string
	s := Sidebar().Section("工作台",
		SidebarItem{ID: "inbox", Label: "收件箱", Icon: IconInbox, Badge: 3},
		SidebarItem{ID: "orders", Label: "订单", Icon: IconCopy}).
		Section("", SidebarItem{ID: "settings", Label: "设置", Icon: IconUser}).
		OnChange(func(id string) { picked = id })
	h := page(s)
	click(t, h, "订单")
	h.Key(key.NameDownArrow, 0)
	h.Key(key.NameReturn, 0)
	if picked != "settings" || s.Value() != "settings" {
		t.Fatalf("picked %q", picked)
	}
	if _, ok := semanticNode(h, "badge:3"); !ok {
		t.Fatal("badge")
	}
	click(t, h, "收起侧栏")
	h.Frame()
	if !s.Collapsed() || !shown(h, "展开侧栏") {
		t.Fatal("collapse")
	}
	if n, _ := node(h, "订单"); n.Desc.Bounds.Dx() > 56 {
		t.Fatalf("collapsed item too wide: %v", n.Desc.Bounds)
	}
}

func TestToolbarOverflowAndArrowKeys(t *testing.T) {
	ran := ""
	items := []ToolbarItem{}
	for _, l := range []string{"新建", "打开", "保存", "导出", "打印", "分享"} {
		l := l
		items = append(items, ToolbarItem{Label: l, Action: func() { ran = l }})
	}
	tb := Toolbar(items...)
	h := sized(220, tb)
	h.Frame()
	h.Frame()
	if shown(h, "分享") || !shown(h, "更多") {
		t.Fatal("overflow did not move items into 更多")
	}
	click(t, h, "更多")
	click(t, h, "分享")
	if ran != "分享" {
		t.Fatalf("overflow item ran %q", ran)
	}
	click(t, h, "新建")
	h.Key(key.NameRightArrow, 0)
	h.Key(key.NameReturn, 0)
	if ran != "打开" {
		t.Fatalf("→ Enter ran %q", ran)
	}
	wide := Toolbar(items...)
	h = sized(600, wide)
	h.Frame()
	h.Frame()
	if !shown(h, "分享") || shown(h, "更多") {
		t.Fatal("wide toolbar should show everything")
	}
}

func TestTabsCloseAndOverflow(t *testing.T) {
	tabs := Tabs()
	for _, s := range []string{"main.go", "kit.go", "table.go", "menu.go", "README.md", "go.mod"} {
		tabs.Add(s, text("内容 "+s))
	}
	tabs.Closable(tabs.Remove)
	h := sized(260, tabs)
	h.Frame()
	h.Frame()
	if shown(h, "go.mod") || !shown(h, "更多") {
		t.Fatal("tabs did not overflow")
	}
	click(t, h, "更多")
	click(t, h, "go.mod")
	h.Frame()
	if tabs.Value() != 5 || !shown(h, "内容 go.mod") {
		t.Fatalf("menu choice %d", tabs.Value())
	}
	click(t, h, "关闭 main.go")
	if tabs.Len() != 5 || tabs.Value() != 4 {
		t.Fatalf("close: len %d current %d", tabs.Len(), tabs.Value())
	}
}

func TestSettingsSectionsAndSearch(t *testing.T) {
	s := Settings().
		Section("通用", IconUser, SettingItem{Label: "语言", Description: "界面显示的语言", Control: Select("", "中文", "English")}).
		Section("通知", IconInbox, SettingItem{Label: "邮件提醒", Control: Switch("", true)}, SettingItem{Label: "声音"})
	h := uitest.New(el.Root(s))
	if !shown(h, "语言") || shown(h, "邮件提醒") {
		t.Fatal("first section")
	}
	click(t, h, "通知")
	h.Frame()
	if !shown(h, "邮件提醒") || shown(h, "语言") {
		t.Fatal("switch section")
	}
	clickClass(t, h, "Editor", "搜索设置")
	h.Type("语言")
	h.Frame()
	if !shown(h, "语言") || shown(h, "邮件提醒") {
		t.Fatal("search across sections")
	}
}

func TestCarouselKeysDotsAutoplay(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	car := Carousel(text("第一张"), text("第二张"), text("第三张")).Autoplay(3 * time.Second)
	h := c.harness(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(300)).Child(car.Render(cx)) })
	click(t, h, "下一张")
	click(t, h, "3")
	if car.Value() != 2 {
		t.Fatalf("buttons/dots: %d", car.Value())
	}
	h.Move(390, 290) // away from the carousel
	c.advance(h, 3*time.Second)
	h.Frame()
	c.advance(h, carouselTransition)
	if car.Value() != 0 || !shown(h, "第一张") {
		t.Fatalf("autoplay wraps: %d", car.Value())
	}
}

func TestDockMoveCloseResizeAndLayout(t *testing.T) {
	var saved DockLayout
	d := Dock(text("编辑器")).
		Panel(DockPanel{ID: "files", Title: "文件", View: text("文件树")}, DockLeft).
		Panel(DockPanel{ID: "outline", Title: "大纲", View: text("大纲内容")}, DockLeft).
		Panel(DockPanel{ID: "term", Title: "终端", View: text("$ go test")}, DockBottom).
		OnLayoutChange(func(l DockLayout) { saved = l })
	root := el.Root(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(900, 600); root.Layout(gtx) })
	if !shown(h, "文件树") || !shown(h, "$ go test") || shown(h, "大纲内容") {
		t.Fatal("initial layout")
	}
	click(t, h, "大纲")
	click(t, h, "更多 大纲")
	click(t, h, "停靠到右侧")
	h.Frame()
	if saved.Right == nil || saved.Right[0] != "outline" || !shown(h, "大纲内容") || !shown(h, "文件树") {
		t.Fatalf("move right: %+v", saved)
	}
	click(t, h, "更多 终端")
	click(t, h, "关闭")
	h.Frame()
	if shown(h, "$ go test") || d.Visible("term") {
		t.Fatal("close")
	}
	data, _ := json.Marshal(d.Layout())
	other := Dock(text("编辑器")).
		Panel(DockPanel{ID: "files", Title: "文件", View: text("文件树")}, DockLeft).
		Panel(DockPanel{ID: "outline", Title: "大纲", View: text("大纲内容")}, DockLeft).
		Panel(DockPanel{ID: "term", Title: "终端", View: text("$ go test")}, DockBottom)
	var l DockLayout
	if err := json.Unmarshal(data, &l); err != nil {
		t.Fatal(err)
	}
	other.SetLayout(l)
	if got := other.Layout(); len(got.Right) != 1 || got.Right[0] != "outline" || other.Visible("term") {
		t.Fatalf("restored %+v", got)
	}
	other.SetVisible("term", true)
	if other.Layout().BottomActive != "term" {
		t.Fatal("reopen in its region")
	}
}

func TestCarouselDisabledCancelsNavigationAndTimer(t *testing.T) {
	c := &clock{now: time.Unix(100, 0)}
	calls := 0
	car := Carousel(text("one"), text("two")).Autoplay(3 * time.Second).OnChange(func(int) { calls++ })
	parentDisabled := false
	var cx *el.Context
	h := c.harness(func(ctx *el.Context) el.Element {
		cx = ctx
		return el.Div().W(el.Dp(300)).Disabled(parentDisabled).Child(car.Render(ctx))
	})
	cx.Focus(autoID("carousel", car))
	h.Frame()
	h.Frame()
	car.SetDisabled(true)
	h.Frame()
	h.Key(key.NameRightArrow, 0)
	click(t, h, "下一张")
	click(t, h, "2")
	h.Move(390, 290)
	c.advance(h, 10*time.Second)
	h.Frame()
	if car.Value() != 0 || calls != 0 || cx.Focused(autoID("carousel", car)) {
		t.Fatal("disabled carousel changed or retained focus")
	}
	car.SetValue(1)
	if calls != 0 || car.Value() != 1 {
		t.Fatal("programmatic disabled value")
	}
	car.SetDisabled(false)
	h.Frame()
	c.advance(h, 2*time.Second)
	if car.Value() != 1 {
		t.Fatal("timer resumed too soon")
	}
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 0 || calls != 1 {
		t.Fatal("timer did not resume")
	}
	// Inherited disabled must pause timers too, not only pointer dispatch.
	parentDisabled = true
	h.Frame()
	c.advance(h, 20*time.Second)
	h.Frame()
	if car.Value() != 0 || calls != 1 {
		t.Fatal("autoplay ignored inherited disabled")
	}
	parentDisabled = false
	h.Frame()
	c.advance(h, 2*time.Second)
	if car.Value() != 0 {
		t.Fatal("inherited pause retained expired deadline")
	}
	c.advance(h, time.Second)
	h.Frame()
	if car.Value() != 1 || calls != 2 {
		t.Fatal("inherited reenable did not restart timer")
	}
}

func TestResizableDisabledBlocksDragAndKeys(t *testing.T) {
	for _, vertical := range []bool{false, true} {
		calls, childClicks := 0, 0
		split := Resizable(Button("pane", func() { childClicks++ }), text("other")).Min(40, 40).OnChange(func(float32) { calls++ })
		if vertical {
			split.Vertical()
		}
		split.SetValue(120)
		h := render(func(cx *el.Context) el.Element { return el.Div().W(el.Dp(360)).H(el.Dp(260)).Child(split.Render(cx)) })
		n, _ := semanticNode(h, "separator:120")
		x, y := center(n.Desc.Bounds)
		h.Click(x, y)
		h.Key(key.NameRightArrow, 0)
		if split.Value() != 136 || calls != 1 {
			t.Fatal("enabled splitter did not respond")
		}
		split.SetDisabled(true)
		h.Frame()
		n, _ = semanticNode(h, "separator:136")
		if !n.Desc.Disabled {
			t.Fatal("disabled separator missing semantics")
		}
		x, y = center(n.Desc.Bounds)
		x1, y1 := x+50, y
		if vertical {
			x1, y1 = x, y+50
		}
		h.Drag(x, y, x1, y1)
		h.Key(key.NameEnd, 0)
		click(t, h, "pane")
		if split.Value() != 136 || calls != 1 || childClicks != 0 {
			t.Fatal("disabled splitter responded")
		}
		split.SetValue(100)
		h.Frame()
		if split.Value() != 100 || calls != 1 {
			t.Fatal("programmatic disabled assignment")
		}
		split.SetDisabled(false)
		h.Frame()
		n, _ = semanticNode(h, "separator:100")
		x, y = center(n.Desc.Bounds)
		h.Click(x, y)
		h.Key(key.NameRightArrow, 0)
		if split.Value() != 116 || calls != 2 {
			t.Fatal("reenabled splitter did not respond")
		}
	}
}
