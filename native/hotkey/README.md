# native/hotkey

English | [简体中文](README.zh-CN.md)

Register global shortcut keys, which can be triggered when other applications are in the foreground.

- **Dependencies**: `native` (error value), `native/internal/sys`.
- **Use alone**: Yes, but the process must run the macOS main thread event loop before the shortcut key event will be dispatched. When used with `ui`, `ui.Main()` is this loop.

```go
import "github.com/dyike/keel/native/hotkey"

unregister, err := hotkey.Register("cmd+shift+k", func() {
    // Runs in an independent goroutine; to change the interface, use ui.Update
})
```

For details, see [Native capabilities · hotkey](../../docs/native.md#hotkey-global-shortcut-key).
