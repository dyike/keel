package main

import (
	"encoding/json"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

func init() {
	registerSection("dock", "shell", func() core.Widget {
		txt := func(s string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element { return el.Div().P(12).Child(el.Text(s).TextSize(13)) })
		}
		code := func(s string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().Grow().P(16).Bg(theme.Surface).Child(el.Text(s).Mono().TextSize(13))
			})
		}
		saved := "移动、关闭、拆分或调整面板后，这里显示保存的布局 JSON"
		files := kit.Tree(&kit.TreeNode{ID: "ui", Label: "ui", Children: []*kit.TreeNode{{ID: "kit", Label: "kit"}, {ID: "el", Label: "el"}}}).Fill().Plain()
		var d *kit.DockView
		// The center view shows while no document is open.
		d = kit.Dock(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Grow().P(24).Gap(8).Bg(theme.Surface).Child(el.Text("没有打开的文档").Bold(),
				el.Text("把面板拖到这里，或用面板菜单的“移到中间”打开为文档").TextSize(12).TextColor(theme.Muted))
		})).
			Panel(kit.DockPanel{ID: "files", Title: "文件", View: files}, kit.DockLeft).
			Panel(kit.DockPanel{ID: "search", Title: "搜索", View: kit.Input("").Placeholder("在文件中搜索")}, kit.DockLeft).
			Panel(kit.DockPanel{ID: "main.go", Title: "main.go", View: code("package main\n\nfunc main() {\n    run()\n}")}, kit.DockCenter).
			Panel(kit.DockPanel{ID: "app.go", Title: "app.go", View: code("package main\n\nfunc run() {}")}, kit.DockCenter).
			Panel(kit.DockPanel{ID: "outline", Title: "大纲", View: txt("func main()\nfunc run()")}, kit.DockRight).
			Panel(kit.DockPanel{ID: "terminal", Title: "终端", View: txt("$ go test ./...\nok")}, kit.DockBottom).
			Panel(kit.DockPanel{ID: "layout", Title: "布局", View: el.ViewFunc(func(*el.Context) el.Element {
				return el.Div().P(12).Child(el.Text(saved).TextSize(12).TextColor(theme.Muted))
			})}, kit.DockBottom).
			OnLayoutChange(func(l kit.DockLayout) {
				b, _ := json.Marshal(l)
				saved = string(b)
			}).
			// "Open in new window", or a tab dragged out of the dock, opens the
			// panel in a window; closing that window puts it back.
			OnDetach(func(p kit.DockPanel, reattach func()) {
				window.Open(window.Options{Title: p.Title, Width: 520, Height: 400, Content: el.Root(p.View), OnClose: reattach})
			})
		d.Split("search", "files", kit.DockPlacementBottom)
		d.Split("app.go", "main.go", kit.DockPlacementRight)
		// Side panels narrow enough that the documents stay the widest region
		// in the gallery's 680dp window.
		l := d.Layout()
		l.LeftSize, l.RightSize, l.BottomSize = 180, 160, 140
		d.SetLayout(l)
		return el.Root(d)
	})
}
