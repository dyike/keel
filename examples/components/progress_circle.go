package main

import (
	"fmt"
	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("progress_circle", "controls", func() core.Widget {
		done := kit.ProgressCircle("导入 Import").Size(88)
		done.SetValue(.42)
		done.Child(el.ViewFunc(func(*el.Context) el.Element { return el.Text(fmt.Sprintf("%.0f%%", done.Value()*100)) }))
		empty := kit.ProgressCircle("未开始 Empty").Size(40)
		full := kit.ProgressCircle("已完成 Done").Size(40)
		full.SetValue(1)
		busy := kit.ProgressCircle("同步 Sync").Size(40)
		busy.SetIndeterminate(true)
		return el.Embed(el.ViewFunc(func(cx *el.Context) el.Element {
			return el.Div().P(theme.SpaceXl).Gap(theme.SpaceLg).W(el.Dp(360)).MaxW(el.Full).Child(
				el.Div().Row().Wrap().Items(el.Center).Gap(theme.SpaceLg).Child(done.Render(cx), empty.Render(cx), full.Render(cx), busy.Render(cx)),
				el.Div().Row().Wrap().Gap(theme.SpaceSm).Child(
					kit.Button("+10%", func() { done.SetValue(done.Value() + .1) }).Render(cx),
					kit.Button("重置 Reset", func() { done.SetValue(0) }).Render(cx)))
		}))
	})
}
