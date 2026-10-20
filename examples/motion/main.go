package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/motion"
	"github.com/dyike/keel/ui/theme"
	"github.com/dyike/keel/ui/window"
)

type demo struct {
	onlyEffects              bool
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

func newDemo() *demo {
	d := &demo{
		items: []int{1, 2, 3}, next: 4,
		minDrag: -120, maxDrag: 120,
		keyValues: []float32{0, 1.2, .9, 1}, keyTimes: []float32{0, .45, .75, 1},
	}
	d.spring.Set(0)
	d.tween.Set(0)
	d.keyframe.Set(0)
	d.hover.Set(1)
	d.tap.Set(1)
	d.dragX.Set(0)
	d.dragY.Set(0)
	return d
}

func (d *demo) Render(cx *el.Context) el.Element {
	offset, viewport, content := cx.ScrollState("motion-page")
	progress := float32(0)
	if content > viewport {
		progress = offset / (content - viewport)
	}
	body := el.Div().P(24).Gap(22).Child(
		d.header(cx),
		d.effectsSection(cx),
	)
	if !d.onlyEffects {
		body.Child(
			d.springSection(cx),
			d.easingSection(cx),
			d.keyframeSection(cx),
			d.hoverSection(cx),
			d.presenceSection(cx),
			d.layoutSection(cx),
			d.dragSection(cx),
			d.scrollSection(cx, progress),
		)
	}
	page := el.Div().ID("motion-page").ScrollY().HFull().Child(
		el.Div().H(el.Dp(4)).W(el.Frac(progress)).Bg(theme.Primary),
		body,
	)
	return page
}

func (d *demo) header(cx *el.Context) el.Element {
	reduced := kit.Switch(text("Reduced motion", "减少动态效果"), theme.ReducedMotion).OnChange(theme.SetReducedMotion)
	return el.Div().Row().Items(el.Center).Justify(el.SpaceBetween).Child(
		el.Div().Gap(4).Child(
			el.Text("Keel Motion Lab").TextSize(26).Bold(),
			el.Text(text("Pure Go, frame-time driven, event-idle animation experiments.", "纯 Go、按帧时间采样、静止时事件驱动的动效实验。")).TextColor(theme.Muted),
		),
		reduced.Render(cx),
	)
}

func section(title, subtitle string, children ...el.Element) el.Element {
	return el.Div().P(18).Gap(12).Rounded(12).Bg(theme.Surface).Border(1, theme.Border).Child(
		el.Text(title).TextSize(18).Bold(),
		el.Text(subtitle).TextColor(theme.Muted).TextSize(13),
		el.Div().Gap(12).Child(children...),
	)
}

func box(label string, fill color.NRGBA) *el.DivEl {
	return el.Div().Size(el.Dp(56)).Rounded(12).Bg(fill).Items(el.Center).Justify(el.Center).Child(el.Text(label).TextColor(theme.OnColor).Bold())
}

func (d *demo) springSection(cx *el.Context) el.Element {
	target := float32(0)
	if d.right {
		target = 240
	}
	sx := d.spring.To(cx, target, motion.Spring(180, 16))
	tx := d.tween.To(cx, target, motion.Tween(450*time.Millisecond, motion.EaseInOut))
	return section(text("1 · Spring vs Tween", "1 · 弹簧与补间"), text("Retarget halfway: the spring carries velocity; the tween restarts from its current position.", "中途改目标：弹簧继承速度，补间从当前位置重新开始。"),
		kit.Button(text("Retarget", "切换目标"), func() { d.right = !d.right }).Render(cx),
		el.Div().H(el.Dp(64)).Child(box("S", theme.Primary).Translate(sx, 0)),
		el.Div().H(el.Dp(64)).Child(box("T", theme.Info).Translate(tx, 0)),
	)
}

func (d *demo) easingSection(cx *el.Context) el.Element {
	curves := []struct {
		name string
		ease motion.Easing
	}{{"linear", motion.Linear}, {"easeInOut", motion.EaseInOut}, {"backOut", motion.BackOut}, {"anticipate", motion.Anticipate}}
	rows := make([]el.Element, 0, len(curves))
	for i, curve := range curves {
		p := curve.ease(.6)
		rows = append(rows, el.Div().Row().Items(el.Center).Gap(10).Child(
			el.Text(curve.name).W(el.Dp(90)).Mono(),
			el.Div().Size(el.Dp(14)).Rounded(99).Bg(theme.Primary).Translate(p*220, 0),
		))
		_ = i
	}
	return section(text("2 · Easing", "2 · 缓动曲线"), text("The same 60% time sample mapped by Motion-style easing functions.", "同一个 60% 时间点经过 Motion 风格缓动函数映射。"), rows...)
}

func (d *demo) keyframeSection(cx *el.Context) el.Element {
	p := d.keyframe.To(cx, 1, motion.Keyframes(900*time.Millisecond, d.keyValues, d.keyTimes, motion.EaseInOut))
	card := box(fmt.Sprintf("%.2f", p), theme.Warning)
	motion.Transform(&card.Styled, motion.Affine{Scale: max(.1, p), Rotate: (p - 1) * float32(math.Pi)})
	return section(text("3 · Keyframes", "3 · 关键帧"), text("Scale and rotation follow [0, 1.2, 0.9, 1] with explicit times.", "缩放和旋转按 [0, 1.2, 0.9, 1] 及显式时间点运行。"),
		kit.Button(text("Replay", "重播"), func() { d.keyframe.Set(0) }).Render(cx),
		el.Div().H(el.Dp(90)).Items(el.Center).Child(card),
	)
}

func (d *demo) hoverSection(cx *el.Context) el.Element {
	hovered := cx.Hovered("motion-hover")
	hoverTarget := float32(1)
	if hovered {
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
	}).Child(el.Text(text("Hover and press me", "悬停并按下我")).TextColor(theme.OnColor).Bold())
	motion.Transform(&card.Styled, motion.Affine{Scale: scale})
	return section(text("4 · While hover / tap", "4 · 悬停 / 按下"), text("Pointer state drives interruptible springs; transformed hit testing stays with Gio operations.", "指针状态驱动可打断弹簧；变换后的命中测试随 Gio 操作生效。"), card)
}

func (d *demo) presenceSection(cx *el.Context) el.Element {
	items := d.presence.Items(cx, d.items, motion.Tween(220*time.Millisecond, motion.EaseOut), motion.Tween(180*time.Millisecond, motion.EaseIn))
	rows := make([]el.Element, 0, len(items)+1)
	rows = append(rows, el.Div().Row().Gap(8).Child(
		kit.Button(text("Add", "添加"), func() { d.items = append(d.items, d.next); d.next++ }).Render(cx),
		kit.Button(text("Remove last", "移除末项"), func() {
			if len(d.items) > 0 {
				d.items = d.items[:len(d.items)-1]
			}
		}).Variant(kit.ButtonSecondary).Render(cx),
	))
	for _, item := range items {
		row := el.Div().P(10).Rounded(8).Bg(theme.Subtle).Opacity(item.Progress).Reveal(item.Progress).Disabled(item.Disabled).
			Translate(0, (1-item.Progress)*8).Child(el.Text(fmt.Sprintf("Item %d", item.Key)))
		rows = append(rows, row)
	}
	return section(text("5 · Presence", "5 · Presence 退出动画"), text("Removed keys remain in stable order until their exit animation settles.", "键被移除后仍按稳定顺序保留，直到退出动画静止。"), rows...)
}

func (d *demo) layoutSection(cx *el.Context) el.Element {
	children := make([]el.Element, 0, 4)
	colors := []color.NRGBA{theme.Primary, theme.Info, theme.Success, theme.Danger}
	order := [...]int{0, 1, 2, 3}
	if d.order {
		order = [...]int{3, 2, 1, 0}
	}
	for _, i := range order {
		tile := box(fmt.Sprint(i+1), colors[i])
		children = append(children, d.layout.Animate(cx, fmt.Sprintf("tile-%d", i), tile, motion.Spring(240, 22)))
	}
	return section(text("6 · Position FLIP", "6 · 位置 FLIP"), text("Reordering changes layout immediately while paint compensation animates each tile from its previous position.", "重排立即改变布局，绘制补偿让每个色块从原位置过渡。"),
		kit.Button(text("Reverse", "反转"), func() { d.order = !d.order }).Render(cx),
		el.Div().Row().Gap(12).H(el.Dp(68)).Child(children...),
	)
}

func (d *demo) dragSection(cx *el.Context) el.Element {
	tr := motion.Inertia(500 * time.Millisecond)
	tr.Min, tr.Max = &d.minDrag, &d.maxDrag
	x := d.dragX.To(cx, d.dragTargetX, tr)
	y := d.dragY.To(cx, d.dragTargetY, tr)
	card := box(text("Drag", "拖拽"), theme.Success).ID("motion-drag").CursorPointer().Translate(x, y).OnDrag(func(event el.DragEvent) {
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
	return section(text("7 · Drag and inertia", "7 · 拖拽与惯性"), text("A fixed sample window estimates release velocity; inertia decays toward constrained targets.", "固定采样窗口估算松手速度；惯性向受约束目标衰减。"),
		el.Div().H(el.Dp(140)).Items(el.Center).Justify(el.Center).Child(card),
	)
}

func (d *demo) scrollSection(_ *el.Context, progress float32) el.Element {
	return section(text("8 · Scroll linked", "8 · 滚动联动"), text("The top progress bar and this parallax marker are mapped from the page scroll state without requesting animation frames.", "顶部进度条和此视差标记直接映射页面滚动状态，不额外请求动画帧。"),
		el.Div().H(el.Dp(72)).Child(box(fmt.Sprintf("%d%%", int(progress*100)), theme.Info).Translate(motion.MapRange(progress, 0, 1, 0, 240, true), 0)),
	)
}

var demoConfettiColors = []color.NRGBA{
	{R: 0x8b, G: 0x5c, B: 0xf6, A: 0xff},
	{R: 0x22, G: 0xd3, B: 0xee, A: 0xff},
	{R: 0xfb, G: 0xbf, B: 0x24, A: 0xff},
	{R: 0x34, G: 0xd3, B: 0x99, A: 0xff},
	{R: 0xfb, G: 0x71, B: 0x85, A: 0xff},
}

func (d *demo) effectsSection(cx *el.Context) el.Element {
	glowCard := el.Div().W(el.Dp(360)).H(el.Dp(112)).P(18).Rounded(16).Bg(theme.Primary).
		CursorPointer().Items(el.Center).Justify(el.Center).Child(
		el.Div().Items(el.Center).Gap(4).Child(
			el.Text(text("Move the pointer across this card", "在卡片上移动指针")).TextColor(theme.OnColor).Bold(),
			el.Text(text("The glow and ripple begin at the cursor", "光晕和波纹从指针位置开始")).TextColor(color.NRGBA{R: 0xee, G: 0xf2, B: 0xff, A: 0xdd}).TextSize(12),
		),
	)
	d.pointerGlow.Apply(cx, "motion-pointer-glow", glowCard, motion.PointerGlowStyle{
		Color: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xb0}, Radius: 92,
	})

	stage := el.Div().W(el.Dp(360)).H(el.Dp(150)).Rounded(16).Bg(theme.Subtle).Items(el.Center).Justify(el.Center)
	order := kit.Button(text("Place order", "提交订单"), func() {
		d.orders++
		d.confetti.Emit(cx.Now(), 180, 75, motion.ConfettiOptions{
			Count: 96, Seed: uint64(d.orders), Gravity: 520,
			SpeedMin: 160, SpeedMax: 320,
			LifetimeMin: 700 * time.Millisecond, LifetimeMax: 1200 * time.Millisecond,
			SizeMin: 3, SizeMax: 7, Colors: demoConfettiColors,
		})
	}).Render(cx)
	stage.Child(order)
	d.confetti.Apply(cx, stage)

	return section(text("Featured · Cursor glow and order confetti", "重点实验 · 指针光晕与订单彩纸"),
		text("Move over the blue card to set the glow and ripple origin. Click Place order to emit confetti from the button.", "在蓝色卡片上移动指针，光晕和波纹会从指针位置出现；点击“提交订单”，彩纸会从按钮中心喷出。"),
		el.Div().Row().Wrap().Gap(16).Child(glowCard, stage),
	)
}

func main() {
	screenshot := flag.String("screenshot", "", "render a PNG to this path and exit")
	language := flag.String("lang", "en", "example language: en or zh-CN")
	section := flag.String("section", "", "section to show: effects")
	flag.Parse()
	applyLanguage(*language)
	d := newDemo()
	switch *section {
	case "":
	case "effects":
		d.onlyEffects = true
	default:
		log.Fatalf("unknown section %q", *section)
	}
	content := el.Root(d)
	if *screenshot != "" {
		// The focused effects page is idle until pointer input, so it can be
		// captured with motion enabled. The complete lab starts value
		// animations immediately and uses reduced motion for deterministic PNGs.
		theme.SetReducedMotion(!d.onlyEffects)
		if err := window.Screenshot(content, 900, 720, *screenshot); err != nil {
			log.Fatal(err)
		}
		return
	}
	window.Open(window.Options{Title: "Keel Motion Lab", Width: 900, Height: 720, Content: content})
	window.Main()
}
