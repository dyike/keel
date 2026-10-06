package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("label", "basic", func() core.Widget {
		masked := kit.Label(demoText("Account balance ¥12,345", "账户余额 ¥12,345")).Masked(true).Secondary(demoText("(masks the primary text only)", "（仅遮罩主文案）"))
		hidden := true
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(20).Child(
				kit.Label(demoText("Company name", "公司名称")).Secondary(demoText("(optional)", "（可选）")).Render(cx),
				kit.Label(demoText("Hello world", "Hello 世界 Hello")).Highlights("Hello").Style(func(t *el.TextEl) { t.TextSize(24).Bold() }).Render(cx),
				kit.Label(demoText("Prefix matching; highlights the prefix only", "前缀匹配，只高亮前缀")).HighlightPrefix(demoText("Prefix", "前缀")).Render(cx),
				kit.Label(demoText("This text wraps to the container width. Highlighted matches retain their layout even across line breaks.", "这是一段可以自动换行的文案，匹配文字跨行仍保留完整排版。文案会按照容器宽度布局。")).Highlights(demoText("Text", "文案")).Render(cx),
				masked.Render(cx), kit.Button(demoText("Show / mask", "显示 / 遮罩"), func() { hidden = !hidden; masked.Masked(hidden) }).Render(cx),
			)
		}))
	})
}
