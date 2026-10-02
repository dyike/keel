# List

默认单选的虚拟列表，支持多选、单项禁用、稳定 ID 和拖动重排。

```go
contacts := kit.List(names...).Height(240).OnChange(show).OnActivate(open)
```

- 点击或 ↑ ↓ Home End PageUp PageDown 选择，双击或回车激活。
- 焦点在整个列表上，不在某一行上：行会随滚动被回收，焦点没法留在某一行。
- `Value()` 返回选中项的序号（没有时为 -1），`SetValue` 不触发回调；`SetItems` 替换内容，原选择不存在时清空；`Items()`、`SetDisabled`。
- `Height(dp)` 或 `Fill()` 设置高度。`Plain()` 去掉边框和背景，用于已经有边框的面板、侧栏里；聚焦时仍显示轮廓。

Agent：容器角色 `listbox`，每项是 `option`，`selected` 表示选中。

验证：`go run ./examples/components -section list`，加 `-theme dark` 检查深色。

构造和 `SetItems` 复制选项切片，`Items()` 返回副本。通过 `SetItems` 更新数据；修改传入或返回的切片不影响列表。选择仍以索引表示，替换后索引越界时清空，不触发用户回调。

`SetEntries(...ListItem)` 接收 `{ID, Label, Disabled}`，复制数据并按 ID 保留选择；空 ID、重复 ID 在修改前 panic。`Entries()` 返回副本。`SetItems` 仍使用索引身份，需要跨插入、删除、重排保持选择时用 `SetEntries`。`SetItemDisabled(index, on)` 禁用单项，鼠标、方向键、范围选择和全选会跳过它，回车也不会激活它。程序选择允许保留禁用项。

`MultiSelect()` 启用 Ctrl/Cmd 加选、Shift 范围选择和 Ctrl/Cmd+A 全选。`SelectedValues()` 返回按显示顺序排列的索引副本，`SetSelectedValues` 程序赋值，`OnSelectionChange` 接收选区副本；`Value()` 是活动项，`SetValue` 替换为单项选择。程序赋值不触发用户回调。

`Reorderable(func(from, to int))` 启用拖动重排，松手后才修改内部顺序并通知，取消拖动不修改顺序。`Move(from, to)` 提供不触发回调的程序重排。两种方式都按 ID 保留选区，回调索引分别指移动前、移动后的位置；当前拖动面向视口内的目标，不自动滚动到远处条目。
