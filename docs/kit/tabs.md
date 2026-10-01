# Tabs

标签页。

```go
tabs := kit.Tabs().Add("基本", basicForm).Add("通知", notifySettings).OnChange(onTab)
```

- 只渲染当前页。每一页都是应用自己持有的 View，切换后状态还在。
- Tab 键聚焦到当前标签，← → 切换，首尾循环，Home / End 跳到首尾。
- `Value()` / `SetValue(i)`（不触发回调）。

Agent：标签栏是 `tablist`，每个标签是 `tab`，`selected` 表示当前标签；当前页是以标签标题为名的 `tabpanel`。

验证：`go run ./examples/components -section tabs`，加 `-theme dark` 检查深色。
