# Stepper

显示多步流程的进度。

```go
steps := kit.Stepper("填写订单", "确认付款", "发货").Navigable()
steps.SetValue(1) // 到第二步
```

- 当前步之前的为已完成（显示对勾），当前步加粗，之后的为未开始。
- `Navigable()` 允许点击已完成的步骤回到那一步，这时调用 `OnChange`。
- `Value()` 返回当前步的序号，等于步骤数时表示全部完成；`SetValue` 不触发回调。

Agent：容器角色 `list`，每一步是 `step`，`value` 为 `done` / `current` / `upcoming`。

验证：`go run ./examples/components -section stepper`，加 `-theme dark` 检查深色。
