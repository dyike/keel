package main

import (
	"fmt"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("notifier", "overlays", func() core.Widget { return el.Root(&notifierGallery{n: kit.Notifier()}) })
}

type notifierGallery struct {
	n     *kit.NotifierView
	count int
}

func (g *notifierGallery) Render(cx *el.Context) el.Element {
	notify := func(tone kit.Tone, title string) func() {
		return func() {
			g.count++
			g.n.Notify(kit.Notice{Title: fmt.Sprintf("%s #%d", title, g.count), Body: "悬停可以延长显示时间 Hover to keep.", Tone: tone})
		}
	}
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text("Notifier：右上角堆叠，5 秒后消失，最多显示 5 条").Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button("成功", notify(kit.ToneSuccess, "保存成功")).Render(cx),
			kit.Button("警告", notify(kit.ToneWarning, "网络较慢")).Variant(kit.ButtonSecondary).Render(cx),
			kit.Button("错误", notify(kit.ToneDanger, "同步失败")).Variant(kit.ButtonDanger).Render(cx),
			kit.Button("常驻", func() { g.n.Notify(kit.Notice{Title: "需要处理", Body: "不会自动消失。", Timeout: -1}) }).Variant(kit.ButtonGhost).Render(cx),
		),
		g.n.Render(cx),
	)
}
