# native/screen

[English](README.md) | 简体中文

列出显示器，截取整屏 PNG。

- **依赖**：`native`（错误值）、`native/internal/sys`。
- **单独使用**：可以，不需要窗口。截图需要屏幕录制权限，自己不弹框，可以配合 `native/permission` 申请。

```go
import "github.com/dyike/keel/native/screen"

displays, err := screen.Displays()
png, err := screen.Capture(displays[0].ID) // 最多阻塞 10 秒，不能在主线程调用
```

详见 [原生能力 · screen](../../docs/native.zh-CN.md#screen显示器与截图)。
