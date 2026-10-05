# ui/window

窗口：打开、关闭、置前（含激活令牌）、窗口快捷键、事件循环、离屏截图，以及读取系统偏好、补齐 Gio 没有的平台事件。

| 文件 | 内容 |
| --- | --- |
| `window.go` | `Open`、`Main`、`Options`（含 `Overlay`）、`Window`（`Raise`、`Activate`、`WaylandDisplay` …） |
| `shortcut.go` | 快捷键解析与分发 |
| `icon*.go` | `SetIcon`：运行时的应用图标（macOS 程序坞、Windows 窗口、X11 `_NET_WM_ICON`），形状由 `internal/appicon` 按平台规范裁出 |
| `root.go` | 窗口根视图：背景、滚动、24dp 边距 |
| `position_*` | 首次显示居中；macOS 按屏幕可用区域计算，其他平台使用 Gio 动作 |
| `screenshot.go` | `Screenshot` 离屏渲染成 PNG |
| `decorations*.go` | Linux 合成器不画标题栏时由 Keel 绘制 |
| `titlebar_darwin.*` | macOS 无边框窗口的标题栏拖动区域 |
| `activation_*` | `Activate(token)`：Wayland 走 xdg-activation，X11 写启动 ID 后请求激活 |
| `motion_*` | 系统偏好：减少动态效果（macOS）、滚动条自动隐藏（macOS、Windows），写进 `theme` |
| `scroll_darwin.m`、`scroll_wayland*` | 滚动的设备和手势阶段（触控板抬手、滚轮），Gio 不提供，交给 `core.ReportScrollGesture` |
| `automation.go` | 自动化模式：内存窗口、语义快照、模拟点击输入滚动 |
| `automation_server.go` | 自动化协议：`KEEL_AUTOMATION` socket 上的 JSON 请求 |
| `testdata/raise` | 真实窗口死锁回归测试 |

- **依赖**：`core`、`theme`、`internal/appicon`（图标形状，和脚手架共用），以及 Linux 上的 `jezek/xgb`（X11 激活）和 libwayland-client（Gio 本来就链接）。不依赖 `el`、`kit`：窗口只认 `core.Widget` 接口，系统偏好经 `theme` 交给 el。
- **被谁依赖**：应用代码。`cmd/keel-mcp` 通过 socket 协议驱动它，不引用它的代码。

设置环境变量 `KEEL_AUTOMATION=1`（或 socket 路径）启动应用时，每个窗口会多一个影子窗口，供 Agent 操作；真实窗口照常显示，Agent 的操作会实时反映在屏幕上。再加 `KEEL_HEADLESS=1` 则不显示窗口，见 [Agent 端到端测试](../../docs/automation.md)。

```go
window.Open(window.Options{Title: "Hello", Content: page})
window.Main() // 最后一个窗口关闭后退出进程
```

改这里的代码前先读 [架构 · 不能在锁内等待主线程](../../docs/architecture.md#不能在锁内等待主线程)。详见 [窗口与应用](../../docs/app.md)。

macOS 的 `Main` 会订阅 NSWorkspace 的辅助功能显示偏好和滚动条样式，启动时读取“减少动态效果”和“显示滚动条”，变化时更新 `theme.ReducedMotion` 和 `theme.SystemScrollbarsAutoHide`。AppKit 回调通过有界队列交给后台消费者，再在 `core.Update` 中更新主题，避免主线程等待帧锁。Windows 启动时读一次“自动隐藏滚动条”。无窗口和离屏自动化模式不安装原生观察者，Linux 目前使用应用设置。
