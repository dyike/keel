# ui/layout

摆放组件：`Column`、`Row`、`Card`、`Grow`、`Divider`、`Space`，以及写组件时画圆角边框用的 `Frame`。

- **依赖**：`core`、`theme`。
- **被谁依赖**：`widget`（输入框用 `Frame`）。
- 容器只摆放，不调用用户回调。有回调的放 `widget`。

```go
layout.Column(
    widget.Heading("标题"),
    layout.Row(layout.Grow(widget.Input("")), widget.Button("搜索", search)),
)
```

与 Gio 自带的 `gioui.org/layout` 同名。一个文件里两个都要用时，给 Gio 的起别名：`giolayout "gioui.org/layout"`。

详见 [组件与布局 · 布局容器](../../docs/widgets.md#布局容器)。
