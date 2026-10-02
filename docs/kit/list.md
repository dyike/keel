# List

单选列表，适合联系人、文件等长列表，内部使用 VirtualList。

```go
contacts := kit.List(names...).Height(240).OnChange(show).OnActivate(open)
```

- 点击或 ↑ ↓ Home End PageUp PageDown 选择，双击或回车激活。
- 焦点在整个列表上，不在某一行上：行会随滚动被回收，焦点没法留在某一行。
- `Value()` 返回选中项的序号（没有时为 -1），`SetValue` 不触发回调；`SetItems` 替换内容，原选择不存在时清空；`Items()`、`SetDisabled`。
- `Height(dp)` 或 `Fill()` 设置高度。

Agent：容器角色 `listbox`，每项是 `option`，`selected` 表示选中。

验证：`go run ./examples/components -section list`，加 `-theme dark` 检查深色。

构造和 `SetItems` 复制选项切片，`Items()` 返回副本。通过 `SetItems` 更新数据；修改传入或返回的切片不影响列表。选择仍以索引表示，替换后索引越界时清空，不触发用户回调。
