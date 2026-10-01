# NumberInput

带 − / + 按钮的数字输入框。

```go
qty := kit.NumberInput("数量").Range(1, 99).Step(1)
price := kit.NumberInput("单价").Range(0, 1e6).Step(0.5).Decimals(2)
```

- 输入过程中允许超出范围的中间状态，比如想输入 15，先打出的 1 可能小于下限；按回车或离开输入框时，限制到范围内并规整格式，无法解析的文字恢复为原值。
- − / + 按钮按步长调整，到达边界时禁用。
- ↑ ↓ 按步长加减，PageUp / PageDown 一次 10 步。
- `Value()` / `SetValue`、`SetDisabled`、`SetError`；`Decimals(n)` 固定显示的小数位，默认按需显示。

Agent：输入框角色 `textbox`，`value` 是显示的文字；两个按钮名为"减少 标签""增加 标签"。

验证：`go run ./examples/components -section number_input`，加 `-theme dark` 检查深色。
