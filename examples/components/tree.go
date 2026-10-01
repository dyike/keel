package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("tree", "data", func() core.Widget {
		n := func(id string, kids ...*kit.TreeNode) *kit.TreeNode {
			return &kit.TreeNode{ID: id, Label: id, Children: kids}
		}
		tr := kit.Tree(n("ui", n("el", n("layout.go"), n("paint.go")), n("kit", n("menu.go"), n("table.go"))), n("docs", n("kit.md")), n("go.mod")).Height(260)
		tr.SetExpanded("ui", true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element { return el.Div().P(24).W(el.Dp(300)).Child(tr.Render(cx)) }))
	})
}
