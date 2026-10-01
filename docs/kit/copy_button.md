# CopyButton

把文字复制到剪贴板，并显示"已复制"作为反馈。

```go
kit.CopyButton(func() string { return order.ID })
```

- 文字在点击时才读取，所以函数总能拿到最新内容。
- 复制后，按钮显示"已复制"和对勾图标，`kit.CopiedFeedback`（1.5 秒）后恢复为"复制"。
- 按钮是 Ghost 样式，高 28dp，支持 Tab 聚焦和 Space / Enter 触发。

Agent：角色是 `button`，名字在"复制"和"已复制"之间切换。

验证：`go run ./examples/components -section copy_button`。
