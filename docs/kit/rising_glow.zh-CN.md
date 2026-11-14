# RisingGlow

[English](rising_glow.md) | 简体中文

参考当前 [Ruixen Rising Glow 预览](https://ruixen.com/docs/components/rising-glow) 实现的装饰粒子场，使用原生 Go/Gio，不依赖网页运行时，不新增 cgo。

参考页面的文字说明称其为文字光晕包装器，但当前预览和安装代码实际是在标题**下方**单独绘制粒子。Keel 遵循实际实现：粒子直径 4–10dp、随机横向位置、初始延迟 0–2 秒，每个粒子保持各自 3–5 秒的周期。每周期先上升、再下降，同时淡入淡出；两个半周期都使用 CSS ease-in-out 曲线。粒子裁剪在区域内，不额外添加模糊或文字阴影。

    glow := kit.RisingGlow().
        ParticleCount(80).
        Color(color.NRGBA{R: 0x13, G: 0x8d, B: 0xff, A: 255}).
        Height(100)

    // glow 保存在视图状态里，不要在 Render 中重新创建。
    return el.Div().Items(el.Center).Child(
        el.Text("keel").TextSize(60).Bold(),
        el.Div().WFull().MaxW(el.Dp(384)).Child(glow.Render(cx)),
    )

导入 image/color、github.com/dyike/keel/ui/el 和 github.com/dyike/keel/ui/kit。

| 方法 | 行为 |
| --- | --- |
| Width(dp) | 零值（默认）填满可用宽度，正数固定 dp 宽度。 |
| Height(dp) | 正数 dp 高度，默认 120。 |
| ParticleCount(n) | 0–512，默认 20。零隐藏粒子并停止动画帧；改变数量会重新生成粒子。 |
| Color(c) | 默认跟随当前主题的 PrimaryText；示例显式使用参考蓝色。 |
| Enabled(on) | 默认启用。关闭后显示静态代表姿态，重新启用会重播。 |

非法尺寸、数量会被忽略。重复设置相同数量或启用状态不会重置动画。粒子参数只在数量变化时重新生成，不在每帧随机生成。减少动态效果模式同样显示静态姿态。完全透明、空粒子场、被完全裁剪的区域不请求动画帧。可见动画使用 Keel 的环境动画绘制频率和保留布局；播放时属于持续装饰动画，并非空闲。

组件不增加无障碍角色、标签、焦点或输入处理，标题和控件仍是独立普通元素。在横向布局中使用时请提供有界容器；显示密度只影响栅格分辨率，不改变区域的 dp 尺寸。

运行示例：

    go run ./examples/components -section rising_glow
