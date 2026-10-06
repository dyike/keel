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
		g := &notifierGallery{n: n}
		n.Notify(kit.Notice{Title: demoText("Default placement", "默认位置"), Body: demoText("This notification follows the container's placement.", "这条通知跟随容器的位置设置。"), Timeout: -1})
		var richID int
		richID = n.Notify(kit.Notice{
			OnClick: func() { g.status = demoText("Notification click callback ran", "通知点击回调已执行") },
			OnClose: func() { g.status = demoText("Notification close callback ran", "通知关闭回调已执行") },
			Title:   demoText("Rich content and actions", "富内容与操作"), Placement: kit.NoticeBottomLeft, Timeout: -1,
			Content: el.ViewFunc(func(cx *el.Context) el.Element {
				return el.Div().Gap(8).Child(
					el.Text(demoText("Download complete. You can open the file.", "下载完成，可以打开文件。")),
					kit.Tag("report.pdf").Render(cx),
				)
			}),
			Action: kit.Button(demoText("Finish and close", "完成并关闭"), func() { n.Dismiss(richID) }).Variant(kit.ButtonSecondary),
		})
		return el.Root(g)
	})
}

type notifierGallery struct {
	n      *kit.NotifierView
	count  int
	status string
}

func (g *notifierGallery) Render(cx *el.Context) el.Element {
	notify := func(tone kit.Tone, title string) func() {
		return func() {
			g.count++
			g.n.Notify(kit.Notice{Title: fmt.Sprintf("%s #%d", title, g.count), Body: demoText("Hover pauses the countdown; moving away resumes the remaining time.", "悬停暂停倒计时，移开后继续剩余时间。"), Tone: tone})
		}
	}
	return el.Div().P(24).Gap(12).Items(el.Start).Child(
		el.Text(demoText("Notifier: top right by default; each position displays up to 5 notifications.", "Notifier：默认右上角，每个位置最多显示 5 条")).Bold(),
		el.Div().Row().Gap(8).Child(
			kit.Button(demoText("Success", "成功"), notify(kit.ToneSuccess, demoText("Saved successfully", "保存成功"))).Render(cx),
			kit.Button(demoText("Warning", "警告"), notify(kit.ToneWarning, demoText("Slow network", "网络较慢"))).Variant(kit.ButtonSecondary).Render(cx),
			kit.Button(demoText("Error", "错误"), notify(kit.ToneDanger, demoText("Sync failed", "同步失败"))).Variant(kit.ButtonDanger).Render(cx),
			kit.Button(demoText("Persistent", "常驻"), func() {
				g.n.Notify(kit.Notice{Title: demoText("Needs attention", "需要处理"), Body: demoText("Does not dismiss automatically.", "不会自动消失。"), Timeout: -1})
			}).Variant(kit.ButtonGhost).Render(cx),
		),
		el.Div().Row().Wrap().Gap(8).Child(
			kit.Button(demoText("Top right by default", "默认右上角"), func() { g.n.Placement(kit.NoticeTopRight) }).Render(cx),
			kit.Button(demoText("Bottom center by default", "默认底部居中"), func() { g.n.Placement(kit.NoticeBottomCenter) }).Render(cx),
			kit.Button(demoText("Update the same task", "更新同一任务"), func() {
				g.count++
				g.n.NotifyKey("download", kit.Notice{Title: demoText("Download task", "下载任务"), Body: fmt.Sprintf(demoText("Update %d; only one notification is retained.", "第 %d 次更新，始终只保留一条通知。"), g.count), Timeout: -1})
			}).Render(cx),
			kit.Button(demoText("Clear all notifications", "清除全部通知"), func() { g.n.Clear() }).Variant(kit.ButtonGhost).Render(cx),
		),
		kit.Button(demoText("Simulate background update", "模拟后台更新"), func() {
			id := g.n.Notify(kit.Notice{Title: demoText("Syncing", "正在同步"), Timeout: -1})
			go func() {
				time.Sleep(time.Second)
				core.Update(func() {
					g.n.Update(id, kit.Notice{Title: demoText("Sync complete", "同步完成"), Body: demoText("Updates in place and closes after 5 seconds.", "原位更新，5 秒后关闭。"), Tone: kit.ToneSuccess})
				})
			}()
		}).Render(cx),
		el.Text(g.status),
		g.n.Render(cx),
	)
}
