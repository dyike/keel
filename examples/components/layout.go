package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("layout", "shell", func() core.Widget {
		selected := "点击任一标签"
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			wrap := el.Div().Wrap().Gap(8).Items(el.Center)
			for _, label := range []string{"Go", "布局 Layout", "中文标签", "键盘交互", "横向滚动 ScrollX", "可变高度", "主题 Theme"} {
				label := label
				wrap.Child(kit.Button(label, func() { selected = "已选：" + label }).Variant(kit.ButtonSecondary).Render(cx))
			}
			grid := el.Div().Grid(3).Gap(12)
			for i := range 7 {
				grid.Child(el.Div().P(16).MinW(el.Dp(90)).Rounded(8).Bg(theme.Subtle).Gap(8).Child(el.Text(fmt.Sprintf("卡片 %d", i+1)).Bold(), el.Text([]string{"短说明", "这一张有更长的说明文字，用来检查自动换行以及同一行的高度对齐。", "Mixed 123 中文"}[i%3])))
			}
			return el.Div().ScrollY().P(24).Gap(20).Child(el.Text("换行与网格").TextSize(24).Bold(), el.Text("缩窄窗口观察标签换行和网格文字排版。"), wrap, el.Text(selected).TextColor(theme.Muted), grid)
		}))
	})
}
