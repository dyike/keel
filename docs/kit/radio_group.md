# RadioGroup

从几个选项中选一个。

```go
pay := kit.RadioGroup("付款方式", "转账", "支票", "现金").OnChange(func(s string) { … })
size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
```

- 键盘行为和原生单选组一致：Tab 进入组时落在当前选中项（没有选中或选中项被禁用时落在第一项可用项），方向键移动并选中，首尾循环。整个组只占一个 Tab 停靠点。
- `SetOptionDisabled(value, bool)` 禁用单项，保留已有选择；方向键跳过禁用项，Home / End 到首末可用项。全禁用时不占 Tab 停靠点。
- `Item(value)` 返回可独立渲染的选项，适用于卡片布局；同一组的所有 Item 共享选择、键盘顺序和禁用状态。每个值在一帧内只渲染一次；外层自行声明 `radiogroup` 角色和名称。未渲染或祖先禁用的项不会成为方向键目标。
- `Options()` 返回副本；传入选项也会复制，空值和重复值被移除。重排保留选项身份。
- `Value()` 返回选中项，没有选中时为空；`SetValue` 不触发回调；`SetOptions` 替换选项，原选择不在新选项里时清空；`SetDisabled`。

Agent：组的角色是 `radiogroup`，每个选项是 `radio`，`checked` 表示是否选中。

验证：`go run ./examples/components -section radio_group`，加 `-theme dark` 检查深色。
