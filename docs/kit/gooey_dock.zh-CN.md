# GooeyDock

[English](gooey_dock.md) | 简体中文

基于 [Ruixen Gooey Dock](https://ruixen.com/docs/components/gooey-dock) 的原生指针感应 Dock。请保留组件实例，不要每帧重新创建。

```go
dock := kit.GooeyDock(
    kit.GooeyDockItem{ID: "home", Label: "Home", Icon: kit.IconHome, Active: true, Action: openHome},
    kit.GooeyDockItem{ID: "search", Label: "Search", Icon: kit.IconSearch, Action: openSearch},
)
// 在 Render 中：
return dock.Render(cx)
```

沿用参考源码的固定尺寸：图标单元从 40 放大到 58 dp，上浮 14 dp，余弦感应半径 150 dp，20 dp 图标放大到 1.2 倍，弹簧质量/刚度/阻尼为 0.1/200/14。栏内间距 2 dp，上/横向/下内边距 8/12/10 dp，圆角 18 dp，边框 1 dp。悬停标签以 150 ms 动画进入和退出。颜色适配浅色和深色主题。栏上方另预留 44 dp，容纳上浮单元和标签。父容器应为放大后的 Dock 留足横向空间，不会压缩图标来适应窄屏。

`IconView` 可传入保留的 SVG 图标代替 `Icon`。Components 示例使用参考中的七个 Lucide 图标，Home 显示激活圆点。`SetItems` 替换项目并重置弹簧，ID 必须非空且唯一。`SetActive` 更新圆点而不触发动作。`SetDisabled` 和项目的 `Disabled` 禁止交互。方向键、Home、End 移动焦点，Enter/Space 激活动作。减少动态效果模式直接更新状态，弹簧稳定后不再请求帧。

`OnHover(func(id string))` 观察指针进入不同项目，间隔限制为 80 ms，供应用接入音频。本实现**尚未自行播放**原版的 Web Audio 轻提示音。视觉参数依据源码，尚未完成浏览器与原生版本的逐像素、逐帧一致性验证；字体栅格化和系统文字缩放也可能存在差异。

```sh
go run ./examples/components -section gooey_dock
```

完全由 Go 实现，没有新增 cgo、JavaScript 或 WebView 依赖。

`Surface(color)` 指定 Dock 背后的不透明底色，默认 `theme.Bg`。放在其他底色上时请设置它：Gio 在线性光空间混合透明色，组件会依据底色预先计算 CSS 风格的 sRGB 混合颜色。
