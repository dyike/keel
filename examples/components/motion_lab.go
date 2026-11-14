package main

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/motion"
	"github.com/dyike/keel/ui/theme"
)

// The ui/motion sections: value animations, gestures, presence and layout,
// scroll linking, spring presets, page transitions, poses, timelines and
// cursor effects. Each section builds its own lab and shows one topic, so
// the state of one demonstration never animates another.
func init() {
	registerSection("motion_values", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.springSection(cx), d.easingSection(cx), d.keyframeSection(cx)}
		})
	})
	registerSection("motion_gestures", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.hoverSection(cx), d.dragSection(cx)}
		})
	})
	registerSection("motion_layout", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.presenceSection(cx), d.layoutSection(cx)}
		})
	})
	registerSection("motion_scroll", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.scrollSection(cx)}
		})
	})
	registerSection("motion_presets", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.presetsSection(cx)}
		})
	})
	registerSection("motion_pages", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.pagesSection(cx)}
		})
	})
	registerSection("motion_pose", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.poseSection(cx)}
		})
	})
	registerSection("motion_timeline", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.timelineSection(cx)}
		})
	})
	registerSection("motion_effects", "motion", func() core.Widget {
		return newMotionLab().root(func(d *motionLab, cx *el.Context) []el.Element {
			return []el.Element{d.effectsSection(cx)}
		})
	})
}

type motionLab struct {
	reduced                  *kit.SwitchView
	progress                 float32
	presetTarget             float32
	presets                  [3]motion.Value
	pages                    *el.RootWidget
	timeline                 motion.Timeline
	timelineOptions          motion.TimelineOptions
	timelineProgress         float32
	timelineScroll           bool
	poseRight                bool
	poseClicks               int
	right                    bool
	spring, tween            motion.Value
	keyframe                 motion.Value
	hover, tap               motion.Value
	presence                 motion.Presence[int]
	items                    []int
	next                     int
	order                    bool
	layout                   motion.LayoutGroup
	dragX, dragY             motion.Value
	dragTargetX, dragTargetY float32
	dragStartX, dragStartY   float32
	pressed                  bool
	drag                     motion.DragTracker
	minDrag, maxDrag         float32
	keyValues, keyTimes      []float32
	pointerGlow              motion.PointerGlow
	confetti                 motion.Confetti
	orders                   int
}

func newMotionLab() *motionLab {
	d := &motionLab{
		items: []int{1, 2, 3}, next: 4,
		minDrag: -120, maxDrag: 120,
		keyValues: []float32{0, 1.2, .9, 1}, keyTimes: []float32{0, .45, .75, 1},
	}
	d.reduced = kit.Switch(demoText("Reduced motion", "减少动态效果"), theme.ReducedMotion).OnChange(theme.SetReducedMotion)
	d.spring.Set(0)
	d.tween.Set(0)
	d.keyframe.Set(0)
	d.hover.Set(1)
	d.tap.Set(1)
	d.timeline.Track("position", motion.TweenSpec{To: 260, Duration: 900 * time.Millisecond, Ease: motion.EaseInOut}).
		Label("arrival", motion.WithPrevious(500*time.Millisecond)).
		Track("title", motion.TweenSpec{From: 1, To: 0, Duration: 400 * time.Millisecond}, motion.AtLabel("arrival", 0)).
		Track("turn", motion.TweenSpec{To: 12, Duration: 300 * time.Millisecond})
	d.dragX.Set(0)
	d.dragY.Set(0)
	return d
}

// root is a scrolling page with the reduced-motion switch above the topic.
// The top bar follows the page scroll (see motion_scroll).
func (d *motionLab) root(topic func(*motionLab, *el.Context) []el.Element) core.Widget {
	return el.Root(el.ViewFunc(func(cx *el.Context) el.Element {
		offset, viewport, content := cx.ScrollState("motion-page")
		d.progress = 0
		if content > viewport {
			d.progress = offset / (content - viewport)
		}
		d.reduced.SetValue(theme.ReducedMotion)
		body := el.Div().P(24).Gap(22).Items(el.Stretch).Child(
			el.Div().Row().Items(el.Center).Justify(el.SpaceBetween).Gap(12).Child(
				el.Text(demoText("Pure Go, frame-time driven, event-idle animation.", "纯 Go、按帧时间采样、静止时事件驱动的动效。")).TextColor(theme.Muted),
				d.reduced.Render(cx),
			),
		)
		body.Child(topic(d, cx)...)
		return el.Div().ID("motion-page").ScrollY().HFull().Child(
			el.Div().H(el.Dp(4)).W(el.Frac(d.progress)).Bg(theme.Primary),
			body,
		)
	}))
}

func motionCard(title, subtitle string, children ...el.Element) el.Element {
	return el.Div().P(18).Gap(12).Rounded(12).Bg(theme.Surface).Border(1, theme.Border).Child(
		el.Text(title).TextSize(18).Bold(),
		el.Text(subtitle).TextColor(theme.Muted).TextSize(13),
		el.Div().Gap(12).Child(children...),
	)
}

func motionBox(label string, fill color.NRGBA) *el.DivEl {
	return el.Div().Size(el.Dp(56)).Rounded(12).Bg(fill).Items(el.Center).Justify(el.Center).Child(el.Text(label).TextColor(theme.OnColor).Bold())
}

func (d *motionLab) springSection(cx *el.Context) el.Element {
	target := float32(0)
	if d.right {
		target = 240
	}
	sx := d.spring.To(cx, target, motion.Spring(180, 16))
	tx := d.tween.To(cx, target, motion.Tween(450*time.Millisecond, motion.EaseInOut))
	return motionCard(demoText("Spring vs tween", "弹簧与补间"), demoText("Retarget halfway: the spring carries velocity; the tween restarts from its current position.", "中途改目标：弹簧继承速度，补间从当前位置重新开始。"),
		kit.Button(demoText("Retarget", "切换目标"), func() { d.right = !d.right }).Render(cx),
		el.Div().H(el.Dp(64)).Child(motionBox("S", theme.Primary).Translate(sx, 0)),
		el.Div().H(el.Dp(64)).Child(motionBox("T", theme.Info).Translate(tx, 0)),
	)
}

func (d *motionLab) easingSection(*el.Context) el.Element {
	curves := []struct {
		name string
		ease motion.Easing
	}{{"linear", motion.Linear}, {"easeInOut", motion.EaseInOut}, {"backOut", motion.BackOut}, {"anticipate", motion.Anticipate}}
	rows := make([]el.Element, 0, len(curves))
	for _, curve := range curves {
		p := curve.ease(.6)
		rows = append(rows, el.Div().Row().Items(el.Center).Gap(10).Child(
			el.Text(curve.name).W(el.Dp(90)).Mono(),
			el.Div().Size(el.Dp(14)).Rounded(99).Bg(theme.Primary).Translate(p*220, 0),
		))
	}
	return motionCard(demoText("Easing", "缓动曲线"), demoText("Motion-style easing functions sampled at a progress of 0.6.", "在 0.6 进度处对比 Motion 风格缓动函数。"), rows...)
}

func (d *motionLab) keyframeSection(cx *el.Context) el.Element {
	p := d.keyframe.To(cx, 1, motion.Keyframes(900*time.Millisecond, d.keyValues, d.keyTimes, motion.EaseInOut))
	card := motionBox(fmt.Sprintf("%.2f", p), theme.Warning)
	motion.Transform(&card.Styled, motion.Affine{Scale: max(.1, p), Rotate: (p - 1) * float32(math.Pi)})
	return motionCard(demoText("Keyframes", "关键帧"), demoText("Scale and rotation follow [0, 1.2, 0.9, 1] with explicit times.", "缩放和旋转按 [0, 1.2, 0.9, 1] 及显式时间点运行。"),
		kit.Button(demoText("Replay", "重播"), func() { d.keyframe.Set(0) }).Render(cx),
		el.Div().H(el.Dp(90)).Items(el.Center).Child(card),
	)
}

func (d *motionLab) hoverSection(cx *el.Context) el.Element {
	hoverTarget := float32(1)
	if cx.Hovered("motion-hover") {
		hoverTarget = 1.06
	}
	tapTarget := float32(1)
	if d.pressed {
		tapTarget = .94
	}
	scale := d.hover.To(cx, hoverTarget, motion.Spring(300, 24)) * d.tap.To(cx, tapTarget, motion.Spring(400, 28))
	card := el.Div().ID("motion-hover").W(el.Dp(260)).H(el.Dp(90)).Rounded(14).Bg(theme.Primary).Shadow(theme.ElevationMd).
		Items(el.Center).Justify(el.Center).CursorPointer().OnDrag(func(event el.DragEvent) {
		d.pressed = event.Kind != el.DragEnd
	}).Child(el.Text(demoText("Hover and press me", "悬停并按下我")).TextColor(theme.OnColor).Bold())
	motion.Transform(&card.Styled, motion.Affine{Scale: scale})
	return motionCard(demoText("While hover / tap", "悬停 / 按下"), demoText("Pointer state drives interruptible springs; transformed hit testing stays with Gio operations.", "指针状态驱动可打断弹簧；变换后的命中测试随 Gio 操作生效。"), card)
}

func (d *motionLab) dragSection(cx *el.Context) el.Element {
	tr := motion.Inertia(500 * time.Millisecond)
	tr.Min, tr.Max = &d.minDrag, &d.maxDrag
	x := d.dragX.To(cx, d.dragTargetX, tr)
	y := d.dragY.To(cx, d.dragTargetY, tr)
	card := motionBox(demoText("Drag", "拖拽"), theme.Success).ID("motion-drag").CursorPointer().Translate(x, y).OnDrag(func(event el.DragEvent) {
		d.drag.Track(cx.Now(), event)
		switch event.Kind {
		case el.DragStart:
			d.dragStartX, d.dragStartY = event.X-x, event.Y-y
		case el.DragMove:
			d.dragTargetX, d.dragTargetY = event.X-d.dragStartX, event.Y-d.dragStartY
			d.dragX.Set(d.dragTargetX)
			d.dragY.Set(d.dragTargetY)
		case el.DragEnd:
			vx, vy := d.drag.Velocity()
			d.dragTargetX = min(max(d.dragTargetX+vx*.18, d.minDrag), d.maxDrag)
			d.dragTargetY = min(max(d.dragTargetY+vy*.18, -40), 40)
		}
	})
	return motionCard(demoText("Drag and inertia", "拖拽与惯性"), demoText("A fixed sample window estimates release velocity; inertia decays toward constrained targets.", "固定采样窗口估算松手速度；惯性向受约束目标衰减。"),
		el.Div().H(el.Dp(140)).Items(el.Center).Justify(el.Center).Child(card),
	)
}

func (d *motionLab) presenceSection(cx *el.Context) el.Element {
	items := d.presence.Items(cx, d.items, motion.Tween(220*time.Millisecond, motion.EaseOut), motion.Tween(180*time.Millisecond, motion.EaseIn))
	rows := make([]el.Element, 0, len(items)+1)
	rows = append(rows, el.Div().Row().Gap(8).Child(
		kit.Button(demoText("Add", "添加"), func() { d.items = append(d.items, d.next); d.next++ }).Render(cx),
		kit.Button(demoText("Remove last", "移除末项"), func() {
			if len(d.items) > 0 {
				d.items = d.items[:len(d.items)-1]
			}
		}).Variant(kit.ButtonSecondary).Render(cx),
	))
	for _, item := range items {
		rows = append(rows, el.Div().P(10).Rounded(8).Bg(theme.Subtle).Opacity(item.Progress).Reveal(item.Progress).Disabled(item.Disabled).
			Translate(0, (1-item.Progress)*8).Child(el.Text(fmt.Sprintf("%s %d", demoText("Item", "条目"), item.Key))))
	}
	return motionCard(demoText("Presence", "Presence 退出动画"), demoText("Removed keys remain in stable order until their exit animation settles.", "键被移除后仍按稳定顺序保留，直到退出动画静止。"), rows...)
}

func (d *motionLab) layoutSection(cx *el.Context) el.Element {
	children := make([]el.Element, 0, 4)
	colors := []color.NRGBA{theme.Primary, theme.Info, theme.Success, theme.Danger}
	order := [...]int{0, 1, 2, 3}
	if d.order {
		order = [...]int{3, 2, 1, 0}
	}
	for _, i := range order {
		children = append(children, d.layout.Animate(cx, fmt.Sprintf("tile-%d", i), motionBox(fmt.Sprint(i+1), colors[i]), motion.Spring(240, 22)))
	}
	return motionCard(demoText("Position FLIP", "位置 FLIP"), demoText("Reordering changes layout immediately while paint compensation animates each tile from its previous position.", "重排立即改变布局，绘制补偿让每个色块从原位置过渡。"),
		kit.Button(demoText("Reverse", "反转"), func() { d.order = !d.order }).Render(cx),
		el.Div().Row().Gap(12).H(el.Dp(68)).Child(children...),
	)
}

func (d *motionLab) scrollSection(*el.Context) el.Element {
	rows := []el.Element{
		el.Div().H(el.Dp(72)).Child(motionBox(fmt.Sprintf("%d%%", int(d.progress*100)), theme.Info).Translate(motion.MapRange(d.progress, 0, 1, 0, 240, true), 0)),
	}
	// Enough content to scroll the page.
	for i := range 12 {
		rows = append(rows, el.Div().H(el.Dp(64)).P(12).Rounded(8).Bg(theme.Subtle).Child(el.Text(fmt.Sprintf("%s %d", demoText("Scroll the page", "滚动页面"), i+1))))
	}
	return motionCard(demoText("Scroll linked", "滚动联动"), demoText("The top progress bar and this parallax marker are mapped from the page scroll state without requesting animation frames.", "顶部进度条和此视差标记直接映射页面滚动状态，不额外请求动画帧。"), rows...)
}

func (d *motionLab) presetsSection(cx *el.Context) el.Element {
	rows := []el.Element{kit.Button(demoText("Retarget presets", "切换弹簧目标"), func() { d.presetTarget = 1 - d.presetTarget }).Render(cx)}
	names := []string{demoText("Smooth · steady", "Smooth · 平稳"), demoText("Snappy · slight bounce", "Snappy · 轻微回弹"), demoText("Bouncy · visible bounce", "Bouncy · 明显回弹")}
	for i, tr := range []motion.Transition{motion.Smooth(), motion.Snappy(), motion.Bouncy()} {
		x := d.presets[i].To(cx, d.presetTarget, tr)
		rows = append(rows, el.Text(names[i]), el.Div().H(el.Dp(48)).Rounded(12).Bg(theme.Subtle).Child(
			el.Div().Absolute().Left(12+160*x).Top(8).Size(el.Dp(32)).Rounded(16).Bg(theme.Primary)))
	}
	return motionCard(demoText("Spring presets", "弹簧预设"), demoText("Compare three responses. Click again during motion to reverse direction.", "对比三种响应，运动中再次点击可改变方向。"), rows...)
}

func (d *motionLab) pagesSection(*el.Context) el.Element {
	if d.pages == nil {
		shell := kit.AppShell(demoText("Pages", "页面")).
			Page(kit.AppPage{ID: "draft", Title: demoText("Draft", "草稿"), Icon: kit.IconFile, Create: func() el.View {
				return kit.Input(demoText("Title", "标题")).Placeholder(demoText("Type, switch pages, then return", "输入后切页，再返回查看"))
			}}).
			Page(kit.AppPage{ID: "settings", Title: demoText("Settings", "设置"), Icon: kit.IconSettings, Create: func() el.View {
				return kit.Switch(demoText("Notifications", "通知"), true)
			}})
		d.pages = el.Root(shell)
	}
	return motionCard(demoText("Page transitions", "页面切换"), demoText("Switch destinations to preview the fade. Input is retained; Reduced motion skips the transition.", "点击导航查看切页淡入。输入会保留；开启减少动态效果可跳过过渡。"),
		el.Div().H(el.Dp(340)).Items(el.Stretch).Child(el.Widget(d.pages).Grow()))
}

func (d *motionLab) poseSection(cx *el.Context) el.Element {
	p := el.Pose{}
	if d.poseRight {
		p = el.Pose{X: 220, Y: 35, Rotate: 12, Scale: 1.15, Fade: .15}
	}
	card := el.Div().ID("posed-card").Name(demoText("Animated card", "动画卡片")).W(el.Dp(160)).H(el.Dp(90)).Rounded(16).Bg(theme.Primary).P(16).
		AnimatePose(p, el.Snappy).OnClick(func() { d.poseClicks++; d.poseRight = !d.poseRight }).Child(
		el.Text(demoText("Catch me", "点我反向")).TextColor(theme.OnColor).Bold(),
		el.Text(fmt.Sprintf("%s %d", demoText("Clicks", "点击"), d.poseClicks)).TextColor(theme.OnColor),
	)
	cx.Overlay("pose-tip", el.Anchored("posed-card", el.Div().P(6).Rounded(6).Bg(theme.Surface).Border(1, theme.Border).Child(el.Text(demoText("The anchor follows", "浮层跟随位置")).TextSize(12))).Placement(el.Bottom, el.Start))
	return motionCard(demoText("Paint-only element motion", "只重绘的元素动画"), demoText("Retarget while moving. The card, its hit region and its anchored overlay travel together. When the spring settles, it stops requesting frames.", "动画中可反向；卡片、点击区域和锚定浮层同步移动。弹簧静止后不再请求动画帧。"),
		kit.Button(demoText("Move / reverse", "移动 / 反向"), func() { d.poseRight = !d.poseRight }).Render(cx),
		el.Div().H(el.Dp(230)).P(24).Items(el.Start).Child(card),
	)
}

func (d *motionLab) timelineSection(cx *el.Context) el.Element {
	options := d.timelineOptions
	if d.timelineScroll {
		options.Seek, options.Progress = nil, nil
		options.Scroll = &motion.ScrollBinding{ID: "timeline-scroll"}
	}
	f := d.timeline.Sample(cx, options)
	controls := el.Div().Row().Wrap().Gap(8).Child(
		kit.Button(demoText("Scroll / clock", "滚动驱动 / 时间驱动"), func() { d.timelineScroll = !d.timelineScroll }).Render(cx),
		kit.Button(demoText("Replay", "重新播放"), func() { d.timelineOptions.Run++; d.timelineOptions.Paused = false; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(demoText("Pause / resume", "暂停 / 继续"), func() { d.timelineOptions.Paused = !d.timelineOptions.Paused; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(demoText("Reverse", "倒放"), func() { d.timelineOptions.Reverse = !d.timelineOptions.Reverse; d.timelineOptions.Progress = nil }).Render(cx),
		kit.Button(demoText("Seek 50%", "跳转到 50%"), func() { d.timelineProgress = .5; d.timelineOptions.Progress = &d.timelineProgress }).Render(cx),
	)
	card := el.Div().ID("timeline-card").W(el.Dp(180)).H(el.Dp(110)).Rounded(16).Bg(theme.Primary).P(16).
		AnimatePose(el.Pose{X: f.Value("position"), Rotate: f.Value("turn")}, el.Instant).Child(
		el.Text(demoText("One playhead", "同一个播放时钟")).TextColor(theme.OnColor).Bold().AnimatePose(el.Pose{Fade: f.Value("title")}, el.Instant),
	)
	scroll := el.Div().ID("timeline-scroll").H(el.Dp(140)).ScrollY().Gap(8).P(12).Bg(theme.Subtle).Rounded(10)
	for i := range 8 {
		scroll.Child(el.Div().H(el.Dp(55)).NoShrink().Child(el.Text(fmt.Sprintf("%s %d", demoText("Scroll to scrub", "滚动控制进度"), i+1))))
	}
	return motionCard(demoText("Timeline", "时间线"), demoText("The title appears before the card arrives; rotation follows. Pause, reverse or seek the whole composition.", "卡片到达前开始显示标题，随后旋转。整组动画可暂停、倒放或跳转。"),
		controls, el.Div().H(el.Dp(220)).P(24).Items(el.Start).Child(card),
		el.Text(fmt.Sprintf("%s: %.0f%%", demoText("Progress", "播放进度"), f.Progress*100)), scroll,
	)
}

var motionConfettiColors = []color.NRGBA{
	{R: 0x8b, G: 0x5c, B: 0xf6, A: 0xff},
	{R: 0x22, G: 0xd3, B: 0xee, A: 0xff},
	{R: 0xfb, G: 0xbf, B: 0x24, A: 0xff},
	{R: 0x34, G: 0xd3, B: 0x99, A: 0xff},
	{R: 0xfb, G: 0x71, B: 0x85, A: 0xff},
}

func (d *motionLab) effectsSection(cx *el.Context) el.Element {
	glowCard := el.Div().W(el.Dp(360)).H(el.Dp(112)).P(18).Rounded(16).Bg(theme.Primary).
		CursorPointer().Items(el.Center).Justify(el.Center).Child(
		el.Div().Items(el.Center).Gap(4).Child(
			el.Text(demoText("Move the pointer across this card", "在卡片上移动指针")).TextColor(theme.OnColor).Bold(),
			el.Text(demoText("The glow and ripple begin at the cursor", "光晕和波纹从指针位置开始")).TextColor(color.NRGBA{R: 0xee, G: 0xf2, B: 0xff, A: 0xdd}).TextSize(12),
		),
	)
	d.pointerGlow.Apply(cx, "motion-pointer-glow", glowCard, motion.PointerGlowStyle{
		Color: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xb0}, Radius: 92,
	})
	stage := el.Div().W(el.Dp(360)).H(el.Dp(150)).Rounded(16).Bg(theme.Subtle).Items(el.Center).Justify(el.Center)
	stage.Child(kit.Button(demoText("Place order", "提交订单"), func() {
		d.orders++
		d.confetti.Emit(cx.Now(), 180, 75, motion.ConfettiOptions{
			Count: 96, Seed: uint64(d.orders), Gravity: 520,
			SpeedMin: 160, SpeedMax: 320,
			LifetimeMin: 700 * time.Millisecond, LifetimeMax: 1200 * time.Millisecond,
			SizeMin: 3, SizeMax: 7, Colors: motionConfettiColors,
		})
	}).Render(cx))
	d.confetti.Apply(cx, stage)
	return motionCard(demoText("Cursor glow and order confetti", "指针光晕与订单彩纸"),
		demoText("Move over the blue card to set the glow and ripple origin. Click Place order to emit confetti from the button. Reduced motion suppresses both.", "在蓝色卡片上移动指针，光晕和波纹会从指针位置出现；点击“提交订单”，彩纸会从按钮中心喷出。减少动态效果时两者都不显示。"),
		el.Div().Row().Wrap().Gap(16).Child(glowCard, stage),
	)
}
