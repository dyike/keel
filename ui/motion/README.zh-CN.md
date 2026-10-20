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
