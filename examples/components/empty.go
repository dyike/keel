package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("empty", "controls", func() core.Widget { return el.Embed(emptyGallery{}) })
}

type emptyGallery struct{}

func (emptyGallery) Render(cx *el.Context) el.Element {
	return el.Div().W(el.Full).Gap(16).Child(
		kit.Empty("暂无订单", "创建订单后将在这里显示。").Render(cx),
		el.Div().W(el.Dp(180)).Child(kit.Empty("没有搜索结果 0", "请尝试其他关键词 Search again，或调整筛选条件。").Render(cx)),
		kit.Empty("暂无通知", "").Render(cx),
	)
}
