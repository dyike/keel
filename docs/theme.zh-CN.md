# 主题

[English](theme.md) | 简体中文

`ui/theme` 管颜色、尺寸刻度和字体。组件每次 Render 都从 theme 读取，所以切换主题后所有窗口下一帧就换成新样子，不用重建视图。

## 切换主题

```go
theme.Use("nord")          // 按名字应用已注册的主题，并记住名字
theme.CurrentName()        // "nord"
theme.Names()              // 已注册的全部名字，可以直接给主题选择框用
theme.Apply(myPalette)     // 应用一个没有名字的调色板
```

`Use` 和 `Apply` 要在打开窗口前调用，或者在界面回调里调用（持有帧锁）。其他 goroutine 用 `core.Update(func() { theme.Use("dark") })`。

内置主题：`light`、`dark`、`nord`、`paper`、`solarized-dark`、`high-contrast`，以及带渐变的 `aurora`。

## 调色板与缓存

`Light()`、`Dark()`、`Current()` 返回调色板副本。自定义时修改副本，再用 `Apply` 替换全部颜色；零值字段也会应用。

```go
p := theme.Light()
p.Primary = theme.RGB(0x15803d)
theme.Apply(p)
```

`Apply` 递增 `theme.Revision()`，使 `cx.Cache` 失效并请求重绘。自己的渲染缓存也要把版本号纳入键。颜色应在 Render 时读取；已经构造好的静态元素、自定义固定色和 Markdown 独立配色需要应用自行更新。直接修改包级颜色变量不会同步 Material、刷新缓存或请求重绘。

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

`OnColor` 用于 Primary / Danger 实心背景，不保证适合所有状态色背景。成功、警告和提示颜色可以用于正文和图标。

## 主题文件

主题文件是 JSON：选一个基础主题，只写要改的颜色。颜色名就是 `Palette` 的字段名，不区分大小写，值可以写成 `#rgb`、`#rgba`、`#rrggbb` 或 `#rrggbbaa`。

```json
{
  "name": "forest",
  "base": "light",
  "colors": {
    "bg": "#e8f3ea",
    "primary": "#2f7d4a",
    "primaryHover": "#26693d",
    "chart": ["#2f7d4a", "#c2410c"]
  }
}
```

```go
name, palette, err := theme.ParseTheme(data)
theme.Register(name, palette)
```

`chart` 最多 8 个颜色，没写的沿用基础主题。写错颜色名或颜色值时 `ParseTheme` 返回错误。

## 渐变

`bgGradient` 给窗口背景，`primaryGradient` 给主按钮、进度条和用户的聊天气泡：

```json
"bgGradient":      {"from": "#10142b", "to": "#1f1843", "angle": 120},
"primaryGradient": {"from": "#7c5cff", "to": "#2fb8ff", "angle": 0}
```

`angle` 的单位是度：0 从左到右，90 从上到下。主按钮悬停和按下时整体压暗，渐变不消失。

渐变只是外观，`Bg` 和 `Primary` 仍然是这两处的纯色，选中边框、焦点框、链接等其他地方都用纯色。所以要把 `bg`、`primary` 设成渐变里的某个颜色，看起来才协调。没设渐变（零值）时完全不影响原来的纯色。

自己画的元素也能用：`el.Div().BgGradient(theme.Gradient{From: a, To: b, Angle: 90})`。传零值什么都不改，所以可以直接传 `theme.PrimaryGradient`，没配渐变的主题下它就是纯色。之后再调 `Bg` 会替换掉渐变。

## 监听主题目录

设计师改主题文件时，不用重启程序就能看到效果：

```go
w, err := theme.WatchThemes("./themes", time.Second, func(names []string, err error) {
    if err != nil {
        log.Print(err) // 某个文件写错了；其他文件照常加载
    }
})
defer w.Stop()
theme.Use("forest")
```

- 启动时立刻注册目录里所有的 `*.json`，所以紧接着就能 `Use` 它们。
- 之后按间隔检查文件的修改时间。有变化的文件在下一帧（持有帧锁时）重新注册；正在使用的主题（`CurrentName`）变了，会立刻重新应用。
- 用轮询而不是文件系统事件，不需要额外依赖，各平台行为一致。
- 删掉文件不会注销已注册的主题。
- 组件库示例可以这样试：`go run ./examples/components -themes ./mythemes -theme forest`。

## 尺寸刻度

组件从刻度里取值，不自己写数字，同一种用途在各处看起来一样，要改设计也只改一处。

| 刻度 | 常量 |
| --- | --- |
| 间距（dp） | `SpaceXxs` 2、`SpaceXs` 4、`SpaceSm` 6、`SpaceMd` 8、`SpaceLg` 12、`SpaceXl` 16、`Space2xl` 24、`Space3xl` 32 |
| 圆角（dp） | `RadiusSm` 4、`RadiusMd` 6、`RadiusLg` 8、`RadiusXl` 12、`RadiusFull` |
| 字号（sp） | `TextXs` 11、`TextSm` 12、`TextMd` 13、`TextControl` 14、`TextBody` 15、`TextLg` 17、`TextXl` 20、`TextHeading` 22 |
| 阴影 | `ElevationSm`、`ElevationMd`、`ElevationLg`，颜色是 `theme.Shadow` |

间距的用法：图标和文字之间用 `SpaceSm`，一行里的控件之间用 `SpaceMd`，控件内边距和表单字段之间用 `SpaceLg`，卡片内边距用 `SpaceXl`，对话框和页面内边距用 `Space2xl`。

```go
el.Div().Row().Gap(theme.SpaceMd).P(theme.SpaceXl).Rounded(theme.RadiusLg)
```

kit 的间距都已改用这套刻度；剩下的 10、14、20dp 是刻意的视觉微调，比如让输入框文字和按钮文字对齐。

## 局部主题

一个窗口里的某一块想换颜色（比如浅色窗口里的深色侧栏），用 `cx.Themed`，见[元素与视图 · 局部主题](el.zh-CN.md#局部主题)。

自己写 Gio 绘制代码时，可以用 `theme.Scope(p)` 临时替换调色板，调用返回的函数恢复；Scope 不触发重绘或改变版本号。局部主题不会自动跟随系统外观。

## 字体

`theme.Face` 指定正文的字体优先级，逐字形回退；`theme.MonoFace` 指定等宽字体优先级。桌面版优先使用系统字体，兜底字体只在系统没有对应字体时使用。

`MonoFace` 优先尝试等宽字体，缺失的中文等字形复用完整的 `Face` 回退列表。macOS 的字体索引没有 PingFang 时，终端中文可以和正文一样使用 Hiragino Sans GB，避免落到 Arial Unicode MS 等任意汉字回退字体。字体内存和中文字形外观取决于实际选择的字体；拉丁文字保留等宽字体。

`theme.LoadFonts(data...)` 接收 TTF、OTF、TTC 文件的字节内容，加载后重绘所有窗口。浏览器通过 `theme.FetchFonts("font.ttf")` 下载并加载；中文字体准备见 [在浏览器里运行](web.zh-CN.md#构建)。

`BodySize` / `SmallSize` / `HeadingSize` 分别为 15 / 13 / 22sp，标准单行字段高 `ControlHeight`（36dp）。直接使用 Gio 绘制时可复用 `theme.Material` 的字形排版器。

## 系统减少动画

macOS 运行 `window.Main()` 后默认跟随系统“减少动态效果”，包括运行时变化。应用可以覆盖或恢复这个偏好：

```go
theme.SetReducedMotion(true) // 显式关闭动画
theme.FollowSystemMotion()   // 恢复跟随最近的系统值
```

在 UI 回调或 `core.Update` 内调用。自动化模式使用显式覆盖，保持截图稳定；其他平台默认允许动画，应用仍可关闭。

### 字形图片缓存

自定义终端绘制器可以按需启用 `theme.GlyphAtlas`。每帧先调用 `BeginFrame(shaper)`，为所有文字段调用 `Prepare(params, glyphs, color)`，再调用 `Commit()` 和 `Paint(ops, params, glyphs, color)`。只准备当前帧，不预计算后续画面。发生变化的页面每帧生成一个不可变的 `ImageOp`；绘制器关闭时调用 `Release()`。

调用方应使用整数像素平移，字号已换算为物理像素，并且不额外缩放或旋转。其他变换使用 `GlyphPainter`。复杂文字段、半透明颜色、GPU 不可用或缓存超限时回退到矢量绘制；彩色位图字形保留 Gio 的位图绘制。

页面上限为八张 512×512 RGBA 图片（8 MiB），掩码上限为 2 MiB、4096 个条目，每个掩码最多缓存八种颜色。旧页面快照和 GPU 副本还会占用额外内存；`Stats()` 提供当前 CPU 缓存占用和绘制次数。第一帧同步准备所需掩码；后续每帧最多准备 32 个新掩码，连续 60 帧没有新掩码后释放临时 GPU。`SubpixelPhases` 是可选的位置取整，与 `GlyphPainter` 有相同的文字质量取舍。
