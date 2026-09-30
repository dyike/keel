# ui/window

窗口：打开、关闭、置前、窗口快捷键、事件循环、离屏截图。

| 文件 | 内容 |
| --- | --- |
| `window.go` | `Open`、`Main`、`Options`（含 `Overlay`）、`Window` |
| `shortcut.go` | 快捷键解析与分发 |
| `root.go` | 窗口根视图：背景、滚动、24dp 边距 |
| `screenshot.go` | `Screenshot` 离屏渲染成 PNG |
| `automation.go` | 自动化模式：内存窗口、语义快照、模拟点击输入滚动 |
| `automation_server.go` | 自动化协议：`KEEL_AUTOMATION` socket 上的 JSON 请求 |
| `testdata/raise` | 真实窗口死锁回归测试 |

- **依赖**：`core`、`theme`。不依赖 `widget` 和 `layout`：窗口只认 `core.Widget` 接口。
- **被谁依赖**：应用代码。`cmd/keel-mcp` 通过 socket 协议驱动它，不引用它的代码。

设置环境变量 `KEEL_AUTOMATION=1`（或 socket 路径）启动应用时，每个窗口会多一个影子窗口，供 Agent 操作；真实窗口照常显示，Agent 的操作会实时反映在屏幕上。再加 `KEEL_HEADLESS=1` 则不显示窗口，见 [Agent 端到端测试](../../docs/automation.md)。

```go
window.Open(window.Options{Title: "Hello", Content: page})
window.Main() // 最后一个窗口关闭后退出进程
```

改这里的代码前先读 [架构 · 不能在锁内等待主线程](../../docs/architecture.md#不能在锁内等待主线程)。详见 [窗口与应用](../../docs/app.md)。
