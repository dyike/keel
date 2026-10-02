# NumberInput

带 − / + 按钮的数字输入框。

```go
qty := kit.NumberInput("数量").Range(1, 99).Step(1)
price := kit.NumberInput("单价").Range(0, 1e6).Step(0.5).Decimals(2)
```

- 输入过程中允许超出范围的中间状态，比如想输入 15，先打出的 1 可能小于下限；按回车或离开输入框时，限制到范围内并规整格式，无法解析、NaN、无穷大和溢出的文字恢复为原值。
- − / + 按钮按步长调整，到达边界时禁用。步长只用于增减，不把手输值吸附到步长倍数；草稿加一步只发出一次最终值回调。0.1 等十进制步长不会积累二进制加法误差。
- ↑ ↓ 按步长加减，PageUp / PageDown 一次 10 步。
- `Value()` / `SetValue`、`SetDisabled`、`SetError`；`Decimals(n)` 将实际值和显示值一起规整到 0–15 位小数，默认 `-1` 保留精度；非法位数不生效。范围端点优先：例如范围为 0.001–0.009，即使设两位小数，也显示精确边界，避免文字与值不符。

`SetValue` 拒绝非有限值；包含 NaN 或只有无穷大一个端点值的 `Range` 不生效。程序赋值不触发回调。组件或祖先禁用会放弃未提交草稿，恢复已提交值；普通失焦仍提交。

Agent：输入框角色 `textbox`，`value` 是显示的文字；两个按钮名为"减少 标签""增加 标签"。

验证：`go run ./examples/components -section number_input`，加 `-theme dark` 检查深色。
