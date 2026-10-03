# DescriptionList

`kit.DescriptionList().Item("订单号", "SO-1001").ItemView("状态", view).LabelWidth(96)` 按行显示标签和值，默认标签列 96dp，值列占剩余宽度并换行，顶部对齐。窄容器允许标签列收缩。

纯文本条目向 Agent 暴露一个 text，名字为“标签：值”。ItemView 接受 el.View，每帧调用其 Render，保留子视图的语义和交互，标签单独可读。SetItems 替换全部文本条目，参数为 Description 列表。

组件本身无键盘操作。验证：`go run ./examples/components -section description_list -theme dark`，省略 theme 查看浅色。

`Columns(n)` 设置每行条目列数，最少一列。`Span(n)` 设置最近追加条目的跨度，布局时限制在 1 到当前列数；剩余列放不下时换行，不回填前面行的空位。`Separator()` 插入满行分隔线，后续条目从新行开始。`SetItems` 会清除之前的富内容、跨度和分隔线。

```go
kit.DescriptionList().Columns(2).Vertical().Bordered(true).
    Item("订单号", "SO-123").Item("客户", "张三").
    Separator().Item("备注", "说明跨两列展示").Span(2)
```

`Vertical()` 把每个条目的标签放在值上方；列数和跨度仍然有效。默认横排，`LabelWidth` 仅在横排时生效。

`Bordered(true)` 为各条目增加内边距和主题边框，默认无边框。`Size(sp)` 调整文字大小和间距档，推荐 `theme.TextSm`、`theme.TextBody`、`theme.TextLg`；默认继承文字大小。尺寸忽略非正数和非有限值，标签宽度允许 0，但忽略负数和非有限值。

列数由调用方设置，不会根据窗口宽度自动切换；窄容器内文字换行。小屏可用 `Columns(1)`，富内容自己的固定最小宽度仍须适合容器。多列和跨列保留文本语义，富内容按钮保留键盘、点击和禁用继承。
