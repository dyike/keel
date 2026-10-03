package main

import (
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"time"
)

func init() {
	registerSection("shimmer_text", "feedback", func() core.Widget {
		normal := kit.ShimmerText("正在生成内容，文字仍然可以阅读……").Size(20)
		reverse := kit.ShimmerText("反向扫光 · Reverse").Size(24).Style(kit.ShimmerStyle{Duration: 3 * time.Second, Spread: .4, Reverse: true})
		once := kit.ShimmerText("只播放一次 · Once").Size(20).Once(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(24).Gap(20).Child(normal.Render(cx), reverse.Render(cx), once.Render(cx), kit.Button("重新播放", once.Restart).Render(cx))
		}))
	})
}
