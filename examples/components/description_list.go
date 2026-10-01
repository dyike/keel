package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("description-list", "controls", func() core.Widget {
		return el.Embed(kit.DescriptionList(kit.Description{Label: "订单号", Text: "SO-123"}, kit.Description{Label: "客户", Text: "张三 / Ada Lovelace"}, kit.Description{Label: "备注", Text: "中英文 mixed description；值为空时仍保留字段名。"}, kit.Description{Label: "附加信息"}))
	})
}
