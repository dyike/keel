package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("tree", "data", func() core.Widget {
		n := func(id string, kids ...*kit.TreeNode) *kit.TreeNode {
			return &kit.TreeNode{ID: id, Label: id, Children: kids}
		}
		msg := "Ctrl/Cmd 多选，Shift 连选，拖动到另一节点前后重排"
		tr := kit.Tree(n("ui", n("el", n("layout.go"), n("paint.go")), n("kit", n("menu.go"), n("table.go"))), n("docs", n("kit.md")), n("go.mod")).MultiSelect().Reorderable(func(id, parent string, index int) { msg = fmt.Sprintf("%s → %s 的第 %d 项", id, parent, index+1) }).Height(260)
		tr.SetExpanded("ui", true)
		tr.SetNodeDisabled("paint.go", true)
		tr.OnSelectionChange(func(ids []string) { msg = fmt.Sprintf("选中 %d 个节点", len(ids)) })
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(12).W(el.Dp(420)).Child(tr.Render(cx), kit.Button("将 go.mod 移入 docs", func() {
				if err := tr.MoveNode("go.mod", "docs", 0); err != nil {
					msg = err.Error()
				}
			}).Render(cx), el.Text(msg))
		}))
	})
}
