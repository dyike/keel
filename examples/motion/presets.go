package main

import (
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/motion"
	"github.com/dyike/keel/ui/theme"
)

func (d *demo) presetsSection(cx *el.Context) el.Element {
	rows := []el.Element{kit.Button(text("Retarget presets", "切换弹簧目标"), func() { d.presetTarget = 1 - d.presetTarget }).Render(cx)}
	names := []string{text("Smooth · steady", "Smooth · 平稳"), text("Snappy · slight bounce", "Snappy · 轻微回弹"), text("Bouncy · visible bounce", "Bouncy · 明显回弹")}
	for i, tr := range []motion.Transition{motion.Smooth(), motion.Snappy(), motion.Bouncy()} {
		x := d.presets[i].To(cx, d.presetTarget, tr)
		rows = append(rows, el.Text(names[i]), el.Div().H(el.Dp(48)).Rounded(12).Bg(theme.Subtle).Child(
			el.Div().Absolute().Left(12+160*x).Top(8).Size(el.Dp(32)).Rounded(16).Bg(theme.Primary)))
	}
	return section(text("Spring presets", "弹簧预设"), text("Compare three responses. Click again during motion to reverse direction.", "对比三种响应，运动中再次点击可改变方向。"), rows...)
}

func newPageDemo() *el.RootWidget {
	shell := kit.AppShell(text("Pages", "页面")).
		Page(kit.AppPage{ID: "draft", Title: text("Draft", "草稿"), Icon: kit.IconFile, Create: func() el.View {
			return kit.Input(text("Title", "标题")).Placeholder(text("Type, switch pages, then return", "输入后切页，再返回查看"))
		}}).
		Page(kit.AppPage{ID: "settings", Title: text("Settings", "设置"), Icon: kit.IconSettings, Create: func() el.View { return kit.Switch(text("Notifications", "通知"), true) }})
	return el.Root(shell)
}

func (d *demo) pagesSection(cx *el.Context) el.Element {
	return section(text("Page transitions", "页面切换"), text("Switch destinations to preview the fade. Input is retained; Reduced motion skips the transition.", "点击导航查看切页淡入。输入会保留；开启减少动态效果可跳过过渡。"), el.Div().H(el.Dp(340)).Items(el.Stretch).Child(el.Widget(d.pages).Grow()))
}
