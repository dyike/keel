# Tabs

标签页。

```go
tabs := kit.Tabs().Add("基本", basicForm).Add("通知", notifySettings).OnChange(onTab)
```

- 只渲染当前页。每一页都是应用自己持有的 View，切换后状态还在。
- Tab 键聚焦到当前标签，← → 切换，首尾循环，Home / End 跳到首尾。
- 标签放不下时，多出来的收进末尾的"更多"菜单，从菜单选中的标签会作为菜单按钮的文字。和 Toolbar 一样，标签页要放在有宽度约束的位置。
- `Closable(fn)` 给每个标签加关闭按钮，`fn(i)` 决定关闭的含义，通常直接传 `tabs.Remove`。关闭按钮在标签旁边而不在标签里面，所以点击它不会先切换到这个标签。
- `Value()` / `SetValue(i)`（不触发回调）、`Remove(i)`、`Len()`。

Agent：标签栏是 `tablist`，每个标签是 `tab`，`selected` 表示当前标签；当前页是以标签标题为名的 `tabpanel`。

验证：`go run ./examples/components -section tabs`，加 `-theme dark` 检查深色。
