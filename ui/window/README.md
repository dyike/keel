# ui/window

窗口：打开、关闭、置前、窗口快捷键、事件循环、离屏截图。

| 文件 | 内容 |
| --- | --- |
| `window.go` | `Open`、`Main`、`Options`、`Window` |
| `shortcut.go` | 快捷键解析与分发 |
| `root.go` | 窗口根视图：背景、滚动、24dp 边距 |
| `screenshot.go` | `Screenshot` 离屏渲染成 PNG |
| `testdata/raise` | 真实窗口死锁回归测试 |

- **依赖**：`core`、`theme`。不依赖 `widget` 和 `layout`：窗口只认 `core.Widget` 接口。
- **被谁依赖**：应用代码。

```go
window.Open(window.Options{Title: "Hello", Content: page})
window.Main() // 最后一个窗口关闭后退出进程
```

改这里的代码前先读 [架构 · 不能在锁内等待主线程](../../docs/architecture.md#不能在锁内等待主线程)。详见 [窗口与应用](../../docs/app.md)。
