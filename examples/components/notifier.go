package main

import (
	"fmt"
	"time"

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
			g.n.Notify(kit.Notice{Title: fmt.Sprintf("%s #%d", title, g.count), Body: "悬停暂停倒计时，移开后继续剩余时间。", Tone: tone})
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
		kit.Button("模拟后台更新", func() {
			id := g.n.Notify(kit.Notice{Title: "正在同步", Timeout: -1})
			go func() {
				time.Sleep(time.Second)
				core.Update(func() {
					g.n.Update(id, kit.Notice{Title: "同步完成", Body: "原位更新，5 秒后关闭。", Tone: kit.ToneSuccess})
				})
			}()
		}).Render(cx),
		g.n.Render(cx),
	)
}
