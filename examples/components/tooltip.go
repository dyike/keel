package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("tooltip", "overlays", func() core.Widget { return el.Root(newTooltipGallery()) })
}

type tooltipGallery struct{ copy, save *kit.TooltipView }

func newTooltipGallery() *tooltipGallery {
	return &tooltipGallery{
		copy: kit.WithTooltip(kit.Button("", nil).Icon(kit.IconCopy).Variant(kit.ButtonGhost), "复制 Copy（⌘C）"),
		save: kit.WithTooltip(kit.Button("保存", nil), "保存到本地，最多 30 个版本"),
	}
}

func (g *tooltipGallery) Render(cx *el.Context) el.Element {
	return el.Div().P(48).Gap(16).Items(el.Start).Child(
		el.Text("Tooltip：悬停 0.5 秒显示，Tab 聚焦立即显示，Esc 隐藏").Bold(),
		el.Div().Row().Gap(12).Child(g.copy.Render(cx), g.save.Render(cx)),
	)
}
