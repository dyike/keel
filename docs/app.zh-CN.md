# 窗口与应用

[English](app.md) | 简体中文

`ui/window` 模块负责：打开、关闭、置前、窗口快捷键，以及不开窗口直接渲染成 PNG。

## 打开窗口

```go
w := window.Open(window.Options{
    Title:   "设置",
    Width:   420,          // dp，0 表示 640
    Height:  340,          // dp，0 表示 480
    Content: el.Root(page), // core.Widget，通常是 el.Root(视图)
    Shortcuts: map[string]func(){
        "mod+s": save,
        "esc":   func() { w.Close() },
    },
    OnClose: func() { log.Print("设置窗口已关闭") },
})
window.Main()
```

| 字段 | 说明 |
| --- | --- |
| `Title` | 窗口标题 |
| `Width`、`Height` | 初始尺寸，单位 dp（在 2 倍屏上 1dp = 2 像素）。用户可以拖动改变 |
| `Content` | 窗口内容，通常是 `el.Root(view)`：占满窗口，滚动和边距由视图自己决定。其他 `core.Widget` 放进自带滚动条的根视图，四周留 24dp 边距，背景色 `theme.Bg` |
| `Overlay` | 盖在整个窗口上的一层，给不用 el 的自定义 Gio 内容用；el 视图的对话框、菜单用 `cx.Overlay`。没东西显示时应该不占空间 |
| `Shortcuts` | 窗口获得焦点时生效的快捷键，写法见下文 |
| `OnClose` | 窗口销毁后调用，在锁内运行，可以直接改组件 |
| `Frameless` | 隐藏系统标题栏，内容从窗口最上沿开始，由应用自己画标题栏，通常用 [kit.TitleBar](kit/title_bar.zh-CN.md)。不设时，Linux 上合成器不画标题栏（WSLg、GNOME Wayland）的情况下由 keel 画一个跟随主题和字体的标题栏 |

`window.Open` 可以在 `window.Main()` 之前调用，也可以在任何回调里调用，用来运行时开新窗口。

新窗口首次显示后默认居中，后续拖动、缩放和置前不会重新定位。macOS 按所在屏幕扣除菜单栏和 Dock 后的可用区域居中，计算包含系统标题栏；无边框窗口同样适用。其他平台向 Gio 请求居中，最终位置由平台和窗口管理器决定。

`window.Main()` 必须在 `main` goroutine 里调用，而且不会返回。最后一个窗口关闭时进程退出。

## macOS 玻璃背景

在 `Options.Glass` 中启用原生背景，并让需要透出玻璃的区域保持透明：

```go
page := el.Root(el.ViewFunc(func(*el.Context) el.Element {
    return el.Div().Row().Bg(color.NRGBA{}).Child(
        el.Div().W(el.Dp(220)).P(20).Child(el.Text("玻璃侧栏")),
        el.Div().Grow().Bg(theme.Bg).P(24).Child(el.Text("实色内容区")),
    )
}))
window.Open(window.Options{
    Title: "Glass", Width: 900, Height: 600,
    Glass: &window.GlassOptions{Style: window.GlassRegular, CornerRadius: 20},
    Content: page,
})
window.Main()
```

上例另需导入 `image/color`。`el.Root` 默认填充主题背景，因此最外层必须显式调用 `Bg(color.NRGBA{})`；其他容器的实色背景也会遮住玻璃。可以只让侧栏、顶部透明，其余区域正常绘制。

`GlassRegular`、`GlassClear` 在 macOS 26+ 使用 `NSGlassEffectView`，构建需要 macOS 26+ SDK；旧系统或旧 SDK 自动使用 `NSVisualEffectView`。`GlassFrosted` 始终使用传统毛玻璃。`CornerRadius` 单位为 dp，零保留系统玻璃默认曲率；无效样式、负数或非有限圆角会触发 panic。

玻璃窗口使用独立的透明 Metal 表面。AppKit 负责背后窗口的采样和玻璃效果，Gio 继续处理文字、布局和输入。`GlassSupported()` 查询当前构建的原生背景能力，`LiquidGlassSupported()` 查询 Liquid Glass 能力。Windows、Linux、浏览器、iOS 和 `-tags=nometal` 构建保留实色主题背景。未配置 `Glass` 的窗口使用原有渲染路径。

材质随原生外观和系统辅助功能设置变化；应用仍需为 Go 内容选择合适的调色板。此接口覆盖整个客户区，不提供单个 Go 控件的折射或玻璃融合动画。`Screenshot` 和自动化离屏图无法包含原生合成效果，应在真实窗口验证。

```sh
go run ./examples/glass -backdrop
go run ./examples/glass -style clear -dark
go run ./examples/glass -style frosted
go run ./examples/glass -opaque
```

`-backdrop` 打开一个彩色参考窗口，便于观察透光；示例支持导航、切换明暗、打开 Clear 和毛玻璃对比窗口。

## 空闲时归还堆页

窗口空闲后，Keel 把不再使用的 Go 堆页归还给系统。仅启动过程（读取系统字体索引、最初几次布局）就会留下几十 MB 已释放的堆，Go 运行时本会保留几分钟，而系统把它们全算作占用：在 macOS 上，hello 示例空闲时的内存因此减少约一半。有低延迟后台任务的应用可以在打开窗口前关闭：

```go
window.SetIdleMemoryReclaim(false)
```

这是进程级策略，默认开启。全部 Keel 窗口连续两秒没有 UI 帧或回调，且至少有 32 MiB 空闲堆页尚未归还系统时，才执行回收。策略避开正在执行的 UI 回调，最多每 30 秒回收一次；静止窗口不会为此重绘，也不会周期性反复 GC。传入 `false` 也会取消尚未执行的任务。

回收使用 Go 的 `debug.FreeOSMemory`，其中包含一次 GC；不修改 `GOGC` 或 `GOMEMLIMIT`，不释放仍在使用的缓存和 GPU 纹理。GC 可能短暂停顿其他 goroutine，Keel 外部的后台任务也不属于 UI 空闲判断范围。有低延迟后台任务的应用可在目标设备上验证交互延迟与内存后关闭它。

## 应用图标

`window.SetIcon(png)` 在程序运行时设置应用图标，参数是满版的正方形 PNG 原图（和脚手架项目的 `appicon.png` 一样），按各平台的规范裁形状：

- **macOS**：程序坞图标，Apple 的连续圆角和阴影。
- **Windows**：每个窗口的标题栏和任务栏按钮，按窗口的 DPI 选尺寸。
- **Linux X11**：每个窗口的 `_NET_WM_ICON`。
- **Linux Wayland、浏览器**：不处理。Wayland 由合成器按 app_id 找已安装的 `.desktop` 文件取图标（`keel build` 生成的安装包会装上），浏览器用页面图标。

打好的包本身就带图标。开发时，`keel run` 读取 `keel.json` 的 `icon`、`icon_mask` 和 `icons` 平台覆盖配置，在打开首个窗口时设置图标，无需打包或额外编写应用代码。直接使用 `go run` 时可以调用 `SetIcon`；脚手架已嵌入 `appicon.png` 并调用它。主动调用的 `SetIcon` 优先于命令行传入的图标，对已打开和之后打开的窗口都生效，也可以随时更换。

## Window 的方法

| 方法 | 说明 |
| --- | --- |
| `Close()` | 等同于用户点关闭按钮 |
| `Raise()` | 把窗口置于最前 |
| `Activate(token)` | 用别的程序给的激活令牌（如系统通知被点击时的 `Activation.Token`）把窗口置前，可以通过窗口管理器的防抢焦点；Wayland 走 xdg-activation，X11 写启动 ID 后请求激活，其他平台或空令牌等同 `Raise` |
| `WaylandDisplay()` | Linux Wayland 下这个窗口的 wl_display，其他情况为 nil；读剪贴板时交给 `native/clipboard.UseWaylandDisplay` |
| `Minimize()` | 最小化到 Dock 或任务栏 |
| `ToggleMaximize()` | 最大化（macOS 上是缩放），已最大化时还原 |
| `Maximized()` | 当前是否最大化，在 UI 代码里读取 |
| `Frameless()` | 是否无边框窗口 |
| `Closed() bool` | 窗口是否已销毁。在回调或 `core.Update` 里调用 |

`Close`、`Raise` 和 `Activate` 是异步的：调用立即返回，动作在当前回调结束后才执行。所以 `w.Close()` 之后马上读 `w.Closed()`，得到的仍是 `false`。它们这样设计是为了避免死锁，原因见[架构 · 不能在锁内等待主线程](architecture.zh-CN.md#不能在锁内等待主线程)。

窗口关闭后不能重新打开。"同一时间只保留一个设置窗口"的写法：

```go
var settings *window.Window

openSettings := func() {
    if settings != nil && !settings.Closed() {
        settings.Raise()
        return
    }
    settings = window.Open(window.Options{Title: "设置", Content: settingsPage()})
}
```

完整示例见 `examples/multiwindow`。

## 窗口快捷键

写法是 `修饰键+按键`，不区分大小写：

| 修饰键 | 含义 |
| --- | --- |
| `mod` | macOS 上是 ⌘，Windows、Linux 上是 Ctrl。跨平台快捷键优先用它 |
| `cmd` / `command` / `super` | ⌘ |
| `ctrl` / `control` | Ctrl |
| `shift` | Shift |
| `alt` / `option` | Alt / ⌥ |

按键可以是单个字符（`a`、`,`、`1`）、`f1`–`f12`，或 `esc`、`enter`、`space`、`tab`、`backspace`、`delete`、`up`、`down`、`left`、`right`。

例子：`"mod+,"`、`"mod+shift+s"`、`"esc"`、`"f5"`。写错会在 `window.Open` 时 panic，因为这是写死在代码里的配置，应该在开发时就暴露出来。

窗口快捷键只在该窗口有焦点时生效。要在其他应用在前台时也能触发，用 [`native/hotkey`](native.zh-CN.md#hotkey全局快捷键)。

## 离屏截图

```go
err := window.Screenshot(content, 640, 480, "out.png")
```

用和窗口相同的根视图（背景、边距、滚动）把 `content` 渲染成 PNG，尺寸单位是 dp，输出图片按 2 倍缩放，上面的例子得到 1280×960 像素。不需要 `window.Main()`，也不会弹出窗口。用途：

- 给文档、PR 配图；
- 改样式后对比前后截图，确认没有意外变化（见[测试](testing.zh-CN.md#截图对比)）。

截图只渲染一帧，看不到悬停、按下等交互状态。

## 应用菜单与系统外观

业务项目只需要 Go API，AppKit 桥接由 Keel 管理。菜单内容不固定：支持嵌套子菜单、分隔线、快捷键、勾选、禁用、标准编辑动作和自定义回调。

```go
menu, err := window.NewMenuBar(window.MenuItem{
    Title: "Shell",
    Children: []window.MenuItem{
        {ID: "new-tab", Title: "新标签页", Shortcut: "mod+t", OnSelect: newTab},
        {Separator: true},
        {ID: "copy", Title: "复制", Shortcut: "mod+c", Action: window.MenuCopy},
    },
})
if err != nil { panic(err) }
if err := menu.Install(); err != nil { panic(err) }
// 可在 UI 回调或后台任务中更新；重新安装菜单由 Keel 调度。
_ = menu.UpdateItem("new-tab", func(item *window.MenuItem) {
    item.Title = "新建会话"
    item.Disabled = busy
})
```

`mod` 在 macOS 表示 Command，在其他平台表示 Ctrl。`OnSelect` 在 UI 更新队列中执行；设置它后会覆盖同一项的 `Action`。标准编辑动作包括复制、剪切、粘贴、全选、撤销和重做，macOS 会转发给焦点视图。`Role` 可以标识应用、编辑、窗口、帮助和服务菜单，内容仍由应用决定。

`SetApplicationMenu` 适合一次性配置；`NewMenuBar` 返回可更新的控制器。`SetItems` 可替换整个菜单，`UpdateItem` 可修改单项，验证失败不改变原有内容。显式 `ID` 用于定位；省略时自动生成。`Items` 返回副本，应用可以据此绘制自己的菜单，点击时在 UI 回调中调用 `Invoke`。禁用父菜单也会禁用其子项。

`NativeApplicationMenu()` 查询当前构建的原生菜单后端：macOS 使用 AppKit，Windows 使用 Win32 菜单栏。Linux 自动在窗口内绘制菜单，同时兼容 X11 和 Wayland；不要求桌面环境提供全局菜单服务。三者都使用相同的 Go 配置，支持子菜单、分隔线、勾选、禁用、快捷键和动态更新。

`Options.MenuDisplay` 默认为 `MenuDisplayAuto`。Windows 的普通窗口使用 Win32 菜单栏，由系统决定高度和 DPI 缩放；无边框窗口自动改用 Keel 绘制的菜单，因为 Gio 的无边框模式会隐藏原生菜单所在的非客户区。`MenuDisplayWindow` 强制使用 Keel 绘制的菜单；`MenuDisplayHidden` 隐藏窗口菜单，让应用通过 `Items` 自行绘制。macOS 的系统菜单栏属于应用，不受窗口隐藏选项影响。窗口内菜单支持鼠标、F10 / Alt+M 打开、方向键导航、Enter 执行、Escape 或外部点击关闭，长菜单可纵向滚动。窗口变窄时，顶部标题截断显示，不产生横向滚动条。

标准编辑动作保留编辑器焦点，Keel 的输入框、文本域和富文本输入会自动处理复制、剪切、粘贴、全选、撤销和重做；密码框、只读框继续遵守原有编辑限制。自定义编辑器在 `Layout` 中调用 `core.NextEditAction(gtx, focusTag)`，按自身语义处理动作；例如终端处理复制、粘贴和全选，撤销不应发送 shell 控制字符。设置 `OnSelect` 仍可完全覆盖标准动作。菜单快捷键不能替代全局快捷键。

`SystemAppearance()` 返回系统的浅色/深色偏好：macOS 读取系统偏好，Windows 读取用户主题，Linux 异步读取桌面门户并缓存，未获取到时返回浅色。`SetNativeAppearance(window.AppearanceDark)` 设置原生窗口和菜单外观，`AppearanceSystem` 恢复跟随系统；它不会替换应用自己的 Go 调色板。`NativeAppearanceSupported()` 可查询原生外观设置能力，目前仅 macOS 支持；其他平台保留原生外观。应用仍可自由调用 `theme.Apply` 使用自定义主题。

原生测试桥只在 `-tags=keelnativeqa` 下编译，供开发测试调用菜单、输入法和图标校验；正常构建不包含这些测试接口。

## 当前窗口尺寸

`Window.Size()` 返回最近一帧实际观察到的客户区宽高，单位为 dp，首帧前返回初始请求尺寸。它包含 Keel 绘制的窗口内菜单和标题栏，不包含操作系统的外框；调整大小、缩放、最大化后都会更新，后台读取也安全。

`Window.Resize(width, height)` 请求新的客户区尺寸，参数必须为正数，支持从 UI 回调和后台任务调用。操作系统可能限制请求，实际结果以 `Size()` 和 `Options.OnResize` 为准。已关闭窗口忽略请求。`OnResize` 在第一帧和尺寸改变时执行，回调中的 `Size()` 已更新，回调运行在 UI 队列中。

```go
w := window.Open(window.Options{
    Width: savedWidth, Height: savedHeight,
    OnResize: func(width, height int) { rememberSize(width, height) },
    Content: page,
})
width, height := w.Size()
if err := w.Resize(960, 640); err != nil { panic(err) }
```

应用自行决定尺寸保存位置和时机；布局仍应使用当前视口尺寸，而非启动时的宽高。

## 最小窗口尺寸

`Options.MinWidth`、`Options.MinHeight` 设置客户区最小宽高，单位为 dp，零表示该轴不设下限。例如 `window.Options{Width: 1057, Height: 639, MinWidth: 800, MinHeight: 480}`。

Keel 将下限交给 macOS、Windows、Linux X11/Wayland 的窗口驱动，限制用户拖动缩小窗口；无边框窗口也会传入同样的限制。启动尺寸（包括应用恢复的尺寸）以及 `Resize` 请求小于下限时会提升到下限，离屏自动化遵守同样的规则。Linux 窗口管理器或合成器最终决定是否执行尺寸提示。
