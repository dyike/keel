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
		selected := demoText("Click any tag", "点击任一标签")
		return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
			wrap := el.Div().Wrap().Gap(8).Items(el.Center)
			for _, label := range []string{"Go", demoText("Layout", "布局 Layout"), demoText("Text label", "中文标签"), demoText("Keyboard interaction", "键盘交互"), demoText("Horizontal scrolling", "横向滚动 ScrollX"), demoText("Variable height", "可变高度"), demoText("Theme", "主题 Theme")} {
				label := label
				wrap.Child(kit.Button(label, func() { selected = demoText("Selected: ", "已选：") + label }).Variant(kit.ButtonSecondary).Render(cx))
			}
			grid := el.Div().Grid(3).Gap(12)
			for i := range 7 {
				grid.Child(el.Div().P(16).MinW(el.Dp(90)).Rounded(8).Bg(theme.Subtle).Gap(8).Child(el.Text(fmt.Sprintf(demoText("Card %d", "卡片 %d"), i+1)).Bold(), el.Text([]string{demoText("Short description", "短说明"), demoText("This card has a longer description to check wrapping and equal heights within a row.", "这一张有更长的说明文字，用来检查自动换行以及同一行的高度对齐。"), demoText("Mixed 123 text", "Mixed 123 中文")}[i%3])))
			}
			return el.Div().ScrollY().P(24).Gap(20).Child(el.Text(demoText("Layout containers", "Layout · 布局容器")).TextSize(24).Bold(), el.Text(demoText("Layout arranges components. Resize the window to see how these two layouts adapt.", "Layout 用来排列组件。下面演示两种布局；缩窄窗口可观察变化。")), el.Text(demoText("Wrap · Buttons move to the next line when space runs out", "Wrap · 空间不足时，按钮整体移到下一行")).Bold(), wrap, el.Text(selected).TextColor(theme.Muted), el.Text(demoText("Grid · Three equal columns; cards in each row have equal height", "Grid · 三列等宽，同一行的卡片等高")).Bold(), el.Text(demoText("Three fixed columns; text wraps to column width and the longest content determines row height.", "列数固定为三列；文字随列宽换行，行高由最长内容决定。")).TextColor(theme.Muted), grid)
		}))
	})
}
