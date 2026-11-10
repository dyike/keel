package main

import (
	"fmt"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func (d *demo) posePage(cx *el.Context) el.Element {
	p := el.Pose{}
	if d.poseRight {
		p = el.Pose{X: 220, Y: 35, Rotate: 12, Scale: 1.15, Fade: .15}
	}
	card := el.Div().ID("posed-card").Name(text("Animated card", "动画卡片")).W(el.Dp(160)).H(el.Dp(90)).Rounded(16).Bg(theme.Primary).P(16).
		AnimatePose(p, el.Snappy).OnClick(func() { d.poseClicks++; d.poseRight = !d.poseRight }).Child(
		el.Text(text("Catch me", "点我反向")).TextColor(theme.OnColor).Bold(),
		el.Text(fmt.Sprintf("%s %d", text("Clicks", "点击"), d.poseClicks)).TextColor(theme.OnColor),
	)
	cx.Overlay("pose-tip", el.Anchored("posed-card", el.Div().P(6).Rounded(6).Bg(theme.Surface).Border(1, theme.Border).Child(el.Text(text("The anchor follows", "浮层跟随位置")).TextSize(12))).Placement(el.Bottom, el.Start))
	return el.Div().P(24).Gap(20).Child(d.header(cx),
		el.Text(text("Paint-only element motion", "只重绘的元素动画")).TextSize(22).Bold(),
		el.Text(text("Retarget while moving. The card, its hit region and its anchored overlay travel together.", "动画中可反向；卡片、点击区域和锚定浮层同步移动。")).TextColor(theme.Muted),
		kit.Button(text("Move / reverse", "移动 / 反向"), func() { d.poseRight = !d.poseRight }).Render(cx),
		el.Div().H(el.Dp(230)).P(24).Items(el.Start).Child(card),
		el.Text(text("When the spring settles, it stops requesting frames.", "弹簧静止后不再请求动画帧。")).TextColor(theme.Muted),
	)
}
