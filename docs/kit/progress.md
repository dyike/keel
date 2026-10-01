# Progress

带标签的进度条。

```go
p := kit.Progress("导入订单")
p.SetValue(0.42)          // 显示 42%
p.SetIndeterminate(true)  // 不知道总量时显示来回滑动的进度块
```

- `SetValue` 会把值限制在 0–1，并退出不确定模式。开启减少动画时，不确定进度静止显示。

Agent：角色 `progressbar`，`value` 是百分比或 `indeterminate`。

验证：`go run ./examples/components -section progress`，加 `-theme dark` 检查深色。
