# native/input

合成鼠标、键盘事件：移动、点击、按键。

- **依赖**：`native`（错误值）、`native/internal/sys`。
- **单独使用**：可以，不需要窗口。除读取鼠标位置外都需要辅助功能权限。

```go
import "github.com/dyike/keel/native/input"

input.MouseMove(100, 200)
input.Click(input.Left)
input.Tap("enter")
```

详见 [原生能力 · input](../../docs/native.md#input合成键鼠)。
