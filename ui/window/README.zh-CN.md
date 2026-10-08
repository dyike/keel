# ui/window

[English](README.md) | 简体中文

窗口：打开、关闭、置前（含激活令牌）、窗口快捷键、事件循环、离屏截图，以及读取系统偏好、补齐 Gio 没有的平台事件。

| 文件 | 内容 |
| --- | --- |
| `window.go` | `Open`、`Main`、`Options`（含 `Overlay`）、`Window`（`Raise`、`Activate`、`WaylandDisplay` …） |
| `memory.go` | 按需启用的进程级空闲 Go 堆回收 |
| `shortcut.go` | 快捷键解析与分发 |
| `icon*.go` | `SetIcon`：运行时的应用图标（macOS 程序坞、Windows 窗口、X11 `_NET_WM_ICON`），形状由 `internal/appicon` 按平台规范裁出 |
| `root.go` | 窗口根视图：背景、滚动、24dp 边距 |
| `position_*` | 首次显示居中；macOS 按屏幕可用区域计算，其他平台使用 Gio 动作 |
| `screenshot.go` | `Screenshot` 离屏渲染成 PNG |
| `decorations*.go` | Linux 合成器不画标题栏时由 Keel 绘制 |
| `titlebar_darwin.*` | macOS 无边框窗口的标题栏拖动区域 |
| `scene_ios.*` | 实验性 iOS 单场景生命周期适配，需在 Info.plist 配置 `KeelSceneDelegate`，见 [iOS 验证](../../docs/ios.zh-CN.md) |
| `activation_*` | `Activate(token)`：Wayland 走 xdg-activation，X11 写启动 ID 后请求激活 |
| `motion_*` | 系统偏好：减少动态效果（macOS）、滚动条自动隐藏（macOS、Windows），写进 `theme` |
| `scroll_darwin.m`、`scroll_wayland*` | 滚动的设备和手势阶段（触控板抬手、滚轮），Gio 不提供，交给 `core.ReportScrollGesture` |
| `automation.go` | 自动化模式：内存窗口、语义快照、模拟点击输入滚动 |
| `automation_server.go` | 自动化协议：`KEEL_AUTOMATION` socket 上的 JSON 请求 |
| `testdata/raise` | 真实窗口死锁回归测试 |

- **依赖**：`core`、`theme`、`internal/appicon`（图标形状，和脚手架共用），以及 Linux 上的 `jezek/xgb`（X11 激活）和 libwayland-client（Gio 本来就链接）。不依赖 `el`、`kit`：窗口只认 `core.Widget` 接口，系统偏好经 `theme` 交给 el。
- **被谁依赖**：应用代码。`cmd/keel-mcp` 通过 socket 协议驱动它，不引用它的代码。

设置环境变量 `KEEL_AUTOMATION=1`（或 socket 路径）启动应用时，每个窗口会多一个影子窗口，供 Agent 操作；真实窗口照常显示，Agent 的操作会实时反映在屏幕上。再加 `KEEL_HEADLESS=1` 则不显示窗口，见 [Agent 端到端测试](../../docs/automation.zh-CN.md)。

```go
window.Open(window.Options{Title: "Hello", Content: page})
window.Main() // 最后一个窗口关闭后退出进程
```

改这里的代码前先读 [架构 · 不能在锁内等待主线程](../../docs/architecture.zh-CN.md#不能在锁内等待主线程)。详见 [窗口与应用](../../docs/app.zh-CN.md)。

macOS 的 `Main` 会订阅 NSWorkspace 的辅助功能显示偏好和滚动条样式，启动时读取“减少动态效果”和“显示滚动条”，变化时更新 `theme.ReducedMotion` 和 `theme.SystemScrollbarsAutoHide`。AppKit 回调通过有界队列交给后台消费者，再在 `core.Update` 中更新主题，避免主线程等待帧锁。Windows 启动时读一次“自动隐藏滚动条”。无窗口和离屏自动化模式不安装原生观察者，Linux 目前使用应用设置。

## macOS 原生交通灯布局

无边框窗口可以保留系统按钮，并按项目的标题栏高度居中。尺寸单位为 dp，按钮保留 AppKit 的外观、大小和行为。`TrafficLightLayout` 为 nil 时使用系统默认位置；`Height` 为正值时启用自定义位置。`OffsetY` 正值向下，负值向上；`Spacing` 为 0 时保留系统按钮间距。

```go
w := window.Open(window.Options{
    Frameless: true,
    NativeTrafficLights: true,
    TrafficLightLayout: &window.TrafficLightLayout{
        Height: 44, Left: 15, Spacing: 23,
    },
    Content: page,
})
// 标题栏样式变化时，无须重新创建窗口。
w.SetTrafficLightLayout(window.TrafficLightLayout{Height: 64, Left: 20})
```

应用需为按钮预留左上区域。调整窗口大小、切换全屏及系统重新布局后，Keel 会恢复配置的位置。此选项仅在 macOS 生效；运行时设置可以在 UI 回调或后台 goroutine 中调用。

应用菜单和系统外观通过 Go API 配置：`NewMenuBar`、`SetApplicationMenu`、`SystemAppearance`、`SetNativeAppearance`。自定义方式和平台能力见[窗口与应用](../../docs/app.zh-CN.md#应用菜单与系统外观)。

菜单后端：macOS 使用 AppKit，Windows 使用 Win32，Linux（X11/Wayland）由 Keel 绘制窗口内菜单。`Options.MenuDisplay` 可选择窗口内菜单或应用自行绘制；自定义编辑器通过 `core.NextEditAction` 处理标准编辑操作。
