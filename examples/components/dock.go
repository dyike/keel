package main

import (
	"encoding/json"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("dock", "shell", func() core.Widget {
		txt := func(s string) el.View {
			return el.ViewFunc(func(*el.Context) el.Element { return el.Div().P(12).Child(el.Text(s).TextSize(13)) })
		}
		saved := "移动、关闭或调整面板后，这里显示保存的布局 JSON"
		files := kit.Tree(&kit.TreeNode{ID: "ui", Label: "ui", Children: []*kit.TreeNode{{ID: "kit", Label: "kit"}, {ID: "el", Label: "el"}}}).Fill().Plain()
		var d *kit.DockView
		d = kit.Dock(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Grow().P(24).Gap(8).Bg(theme.Surface).Child(el.Text("编辑器 Editor").Bold(), el.Text(saved).TextSize(12).TextColor(theme.Muted))
		})).
			Panel(kit.DockPanel{ID: "files", Title: "文件", View: files}, kit.DockLeft).
			Panel(kit.DockPanel{ID: "search", Title: "搜索", View: kit.Input("").Placeholder("在文件中搜索")}, kit.DockLeft).
			Panel(kit.DockPanel{ID: "outline", Title: "大纲", View: txt("func main()\nfunc render()")}, kit.DockRight).
			Panel(kit.DockPanel{ID: "terminal", Title: "终端", View: txt("$ go test ./...\nok")}, kit.DockBottom).
			OnLayoutChange(func(l kit.DockLayout) {
				b, _ := json.Marshal(l)
				saved = string(b)
			})
		d.Split("search", "files", kit.DockPlacementBottom)
		// Side panels narrow enough that the editor stays the widest region
		// in the gallery's 680dp window.
		l := d.Layout()
		l.LeftSize, l.RightSize, l.BottomSize = 180, 160, 140
		d.SetLayout(l)
		return el.Root(d)
	})
}
