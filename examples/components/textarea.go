package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/layout"
	"github.com/dyike/keel/ui/widget"
)

func init() { registerSection("textarea", "inputs", textareaGallery) }
func textareaGallery() core.Widget {
	notes := widget.TextArea("备注：自动高度 1–4 行").Hint("输入多行或长句，观察高度变化").AutoHeight(1, 4)
	return layout.Card(widget.Heading("多行输入自动高度"), notes, widget.Button("填入多行", func() { notes.SetValue("第一行\n第二行\n第三行\n第四行\n第五行") }).Secondary(), widget.Button("清空备注", func() { notes.SetValue("") }).Secondary())
}
