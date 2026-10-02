# Rating

星级评分。

```go
score := kit.Rating("评分", 5).OnChange(func(n int) { … })
avg := kit.Rating("平均", 5).ReadOnly()
avg.SetScore(3.7)
```

- 点击星星评分；获得焦点后 ← → 调整，Home / End 跳到 0 或满分。悬停时预览将要设置的分数。
- `Score()` / `SetScore(float64)` 保留小数评分，用星形的横向填充比例显示半星或任意小数。`ReadOnly()` 下不能点击或用按键改变；编辑模式仍按整星选择。NaN 归零，无穷值钳制到 0 或满分，程序设置不触发回调。
- `Value()` 返回整数部分，兼容已有整数评分；读取平均分使用 `Score()`。
- `Value()` / `SetValue`（限制在 0–max）、`ReadOnly()`、`SetDisabled`。

Agent：角色 `slider`，`value` 为"分数/满分"，如 `3/5`。

验证：`go run ./examples/components -section rating`，加 `-theme dark` 检查深色。
