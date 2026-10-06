# native/screen

English | [简体中文](README.zh-CN.md)

List monitors, capture full screen PNG.

- **Dependencies**: `native` (error value), `native/internal/sys`.
- **Used alone**: Yes, no window required. Taking screenshots requires screen recording permission. If you don’t want to pop up the screen yourself, you can apply with `native/permission`.

```go
import "github.com/dyike/keel/native/screen"

displays, err := screen.Displays()
png, err := screen.Capture(displays[0].ID) // Blocks for up to 10 seconds and cannot be called from the main thread
```

For details, see [Native capabilities · screen](../../docs/native.md#screen-monitor-and-screenshots).
