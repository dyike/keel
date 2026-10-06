package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"time"
)

func init() {
	registerSection("tree", "data", func() core.Widget {
		n := func(id string, kids ...*kit.TreeNode) *kit.TreeNode {
			return &kit.TreeNode{ID: id, Label: id, Children: kids}
		}
		msg := demoText("Ctrl/Cmd selects multiple nodes; Shift selects a range. Drag before or after another node to reorder.", "Ctrl/Cmd 多选，Shift 连选，拖动到另一节点前后重排")
		tr := kit.Tree(n("ui", n("el", n("layout.go"), n("paint.go")), n("kit", n("menu.go"), n("table.go"))), n("docs", n("kit.md")), n("go.mod"), &kit.TreeNode{ID: "remote", Label: demoText("Load on demand", "按需加载"), Lazy: true}).MultiSelect().Reorderable(func(id, parent string, index int) {
			msg = fmt.Sprintf(demoText("%s → %s item %d", "%s → %s 的第 %d 项"), id, parent, index+1)
		}).Height(260)
		tr.RowHeight(36).RenderItem(func(state kit.TreeItemContext) el.View {
			return el.ViewFunc(func(cx *el.Context) el.Element {
				icon := kit.IconFile
				if state.HasChildren {
					icon = kit.IconFolder
				}
				row := el.Div().Row().Gap(6).Child(kit.Icon(icon).Size(16).Render(cx), el.Text(state.Label))
				if state.Loading {
					row.Child(kit.Spinner().Render(cx))
				}
				if state.Error != "" {
					row.Child(kit.Button(demoText("Retry", "重试"), state.Retry).Size(24).Render(cx))
				}
				return row
			})
		})
		tr.OnExpand(func(id string, on bool) { msg = fmt.Sprintf(demoText("%s expanded: %v", "%s 展开：%v"), id, on) })
		tr.OnLoad(func(id string, token uint64) {
			go func() {
				time.Sleep(250 * time.Millisecond)
				core.Update(func() {
					_, err := tr.SetChildResults(id, token, n(id+"/readme.md"), n(id+"/config.json"))
					if err != nil {
						msg = err.Error()
					}
				})
			}()
		})
		tr.SetExpanded("ui", true)
		tr.SetNodeDisabled("paint.go", true)
		tr.OnSelectionChange(func(ids []string) { msg = fmt.Sprintf(demoText("Selected %d nodes", "选中 %d 个节点"), len(ids)) })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).W(el.Dp(420)).MaxW(el.Full).Child(tr.Render(cx), kit.Button(demoText("Move go.mod into docs", "将 go.mod 移入 docs"), func() {
				if err := tr.MoveNode("go.mod", "docs", 0); err != nil {
					msg = err.Error()
				}
			}).Render(cx), el.Text(msg))
		}))
	})
}
