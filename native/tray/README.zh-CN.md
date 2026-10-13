# native/tray

[English](README.md) | 简体中文

托盘图标与原生弹出菜单。仅依赖 `native` 和 `native/internal/sys`，不依赖 UI 或 Gio。应用可以与 `ui/window` 组合，也可以独立运行事件循环。

```go
icon, err := tray.New(tray.Options{
    Icon: artwork, // image.Image，每个轴 1..1024 像素，编码为 PNG 快照
    Template: true,
    Tooltip: "My app",
    Menu: []tray.MenuItem{
        {ID: "open", Title: "Open", OnSelect: func() {
            core.Update(openWindow)
        }},
        {Separator: true},
        {ID: "quit", Title: "Quit", OnSelect: window.Quit},
    },
})
if err != nil { log.Fatal(err) }
window.MainWithOptions(window.MainOptions{
    KeepAlive: true,
    OnQuit: func() { icon.Close() },
})
```

`MenuItem` 支持 `Children`、`Disabled`、`Checked`、`Separator` 和 `OnSelect`。字符串 ID 可选，在整棵树中必须唯一。子菜单不能同时有点击回调。菜单树会复制，最多 1024 项、16 层嵌套。文字和提示不能含 NUL。程序更新不触发用户回调。`SetMenu` 整体替换菜单，非法更新保留原菜单；原生 ID 不复用，旧菜单事件不会触发替换后的新动作。

`SetIcon`、`SetTooltip`、`SetMenu` 和 `Close` 支持并发调用。每个托盘在 Go goroutine 中串行执行回调，不在原生事件线程或 UI frame lock 内执行。写 UI 状态仍要经过 `core.Update`。回调可以更新或关闭自身托盘，耗时任务另开 goroutine。Close 可重复调用，丢弃排队事件，已开始执行的回调可以结束；它只移除图标，不退出进程。

| 平台 | 后端与行为 |
| --- | --- |
| macOS | purego 调用 `NSStatusBar` / `NSMenu`。`Template` 标记模板图片，由系统按明暗外观着色。有菜单时点击打开菜单，没有菜单时调用 `OnActivate`。`New` 同步使用主线程，放在 `window.Main` 前，或不持有 frame lock 的后台 goroutine 中。更新和关闭进入主队列，返回时可能还未应用。 |
| Windows | Win32 通知图标，共享隐藏顶层 HWND 和休眠的消息循环。激活触发 `OnActivate`，上下文激活打开原生菜单。Explorer 的 `TaskbarCreated` 广播会重新添加存活图标，替换和移除时释放原生资源。 |
| Linux | 纯 Go 会话 D-Bus，SNI（KDE 与 freedesktop 命名空间）和 DBusMenu。需要运行中的 StatusNotifierHost，没有时 New 返回 `native.ErrUnsupported`。点击与菜单行为由 Host 决定，Watcher 重启后重新注册，不需要 X11 / Wayland 窗口句柄。 |
| 其他平台、iOS、Android、Web | New 返回 `native.ErrUnsupported`。 |

`Template` 只影响 macOS。提示文字、图标缩放与位置由桌面控制；Windows 提示最多 127 个 UTF-16 单元。注册后的图标可能在通知溢出区，也可能被桌面隐藏。不需要权限或 cgo。

独立 macOS 应用在 `main` 开始时调用 `runtime.LockOSThread()`，在该线程创建托盘并运行 `tray.Main()`。Windows / Linux 的 `tray.Main()` 保持进程运行，后端处理原生事件。不要与 `window.Main` 同时运行。独立应用自己管理退出与清理；tray 不引入 UI 生命周期管理器。

运行 `go run ./examples/tray`，默认只有托盘，通过菜单打开窗口，关闭后再打开。`-smoke` 自动验证图标与菜单更新、无窗口重开和显式退出。`-screenshot <path>` 仅渲染配套窗口。Windows / Linux 运行验收仍需要对应桌面；macOS smoke 与跨平台编译是不同的检查。
