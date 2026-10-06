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

## 应用图标

`window.SetIcon(png)` 在程序运行时设置应用图标，参数是满版的正方形 PNG 原图（和脚手架项目的 `appicon.png` 一样），按各平台的规范裁形状：

- **macOS**：程序坞图标，Apple 的连续圆角和阴影。
- **Windows**：每个窗口的标题栏和任务栏按钮，按窗口的 DPI 选尺寸。
- **Linux X11**：每个窗口的 `_NET_WM_ICON`。
- **Linux Wayland、浏览器**：不处理。Wayland 由合成器按 app_id 找已安装的 `.desktop` 文件取图标（`keel build` 生成的安装包会装上），浏览器用页面图标。

打好的包本身就带图标，`SetIcon` 补的是 `go run` / `keel run` 这种直接跑可执行文件的情况，否则系统只会显示通用图标。它对已打开和之后打开的窗口都生效，可以在 `Open` 之前调用，也可以随时换。脚手架生成的 `main.go` 已经用 `//go:embed appicon.png` 把图标编进程序并调用它。

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
