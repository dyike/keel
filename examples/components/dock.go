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
		searchFactory := func(state kit.DockPanelState) (kit.DockPanel, error) {
			input := kit.Input("").Placeholder("在文件中搜索")
			var query string
			if len(state.State) > 0 {
				if err := json.Unmarshal(state.State, &query); err != nil {
					return kit.DockPanel{}, err
				}
			}
			input.SetValue(query)
			return kit.DockPanel{ID: state.ID, Kind: "search", Title: state.Title, View: input,
				SaveState: func() (json.RawMessage, error) { return json.Marshal(input.Value()) }}, nil
		}
		search, _ := searchFactory(kit.DockPanelState{ID: "search", Title: "搜索"})
		var snapshot []byte
		status := "先输入搜索词，再保存、修改和恢复"
		var d *kit.DockView
		// The center view shows while no document is open.
		d = kit.Dock(el.ViewFunc(func(*el.Context) el.Element {
			return el.Div().Grow().P(24).Gap(8).Bg(theme.Surface).Child(el.Text("没有打开的文档").Bold(),
				el.Text("把面板拖到这里，或用面板菜单的“移到中间”打开为文档").TextSize(12).TextColor(theme.Muted))
		})).
			Panel(kit.DockPanel{ID: "files", Title: "文件", View: files}, kit.DockLeft).
			Panel(search, kit.DockLeft).
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
		if err := d.RegisterPanel("search", searchFactory); err != nil {
			panic(err)
		}
		d.Split("search", "files", kit.DockPlacementBottom)
		d.Split("app.go", "main.go", kit.DockPlacementRight)
		// Side panels narrow enough that the documents stay the widest region
		// in the gallery's 680dp window.
		l := d.Layout()
		l.LeftSize, l.RightSize, l.BottomSize = 180, 160, 140
		d.SetLayout(l)
		customSkin := &kit.DockSkin{
			Header: func(e *el.DivEl) { e.Bg(theme.Bg).Py(theme.SpaceSm) },
			Body:   func(e *el.DivEl) { e.P(theme.SpaceLg) },
			Tab: func(e *el.DivEl, selected bool) {
				if selected {
					e.Bg(theme.Primary).TextColor(theme.PrimaryText)
				}
			},
			Separator: func(e *el.DivEl) { e.Bg(theme.Primary) },
		}
		styled := false
		save := kit.Button("保存工作区", func() {
			state, err := d.Snapshot()
			if err != nil {
				status = err.Error()
				return
			}
			data, err := json.Marshal(state)
			if err != nil {
				status = err.Error()
				return
			}
			snapshot = data
			status = "已保存布局和搜索词"
		})
		restore := kit.Button("恢复工作区", func() {
			if len(snapshot) == 0 {
				status = "请先保存工作区"
				return
			}
			var state kit.DockState
			if err := json.Unmarshal(snapshot, &state); err != nil {
				status = err.Error()
				return
			}
			if err := d.Restore(state); err != nil {
				status = err.Error()
				return
			}
			status = "已恢复布局并重建搜索面板"
		})
		style := kit.Button("切换 Dock 外观", func() {
			styled = !styled
			if styled {
				d.Skin(customSkin)
			} else {
				d.Skin(nil)
			}
		})
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().Grow().Items(el.Stretch).Child(
				el.Div().Row().Wrap().Gap(theme.SpaceSm).P(theme.SpaceSm).Child(save.Render(cx), restore.Render(cx), style.Render(cx)),
				el.Text(status).TextSize(theme.TextSm).TextColor(theme.Muted), d.Render(cx))
		}))
	})
}
