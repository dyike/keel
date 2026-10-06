# native/input

English | [简体中文](README.zh-CN.md)

Synthesize mouse and keyboard events: movement, click, and key press.

- **Dependencies**: `native` (error value), `native/internal/sys`.
- **Used alone**: Yes, no window required. Accessibility permissions are required except for reading the mouse position.

```go
import "github.com/dyike/keel/native/input"

input.MouseMove(100, 200)
input.Click(input.Left)
input.Tap("enter")
```

For details, see [Native capabilities · input](../../docs/native.md#input-synthetic-keyboard-and-mouse).
