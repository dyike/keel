# Slider

拖动或用键盘选一个范围内的数值。

```go
volume := kit.Slider("音量", 0, 100).Step(5).OnChange(func(v float64) { … })
```

- 拖动滑块，或在轨道上任意位置按下后拖动。
- 键盘：← ↓ 减一步，→ ↑ 加一步，PageUp / PageDown 移动 10 步，Home / End 跳到两端。
- `Step(s)` 让数值对齐到 `min + k·s`；为 0 时连续取值，键盘每次移动范围的 1%。
- `Value()` / `SetValue`（自动对齐和限制在范围内）、`SetRange`、`SetDisabled`。

Agent：角色 `slider`，`value` 为当前数值。

验证：`go run ./examples/components -section slider`，加 `-theme dark` 检查深色。
