# RadioGroup

从几个选项中选一个。

```go
pay := kit.RadioGroup("付款方式", "转账", "支票", "现金").OnChange(func(s string) { … })
size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
```

- 键盘行为和原生单选组一致：Tab 进入组时落在当前选中项（没有选中时落在第一项），方向键移动并选中，首尾循环。整个组只占一个 Tab 停靠点。
- `Value()` 返回选中项，没有选中时为空；`SetValue` 不触发回调；`SetOptions` 替换选项，原选择不在新选项里时清空；`SetDisabled`。

Agent：组的角色是 `radiogroup`，每个选项是 `radio`，`checked` 表示是否选中。

验证：`go run ./examples/components -section radio_group`，加 `-theme dark` 检查深色。
