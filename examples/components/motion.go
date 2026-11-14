package main

import (
	"fmt"
	"image"
	"image/color"
	"slices"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
	"github.com/dyike/keel/ui/kit"
	"github.com/dyike/keel/ui/theme"
)

func init() {
	registerSection("motion", "motion", func() core.Widget {
		v := &motionGallery{items: []int{1, 2, 3}, next: 3, count: 19}
		v.heroImage = motionPhoto()
		v.dialog = kit.Dialog(demoText("Animated dialog", "进退场对话框")).Body(el.ViewFunc(func(*el.Context) el.Element {
			return el.Text(demoText("Close and reopen while the fade is running.", "淡入淡出过程中也可以关闭和重新打开。"))
		})).CloseButton(true)
		v.notifier = kit.Notifier()
		return el.Root(v)
	})
}

type motionGallery struct {
	right, wide, ordered  bool
	items                 []int
	next, clicks          int
	dragReset             uint64
	dragTarget            el.DragPoint
	count                 int
	countDown, swapText   bool
	heroOpen, swapContent bool
	heroImage             image.Image
	dialog                *kit.DialogView
	notifier              *kit.NotifierView
}

func (v *motionGallery) Render(cx *el.Context) el.Element {
	w, _ := cx.ViewportSize()
	pose := el.Pose{}
	if v.right {
		pose = el.Pose{X: min(200, max(0, w-220)), Y: 20, Rotate: 8, Scale: 1.08}
	}
	card := el.Div().ID("motion-card").Name(demoText("Moving card", "移动卡片")).W(el.Dp(140)).H(el.Dp(80)).Rounded(12).Bg(theme.Primary).P(14).AnimatePose(pose, el.Snappy).
		OnClick(func() { v.clicks++; v.right = !v.right }).Child(el.Text(demoText("Tap to reverse", "点击反向")).Bold().TextColor(theme.OnColor), el.Text(fmt.Sprintf("%s %d", demoText("Clicks", "点击"), v.clicks)).TextColor(theme.OnColor))
	cx.Overlay("motion-tip", el.Anchored("motion-card", el.Div().P(6).Rounded(6).Bg(theme.Surface).Border(1, theme.Border).Child(el.Text(demoText("I follow the card", "我会跟随卡片")).TextSize(12))).Placement(el.Bottom, el.Start))
	list := el.Div().Gap(8).Items(el.Start)
	items := el.PresentItems(cx, "motion-items", v.items, func(id int) any { return id }, el.PresenceListOptions{PresenceOptions: el.PresenceOptions{Enter: el.Pose{X: 24, Fade: 1}, Exit: el.Pose{X: -24, Fade: 1}, Animation: el.Snappy}, ExitBeforeEnter: v.ordered})
	width := float32(220)
	if v.wide {
		width = 340
	}
	ordered := demoText("off", "关闭")
	if v.ordered {
		ordered = demoText("on", "开启")
	}
	for _, item := range items {
		list.Child(el.Div().ID(fmt.Sprintf("row-%d", item.Value)).W(el.Dp(width)).MaxW(el.Full).H(el.Dp(48)).Rounded(8).Bg(theme.Subtle).P(12).Presence(item.Frame).AnimateLayout(el.Snappy).
			Child(el.Text(fmt.Sprintf("%s %d", demoText("Item", "条目"), item.Value))))
	}
	buttons := el.Div().Row().Wrap().Gap(8).Child(
		kit.Button(demoText("Replace first", "替换第一项"), func() {
			if len(v.items) > 0 {
				v.items = v.items[1:]
			}
			v.next++
			v.items = append(v.items, v.next)
		}).Render(cx),
		kit.Button(demoText("Reverse order", "倒序排列"), func() { slices.Reverse(v.items) }).Render(cx),
		kit.Button(demoText("Change width", "改变宽度"), func() { v.wide = !v.wide }).Render(cx),
		kit.Button(demoText("Toggle ordered entry", "切换先退后进"), func() { v.ordered = !v.ordered }).Render(cx),
	)
	dragLimit := min(float32(240), max(float32(0), w-220))
	dragCard := el.Div().ID("drag-card").Name(demoText("Draggable card", "拖拽卡片")).W(el.Dp(120)).H(el.Dp(70)).P(12).Rounded(12).Bg(theme.Primary).
		Draggable(el.DragOptions{
			Reset: v.dragReset, Axis: el.DragHorizontal, Bounds: &el.DragBounds{MaxX: dragLimit},
			Snap: []el.DragPoint{{}, {X: dragLimit}}, Elastic: .4, Momentum: true, Animation: el.Snappy,
			OnRelease: func(p el.DragPoint) { v.dragTarget = p },
		}).Child(el.Text(demoText("Drag and release", "拖拽后松手")).TextColor(theme.OnColor))
	title := demoText("Motion moves text", "中文与 English 👩🏽‍💻")
	if v.swapText {
		title = demoText("Text moves with motion", "English 与中文 👩🏽‍💻")
	}
	textControls := el.Div().Row().Wrap().Gap(8).Child(
		kit.Button("−", func() { v.count--; v.countDown = true }).Render(cx),
		kit.Button("+", func() { v.count++; v.countDown = false }).Render(cx),
		kit.Button(demoText("Rearrange text", "重排文字"), func() { v.swapText = !v.swapText }).Render(cx),
	)
	contentStage := el.Div().H(el.Dp(100)).Items(el.Start)
	for _, item := range el.PresentItems(cx, "content-swap", []bool{v.swapContent}, func(value bool) any { return value }, el.PresenceListOptions{
		PresenceOptions: el.PresenceOptions{Enter: el.Pose{Y: 8, Scale: .8, Fade: 1}, Exit: el.Pose{Y: -8, Scale: .8, Fade: 1}, Animation: el.Snappy},
	}) {
		var content el.Element = kit.Icon(kit.IconStar).Size(60).Color(theme.Primary).Render(cx)
		if item.Value {
			content = kit.Image(v.heroImage, demoText("Mountain landscape", "山景")).Size(80, 80).Render(cx)
		}
		contentStage.Child(el.Div().ID(fmt.Sprintf("content-%t", item.Value)).Absolute().Top(8).Left(12).Size(el.Dp(80)).Items(el.Center).Justify(el.Center).Presence(item.Frame).Child(content))
	}
	// Both locations have separate parents and element identities. Only the
	// application-level match identity connects their displayed geometry.
	heroOwner, heroSize, heroInset := "thumbnail", float32(80), float32(0)
	heroRadius, heroColor := float32(10), theme.Subtle
	heroTitle := demoText("Mountain", "山景")
	titleSize := float32(18)
	if v.heroOpen {
		heroOwner, heroSize, heroInset = "detail", 180, min(float32(180), max(float32(0), w-300))
		heroRadius, heroColor = 28, theme.Primary
		heroTitle = demoText("Mountain · a closer look", "山景 · 近看风景")
		titleSize = 24
	}
	hero := el.Div().ID("photo").Name(demoText("Shared photo", "共享图片")).Size(el.Dp(heroSize)).Rounded(heroRadius).Bg(heroColor).Border(3, heroColor).
		SharedLayout(motionPhotoIdentity{}, el.Snappy).Fly().OnClick(func() { v.heroOpen = !v.heroOpen }).
		Child(kit.Image(v.heroImage, demoText("Mountain landscape", "山景")).Size(heroSize, heroSize).Render(cx))
	gradient := theme.Gradient{From: theme.Primary, To: theme.Success, Angle: 20}
	elevation, titleWidth := theme.ElevationSm, float32(180)
	if v.heroOpen {
		gradient = theme.Gradient{From: theme.Info, To: theme.Primary, Angle: 140}
		elevation, titleWidth = theme.ElevationLg, 300
	}
	sharedTitle := el.Div().ID("title-"+heroOwner).W(el.Dp(titleWidth)).MaxW(el.Full).H(el.Dp(72)).P(12).Rounded(heroRadius).
		BgGradient(gradient).Shadow(elevation).SharedLayout(motionTitleIdentity{}, el.Snappy).MatchText().Fly().OnClick(func() { v.heroOpen = !v.heroOpen }).
		Child(el.Text(heroTitle).ID("shared-title").TextSize(titleSize).TextColor(theme.OnColor))
	titleStage := el.Div().H(el.Dp(110)).Items(el.Start).Child(el.Div().ID("title-location-"+heroOwner).Translate(heroInset, 0).Items(el.Start).Child(sharedTitle))
	heroStage := el.Div().H(el.Dp(210)).Items(el.Start).Child(el.Div().ID(heroOwner).W(el.Dp(heroSize)).H(el.Dp(200)).Translate(heroInset, 12).Items(el.Start).ScrollY().Child(hero))
	return el.Div().ScrollY().HFull().P(24).Gap(18).Items(el.Stretch).Child(
		el.Text(demoText("Motion", "动画")).TextSize(24).Bold(),
		el.Text(demoText("Move, interrupt, reorder and remove. Turn on reduced motion for immediate changes.", "移动、反向、重排和移除；开启减少动态效果时直接完成变化。")).TextColor(theme.Muted),
		kit.Button(demoText("Toggle reduced motion", "切换减少动态效果"), func() { theme.SetReducedMotion(!theme.ReducedMotion) }).Render(cx),
		kit.Button(demoText("Move / reverse", "移动 / 反向"), func() { v.right = !v.right }).Render(cx),
		el.Div().H(el.Dp(160)).P(12).Items(el.Start).Child(card),
		el.Text(demoText("Drag, bounds and snapping", "拖拽、边界与吸附")).Bold(),
		el.Div().H(el.Dp(110)).P(12).Items(el.Start).Child(dragCard),
		el.Text(fmt.Sprintf("%s: %.0f dp", demoText("Release target", "松手目标"), v.dragTarget.X)).TextColor(theme.Muted),
		kit.Button(demoText("Reset drag", "重置拖拽"), func() { v.dragReset++; v.dragTarget = el.DragPoint{} }).Render(cx),
		el.Text(demoText("Text and numbers", "文字与数字变化")).Bold(), textControls,
		el.Text(fmt.Sprintf("%04d", v.count)).ID("motion-count").TextSize(32).Bold().ContentTransition(el.ContentTransition{Numeric: true, Down: v.countDown, Animation: el.Snappy}),
		el.Text(title).ID("motion-title").TextSize(24).ContentTransition(el.ContentTransition{Morph: true, Animation: el.Snappy}),
		el.Text(demoText("Image and icon transitions", "图片与图标切换")).Bold(),
		kit.Button(demoText("Swap content", "切换内容"), func() { v.swapContent = !v.swapContent }).Render(cx), contentStage,
		el.Text(demoText("Shared element: thumbnail to detail", "共享元素：缩略图到详情")).Bold(),
		kit.Button(demoText("Open / return", "展开 / 返回"), func() { v.heroOpen = !v.heroOpen }).Render(cx),
		heroStage, titleStage,
		el.Text(demoText("Image, gradient, shadow and title transition together. Descendant text is matched by the container.", "图片、渐变、阴影与标题一同过渡；容器自动匹配内部文字，点击即可反向。")).TextColor(theme.Muted),
		el.Text(demoText("Entry, exit and layout", "进退场与布局变化")).Bold(), buttons,
		el.Text(fmt.Sprintf("%s: %s", demoText("Wait for exits", "先退后进"), ordered)).TextColor(theme.Muted), list,
		el.Div().Row().Wrap().Gap(8).Child(
			kit.Button(demoText("Open dialog", "打开对话框"), func() { v.dialog.SetValue(true) }).Render(cx),
			kit.Button(demoText("Show notification", "显示通知"), func() {
				v.notifier.Notify(kit.Notice{Title: demoText("Saved", "已保存"), Body: demoText("Close to see the exit transition.", "关闭后可以观察退场动画。")})
			}).Render(cx),
		), v.dialog.Render(cx), v.notifier.Render(cx),
	)
}

// A local raster fixture keeps the shared-image example independent of I/O.
type motionPhotoIdentity struct{}
type motionTitleIdentity struct{}

func motionPhoto() image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, 180, 180))
	for y := range 180 {
		for x := range 180 {
			c := color.NRGBA{R: uint8(70 + y/3), G: uint8(130 + y/4), B: 210, A: 255}
			if (x-133)*(x-133)+(y-40)*(y-40) < 17*17 {
				c = color.NRGBA{R: 255, G: 220, B: 134, A: 255}
			}
			if y > 65+absMotion(x-66) {
				c = color.NRGBA{R: 57, G: 88, B: 112, A: 255}
			}
			if y > 112+absMotion(x-138)/2 {
				c = color.NRGBA{R: 35, G: 62, B: 82, A: 255}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}
func absMotion(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
