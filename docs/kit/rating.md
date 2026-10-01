# Rating

星级评分。

```go
score := kit.Rating("评分", 5).OnChange(func(n int) { … })
avg := kit.Rating("平均", 5).ReadOnly()
```

- 点击星星评分；获得焦点后 ← → 调整，Home / End 跳到 0 或满分。悬停时预览将要设置的分数。
- `Value()` / `SetValue`（限制在 0–max）、`ReadOnly()`、`SetDisabled`。

Agent：角色 `slider`，`value` 为"分数/满分"，如 `3/5`。

验证：`go run ./examples/components -section rating`，加 `-theme dark` 检查深色。
