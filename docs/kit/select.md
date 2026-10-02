# Select

从列表中选一项。

```go
status := kit.Select("状态", "待付款", "已付款", "已发货").OnChange(func(s string) { … })
city := kit.Select("城市", cities...).Searchable()
```

- 点击、Enter、Space 或 ↓ 打开列表。列表中 ↑ ↓ 移动（首尾循环），Home / End 跳到首尾，Enter 选中，Esc 关闭；关闭后焦点回到下拉框。
- `Searchable()` 在列表顶部加搜索框，打开时焦点在搜索框里，回车选第一个匹配项。在搜索框里按 ↓ 进入列表。
- `Hint(s)` 是未选择时显示的文字，默认用 locale 的"请选择"。
- `Value()` / `SetValue`、`SetOptions`（原选择不在新选项里时清空）、`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：下拉框角色 `select`，`value` 为当前选项；打开后列表是 `listbox`，每项是 `option`，`selected` 表示当前选中项。

验证：`go run ./examples/components -section select`，加 `-theme dark` 检查深色。

`SetEntries(...SelectOption)` 支持独立 `Value`、`Label`、`Group` 和 `Disabled`。连续同组前显示标题，空 Label 使用 Value。Value 必须非空且唯一，非法数据在修改前 panic；`Entries()` 返回副本。旧的 `Select(label, options...)` / `SetOptions` 仍以文字作为值。所有入口复制数据，动态替换后清理已移除的选择。`SetOptionDisabled(value, on)` 控制单个选项；点击和键盘会跳过禁用项，程序赋值仍允许它。

`Multiple()` 启用多选，点击、空格或回车切换当前项且保持下拉框打开，Esc 或外部点击关闭。`Values()` 返回按选项顺序排列的独立副本，`SetValues` 替换选择且不触发回调，`OnValuesChange` 接收用户修改后的副本。`Value()` 是主值，多选应使用 `Values()`；`SetValue` 会替换成单个值。显示使用 Label，回调使用 Value。

搜索匹配标签或值。列表虚拟化只构建视口附近的行，打开时滚动到当前选项；方向键绕过标题与禁用项，Home/End 跳到首尾，PageUp/PageDown 翻页。键盘焦点保留在选项容器上，避免长列表回收行后失去焦点；关闭后回到字段。示例包含分组、禁用项和一万条多选选项。
