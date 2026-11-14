# motion

实验性模块：API 可能变化。

`motion` 提供参照 Motion 的帧时间动效基础能力：物理弹簧、补间与三次贝塞尔
缓动、关键帧与重复模式、惯性、错峰、Presence、视觉变换、拖拽速度跟踪、颜色
混合、区间映射、指针原点光晕/波纹和彩纸粒子。它依赖 `ui/el`、`ui/core`、
`ui/theme` 和 Gio，不依赖
`ui/kit`、`ui/window`。

请把 `Value` 和 Transition 使用的切片保存在 view 状态中；动画可见的每次渲染都
调用 `Value.To`；不要用无限动画表达静态状态。采样使用 `el.Context.Now`，只有尚未
静止的值才调用 `Context.Animating`，所以静止应用仍由事件驱动。减少动态效果开启时
直接跳到目标值。

`Transform` 只改变绘制：布局和自动化边界不变，而且会替换元素已有的
`Decorate` 回调。`LayoutGroup` 为 `Div` 提供仅位置的 FLIP；它同样独占
`Decorate`，只能留在同一滚动坐标系内，不修正尺寸变化或变换后的自动化几何。

## 指针效果

把 `PointerGlow` 保存在 view 中，每次渲染都应用到一个 `Div`：

```go
card := el.Div().Size(el.Dp(160))
v.glow.Apply(cx, "glow-card", card, motion.PointerGlowStyle{Color: color.NRGBA{R: 255, G: 255, B: 255, A: 120}})
```

稳定 ID 交给 `Context.Hovered` 管理悬停状态，元素绘制坐标系内的 pass-through
指针处理器只提供原点；HiDPI 和滚动容器内，`Origin` 仍返回元素局部 dp。指针事件
直接驱动位置，只有淡出和波纹存活时才调用 `Animating`。减少动态效果会关闭光晕
和波纹。`Apply` 独占 `Decorate`。

`Confetti` 使用不设数量上限的复用粒子切片。发射时可能扩容；容量足够后，粒子更新
与绘制不分配内存。相同 seed 产生相同粒子：

```go
v.confetti.Emit(cx.Now(), 180, 60, motion.ConfettiOptions{Count: 96, Seed: 7})
v.confetti.Apply(cx, stage)
```

坐标单位是元素局部 dp，位置按经过时间计算，不按帧数累加。最后一个粒子到期后，
活动切片归零且不再请求帧；减少动态效果会直接清空 burst。`Apply` 独占
`Decorate`；元素还需要其他装饰时，请增加一层包装元素。

## 弹簧预设

`Smooth()`、`Snappy()`、`Bouncy()` 返回独立的过渡配置，采用 500ms 时长与 0 / 0.15 / 0.3 回弹参数。它们使用 Keel 现有求解器，重定向时保留速度。时长描述感知响应，不是精确停止时间。运行 `go run ./examples/components -section motion_presets` 可交互对比。

数值采样器通过 `ui/internal/animation` 与 `ui/el` 共用，只依赖标准库。元素变换可使用 `element.AnimatePose(el.Pose{...}, transition)`，让合适的动画帧只重绘已有布局。原有 `Value.To` 仍在 Render 中采样；实验性 `Transform`、`LayoutGroup` 的上述限制仍然适用。


## 时间线

在视图状态中构建一次 `Timeline`。所有轨道共用一个播放时钟，通过标签和相对位置
编排顺序、重叠和错峰，无须给每个动画设置定时器：

```go
v.timeline.Track("x", motion.TweenSpec{To: 200, Duration: time.Second}).
    Label("arrival", motion.WithPrevious(600*time.Millisecond)).
    Track("fade", motion.TweenSpec{From: 1, To: 0, Duration: 400*time.Millisecond}, motion.AtLabel("arrival", 0))
// Render 中：
f := v.timeline.Sample(cx, motion.TimelineOptions{Run: v.opened})
card.AnimatePose(el.Pose{X: f.Value("x"), Fade: f.Value("fade")}, el.Instant)
```

`AfterPrevious` 相对于上一轨道的结束时间，负数表示重叠；`WithPrevious` 相对于
开始时间；`AtTime` 使用绝对时间。轨道名称不可重复，首次采样后不能再修改编排。
`TimelineOptions` 支持暂停、倒放、首次及循环间隔延迟、额外循环次数（`-1` 表示
无限）、交替方向，以及保持在指定 `Seek` 时间或 `Progress` 进度。
也可将 `Context.ScrollState` 换算为 `Progress` 来自定义滚动播放区间。
减少动态效果时仍可明确跳转；其余播放直接到终点，无限循环则结束第一轮。
暂停或完成后不再请求帧，构建后的采样与轨道查询不产生分配。

`Sample` 在 Render 中执行，因此播放期间会重建该视图。将采样结果交给
`el.Instant`，避免再次插值。该 API 不触发回调，也不测量实际显示帧率。
这些能力位于当前工作分支，尚未包含在 v0.4.2 中。

时间线可直接绑定滚动容器：

```go
f := v.timeline.Sample(cx, motion.TimelineOptions{
    Scroll: &motion.ScrollBinding{ID: "list"},
})
```

默认将完整可滚动范围映射为一轮播放；Horizontal 选择横向滚动，Start/End 指定 dp
范围。容器缺失或不可滚动时保持起点。Scroll、Seek、Progress 互斥；内容或视口范围
变化后补绘一帧更新进度，不轮询、不在静止时计时。几何信息来自上一绘制帧。
