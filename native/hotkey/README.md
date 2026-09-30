# native/hotkey

注册全局快捷键，其他应用在前台时也能触发。

- **依赖**：`native`（错误值）、`native/internal/sys`。
- **单独使用**：可以，但进程要运行 macOS 主线程事件循环，快捷键事件才会派发。和 `ui` 一起用时，`ui.Main()` 就是这个循环。

```go
import "github.com/dyike/keel/native/hotkey"

unregister, err := hotkey.Register("cmd+shift+k", func() {
    // 在独立 goroutine 里运行；要改界面，用 ui.Update
})
```

详见 [原生能力 · hotkey](../../docs/native.md#hotkey全局快捷键)。
