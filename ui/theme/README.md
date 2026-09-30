# ui/theme

颜色、字号、字体。组件在每一帧读取这些值，打开第一个窗口前修改即可生效。

```go
import "github.com/dyike/keel/ui/theme"

theme.Primary = theme.RGB(0x16a34a)
```

- **依赖**：只依赖 Gio。
- **被谁依赖**：`layout`、`widget`、`window`。

变量和常量的完整列表见 [组件与布局 · 主题](../../docs/widgets.md#主题)。
