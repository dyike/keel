# Select

从列表中选一项。

```go
status := kit.Select("状态", "待付款", "已付款", "已发货").OnChange(func(s string) { … })
city := kit.Select("城市", cities...).Searchable()
```

- 点击、Enter、Space 或 ↓ 打开列表。列表中 ↑ ↓ 移动（首尾循环），Home / End 跳到首尾，Enter 选中，Esc 关闭；关闭后焦点回到下拉框。
- `Searchable()` 在列表顶部加搜索框，打开时焦点在搜索框里，回车选第一个匹配项。搜索框里方向键用于移动光标，要用方向键选择时，先按 Tab 进入列表。
- `Hint(s)` 是未选择时显示的文字，默认用 locale 的"请选择"。
- `Value()` / `SetValue`、`SetOptions`（原选择不在新选项里时清空）、`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：下拉框角色 `select`，`value` 为当前选项；打开后列表是 `listbox`，每项是 `option`，`selected` 表示当前选中项。

验证：`go run ./examples/components -section select`，加 `-theme dark` 检查深色。
