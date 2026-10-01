# ui/theme

全局颜色、字号、字体。默认浅色；组件在每帧读取语义色。

```go
// 开窗前或 UI 回调中同步切换。
theme.Apply(theme.Dark())
theme.Apply(theme.Light())

// 后台 goroutine 中排队到帧锁内。
core.Update(func() { theme.Apply(theme.Dark()) })

// 自定义完整调色板，不修改预设。
p := theme.Light()
p.Primary = theme.RGB(0x15803d)
theme.Apply(p)
```

`Palette` 包含全部颜色，`Light()`、`Dark()`、`Current()` 返回副本。`Apply` 替换全部颜色（零值也会应用），同步 `Material.Palette`，递增 `Revision()`，请求所有窗口重绘；保留 Material 指针、字体和排版器。`Apply` 不自己取得帧锁：回调本来已持锁，后台必须使用 `core.Update`。`Current`、`Revision` 和颜色变量的读取也遵守帧锁规则。

`Success` 表示成功，`Warning` 表示警告，`Info` 表示提示；可用作正文或图标颜色。`OnColor` 用于 Primary/Danger 实心背景上的文字，不保证适合所有语义色背景。

`cx.Cache` 随 Apply 自动失效；应用自己的渲染缓存需将 `theme.Revision()` 纳入键。Render 时重新读取颜色；已构造的静态元素、自定义固定颜色和 Markdown 独立配色不会被自动重写。没有局部主题作用域，也不自动跟随系统外观。

兼容直接修改颜色变量的旧写法，但它不会同步 Material、刷新缓存或请求重绘；运行时请用 Apply。

- 依赖：Gio 和 `ui/internal/loop`（仅通知全局重绘）。
- 被依赖：`el`、`kit`、`layout`、`widget`、`window`、`markdown`。

验证入口：`go run ./examples/components -section theme -theme dark`。颜色列表见[组件与布局 · 主题](../../docs/widgets.md#主题)。
