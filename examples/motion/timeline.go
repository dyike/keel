package main

import (
	"fmt"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/motion"
	"github.com/dyike/keel/ui/theme"
)

func (d *demo) timelinePage(cx *el.Context) el.Element {
	options := d.timelineOptions
	if d.timelineScroll {
		options.Seek, options.Progress = nil, nil
		options.Scroll = &motion.ScrollBinding{ID: "timeline-scroll"}
	}
	f := d.timeline.Sample(cx, options)
	controls := el.Div().Row().Wrap().Gap(8).Child(
		kit.Button(text("Scroll / clock", "滚动驱动 / 时间驱动"), func() { d.timelineScroll = !d.timelineScroll }).Render(cx),
		kit.Button(text("Replay", "重新播放"), func() { d.timelineOptions.Run++; d.timelineOptions.Paused = false; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(text("Pause / resume", "暂停 / 继续"), func() { d.timelineOptions.Paused = !d.timelineOptions.Paused; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(text("Reverse", "倒放"), func() { d.timelineOptions.Reverse = !d.timelineOptions.Reverse; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(text("Seek 50%", "跳转到 50%"), func() { d.timelineProgress = .5; d.timelineOptions.Progress = &d.timelineProgress }).Render(cx),
	)
	card := el.Div().ID("timeline-card").W(el.Dp(180)).H(el.Dp(110)).Rounded(16).Bg(theme.Primary).P(16).
		AnimatePose(el.Pose{X: f.Value("position"), Rotate: f.Value("turn")}, el.Instant).Child(
		el.Text(text("One playhead", "同一个播放时钟")).TextColor(theme.OnColor).Bold().AnimatePose(el.Pose{Fade: f.Value("title")}, el.Instant),
	)
	scroll := el.Div().ID("timeline-scroll").H(el.Dp(140)).ScrollY().Gap(8).P(12).Bg(theme.Subtle).Rounded(10)
	for i := range 8 {
		scroll.Child(el.Div().H(el.Dp(55)).NoShrink().Child(el.Text(fmt.Sprintf("%s %d", text("Scroll to scrub", "滚动控制进度"), i+1))))
	}
	return el.Div().P(24).Gap(20).Child(d.header(cx),
		el.Text(text("Timeline", "时间线")).TextSize(22).Bold(),
		el.Text(text("The title appears before the card arrives; rotation follows. Pause, reverse or seek the whole composition.", "卡片到达前开始显示标题，随后旋转。整组动画可暂停、倒放或跳转。")).TextColor(theme.Muted),
		controls, el.Div().H(el.Dp(220)).P(24).Items(el.Start).Child(card),
		el.Text(fmt.Sprintf("%s: %.0f%%", text("Progress", "播放进度"), f.Progress*100)), scroll,
	)
}
