# 状态按钮

`Toggle(...)` 返回 `*widget.ToggleView`。

```go
pin := widget.Toggle("固定", false).Icon(widget.Icon(widget.IconCheck)).Size(widget.Small).OnChange(onChange)
pin.SetValue(true)
pin.SetDisabled(true)
```
鼠标点击与聚焦后的 Space/Enter 切换状态。`Value` 读取，`SetValue` 静默更新；禁用后不响应输入。`Ghost` 移除未选中时的常驻边框，选中和聚焦状态仍可辨认。自动化快照提供 `toggle` 角色、`selected` 和 `disabled`。

验证入口：`go run ./examples/components -section toggle`。
