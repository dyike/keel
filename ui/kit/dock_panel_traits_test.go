package kit

import (
	"encoding/json"
	"image"
	"testing"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/internal/uitest"
)

func TestDockPanelTabToolbarMenuAndLimits(t *testing.T) {
	refreshed := 0
	var tabSelected bool
	d := Dock(text("编辑器")).
		Panel(DockPanel{ID: "files", Title: "文件", Icon: IconFolder, View: text("文件树"),
			Toolbar: Button("刷新", func() { refreshed++ }),
			Menu:    func(m *MenuView) { m.Item("全部折叠", "", func() { refreshed += 10 }) },
			NoClose: true, NoZoom: true}, DockLeft).
		Panel(DockPanel{ID: "term", Title: "终端", NoPadding: true, View: text("$ go test"),
			Tab: func(selected bool) el.View { tabSelected = selected; return text("终端 ●") }}, DockBottom)
	root := el.Root(d)
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(900, 600); root.Layout(gtx) })
	h.Frame()
	if !shown(h, "终端 ●") || !tabSelected {
		t.Fatal("custom tab content")
	}
	click(t, h, "刷新")
	if refreshed != 1 {
		t.Fatal("toolbar button", refreshed)
	}
	click(t, h, "更多 文件")
	if shown(h, "关闭") || shown(h, "最大化") {
		t.Fatal("NoClose / NoZoom items still in the menu")
	}
	click(t, h, "全部折叠")
	if refreshed != 11 {
		t.Fatal("panel menu item", refreshed)
	}
	d.Zoom("files")
	if d.Zoomed() != "" {
		t.Fatal("NoZoom panel zoomed")
	}
	// Edge to edge: the body starts at the group's left edge.
	if body, group := bounds(h, "$ go test"), bounds(h, "终端"); body.Min.X != group.Min.X {
		t.Fatal("NoPadding body still inset", body, group)
	}
}

func TestDockRegionToggle(t *testing.T) {
	var saved DockLayout
	d := Dock(text("编辑器")).
		Panel(DockPanel{ID: "files", Title: "文件", View: text("文件树")}, DockLeft).
		Panel(DockPanel{ID: "term", Title: "终端", View: text("$ go test")}, DockBottom).
		OnLayoutChange(func(l DockLayout) { saved = l })
	toggle := d.RegionButton(DockLeft)
	root := el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		return el.Div().Grow().Child(toggle.Render(cx), d.Render(cx))
	}))
	h := uitest.NewFunc(func(gtx core.C) { gtx.Constraints.Max = image.Pt(900, 600); root.Layout(gtx) })
	h.Frame()
	if n, _ := node(h, "左侧面板"); !n.Desc.Selected || !shown(h, "文件树") {
		t.Fatal("open region shows selected")
	}
	click(t, h, "左侧面板")
	h.Frame()
	if shown(h, "文件树") || d.RegionOpen(DockLeft) || !saved.LeftClosed {
		t.Fatal("closing the region", saved.LeftClosed)
	}
	if n, _ := node(h, "左侧面板"); n.Desc.Selected {
		t.Fatal("closed region still selected")
	}
	data, _ := json.Marshal(d.Layout())
	d.SetRegionOpen(DockLeft, true)
	var l DockLayout
	if err := json.Unmarshal(data, &l); err != nil || !d.SetLayout(l) || d.RegionOpen(DockLeft) {
		t.Fatal("closed region restores", err)
	}
	d.SetRegionOpen(DockLeft, true)
	h.Frame()
	if !shown(h, "文件树") || !d.Visible("files") {
		t.Fatal("reopened region keeps its panel")
	}
	if !d.RegionOpen(DockCenter) {
		t.Fatal("the center is always open")
	}
}
