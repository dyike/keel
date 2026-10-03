package main

import (
	"fmt"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
)

func init() {
	registerSection("notifier", "overlays", func() core.Widget {
		n := kit.Notifier()
		n.Notify(kit.Notice{Title: "默认位置", Body: "这条通知跟随容器的位置设置。", Timeout: -1})
		var richID int
		richID = n.Notify(kit.Notice{
			Title: "富内容与操作", Placement: kit.NoticeBottomLeft, Timeout: -1,
			Content: el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Gap(8).Child(
					el.Text("下载完成，可以打开文件。"),
					kit.Tag("report.pdf").Render(cx),
				)
			}),
			Action: kit.Button("完成并关闭", func() { n.Dismiss(richID) }).Variant(kit.ButtonSecondary),
		})
		return el.Root(&notifierGallery{n: n})
	})
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
		el.Text("Notifier：默认右上角，每个位置最多显示 5 条").Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button("成功", notify(kit.ToneSuccess, "保存成功")).Render(cx),
			kit.Button("警告", notify(kit.ToneWarning, "网络较慢")).Variant(kit.ButtonSecondary).Render(cx),
			kit.Button("错误", notify(kit.ToneDanger, "同步失败")).Variant(kit.ButtonDanger).Render(cx),
			kit.Button("常驻", func() { g.n.Notify(kit.Notice{Title: "需要处理", Body: "不会自动消失。", Timeout: -1}) }).Variant(kit.ButtonGhost).Render(cx),
		),
		el.Div().Row().Wrap().Gap(8).Child(
			kit.Button("默认右上角", func() { g.n.Placement(kit.NoticeTopRight) }).Render(cx),
			kit.Button("默认底部居中", func() { g.n.Placement(kit.NoticeBottomCenter) }).Render(cx),
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
