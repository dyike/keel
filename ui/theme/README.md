# ui/theme

全局颜色、字号、字体。默认浅色；组件在每帧读取语义色。完整用法见[主题](../../docs/theme.md)。

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

`cx.Cache` 随 Apply 自动失效；应用自己的渲染缓存需将 `theme.Revision()` 纳入键。Render 时重新读取颜色；已构造的静态元素、自定义固定颜色和 Markdown 独立配色不会被自动重写。局部主题用 `Scope`（见下文），不自动跟随系统外观。

直接修改颜色变量不会同步 Material、刷新缓存或请求重绘；运行时请用 Apply。

- 依赖：Gio 和 `ui/internal/loop`（仅通知全局重绘）。
- 被依赖：`el`、`kit`、`window`、`markdown`。

验证入口：`go run ./examples/components -section theme -theme dark`。

| 颜色 | 浅色默认值 | 用在哪里 |
| --- | --- | --- |
| `Bg` | `#f5f6f8` | 窗口背景 |
| `Surface` | `#ffffff` | 卡片、输入框底色 |
| `Border` | `#e3e5e8` | 边框、分隔线 |
| `Text` / `Muted` | `#1f2328` / `#6b7280` | 正文 / 次要文字、占位文字 |
| `Primary` / `PrimaryHover` / `PrimaryText` | `#2563eb` / `#1d4ed8` / `#1d4ed8` | 主按钮、焦点边框 / 悬停 / 链接等蓝色文字 |
| `Danger` / `DangerHover` / `DangerText` | `#dc2626` / `#b91c1c` / `#b91c1c` | 危险按钮 / 悬停 / 错误文字 |
| `Success` / `Warning` / `Info` | `#15803d` / `#a16207` / `#0369a1` | 状态正文、图标 |
| `Subtle` / `SubtleHover` | `#eceef1` / `#e2e5e9` | 次要按钮、悬停底色 |
| `OnColor` | `#ffffff` | 实心主色、危险色背景上的文字 |
| `Highlight` | `#dbeafe` | 选中行、选中项 |
| `Scrim` | 40% 黑 | 模态浮层后面的遮罩 |
| `CodeBg` / `CodeText` | `#f0f1f3` / `#1f2328` | 代码块 |
| `Chart` | 8 个分类色 | 图表系列颜色，按顺序使用；浅色和深色各一套，均通过色觉缺陷校验 |

多主题：`Register(name, p)` 登记调色板，`Named(name)` 取副本，`Names()` 列出全部（内置 light、dark、nord、paper、solarized-dark、high-contrast，以及 aurora，除 light、dark 外都是 `themes/` 里的主题文件），应用拿它做主题选择器，选中后 `Use(name)`（应用并记住名字，`CurrentName()` 读回）。`ParseTheme(json)` 读主题文件：`base` 指定继承 light、dark 或已登记的主题，`colors` 只写要改的颜色，键是 Palette 字段名（大小写不限），值是 `#rgb`、`#rgba`、`#rrggbb` 或 `#rrggbbaa`，`chart` 是最多 8 个颜色的数组，`bgGradient`、`primaryGradient` 是 `{"from", "to", "angle"}` 渐变（内置 `aurora` 用了它们）。`WatchThemes(dir, interval, onChange)` 加载目录里的主题文件，并在文件变化时下一帧重新注册，正在用的主题会重新应用。`Scope(p)` 给窗口的一部分换调色板并返回恢复函数，供 `el.Themed` 使用；它不重绘，也不改 `Revision`。

`LoadFonts(files...)` 加载字体文件（TTF、OTF、TTC）并重绘所有窗口，Web 版必须用它提供中文字体，见[在浏览器里运行](../../docs/web.md)。

尺寸刻度在 `scale.go`：间距 `SpaceXxs` 2 / `SpaceXs` 4 / `SpaceSm` 6 / `SpaceMd` 8 / `SpaceLg` 12 / `SpaceXl` 16 / `Space2xl` 24 / `Space3xl` 32，圆角 `RadiusSm` 4 / `RadiusMd` 6 / `RadiusLg` 8 / `RadiusXl` 12 / `RadiusFull`（药丸和圆），字号 `TextXs` 11 / `TextSm` 12 / `TextMd` 13 / `TextControl` 14 / `TextBody` 15 / `TextLg` 17 / `TextXl` 20 / `TextHeading` 22，阴影层级 `ElevationSm`（提示）/ `ElevationMd`（菜单、弹层、下拉、通知）/ `ElevationLg`（对话框、侧滑面板、命令面板），阴影颜色 `Shadow` 随浅深色切换。`MonoFace` 是等宽字体优先级。

字号常量 `BodySize` / `SmallSize` / `HeadingSize` 为 15 / 13 / 22 sp；`ControlHeight` 为 36dp，是所有单行字段（输入框、下拉框、日期时间、搜索框）的高度；`Face` 是字体优先级（苹方 → 冬青黑体 → 微软雅黑 → Noto Sans CJK → Go），逐字形回退。`Material` 是底层的 Gio `material.Theme`，提供字形排版器，自己写 Gio 代码时用它。
