# native/permission

检查、申请 macOS 隐私权限：辅助功能、屏幕录制、输入监控。

- **依赖**：`native`（错误值）、`native/internal/sys`。
- **单独使用**：可以，不需要窗口。

```go
import "github.com/dyike/keel/native/permission"

ok, err := permission.Granted(permission.ScreenRecording) // 只查
ok, err = permission.Request(permission.Accessibility)    // 可能弹授权框
```

详见 [原生能力 · permission](../../docs/native.md#permission权限)。
