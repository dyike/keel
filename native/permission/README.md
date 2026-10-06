# native/permission

English | [简体中文](README.zh-CN.md)

Check and apply for macOS privacy permissions: accessibility, screen recording, input monitoring.

- **Dependencies**: `native` (error value), `native/internal/sys`.
- **Used alone**: Yes, no window required.

```go
import "github.com/dyike/keel/native/permission"

ok, err := permission.Granted(permission.ScreenRecording) // Check only
ok, err = permission.Request(permission.Accessibility)    // An authorization box may pop up
```

See [Native capabilities · permission](../../docs/native.md#permission-permission) for details.
