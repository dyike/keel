# Slider

拖动或用键盘选择数值或双端范围。

```go
volume := kit.Slider("音量", 0, 100).Step(5)
price := kit.RangeSlider("价格区间", 0, 100).Step(5)
price.SetValues(20, 80)
level := kit.RangeSlider("竖向区间", 0, 100).Vertical(180)
```

- 拖动滑块，或在轨道上任意位置按下后拖动。
- 键盘：← ↓ 减一步，→ ↑ 加一步，PageUp / PageDown 移动 10 步，Home / End 跳到两端。
- `Step(s)` 让数值对齐到 `min + k·s`；为 0 时连续取值，键盘每次移动范围的 1%。
- `RangeSlider` 初始选中整个范围。`Values()` / `SetValues(low, high)` 读取或设置两端；程序设置会排序、对齐和钳制，不触发 `OnRangeChange`。用户拖动时选择距离最近的一端，碰到另一端就停止，不交换端点身份。
- 双端各自支持 Tab 焦点和按键；Home / End 受另一端约束。上下限重叠时，在重叠处按下选择上限，可向右或向上重新拉开；下限也可用 Tab 独立操作。
- `Vertical(height)` 设置竖向轨道高度（dp），底部为最小值、顶部为最大值；↑ 增加、↓ 减少。
- `Value()` / `SetValue`（自动对齐和限制在范围内）、`SetRange`、`SetDisabled`。

Agent：单端角色为 `slider`；双端分别是本地化“下限 / 上限 + 标签”命名的 `slider`，`value` 是该端数值。`FocusID()` 在双端模式中指向下限。

验证：`go run ./examples/components -section slider`，加 `-theme dark` 检查深色。
