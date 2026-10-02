# DatePicker

点击后弹出日历的日期字段。

```go
due := kit.DatePicker("交货日期").Placeholder("选择日期")
trip := kit.DatePicker("出差日期").Range().Months(2)
```

- 点击、Enter 或 ↓ 打开日历，焦点落在已选日期，没有选择时落在今天；目标日期禁用时跳过禁用日期。选好日期（或范围选完两端）后自动关闭；Esc 或点击外部关闭且不改变选择。
- 范围的第一端只作为草稿；Esc、点击外部、禁用字段或祖先容器都会丢弃草稿，保留已提交值，不触发 `OnChange`。重新打开从已提交日期开始，年月选择面板也会重置。
- `Months(n)` 请求显示 1–12 个月。窄窗口自动减少并排月份，仍可用月份箭头访问其余月份；高度不足时弹层可滚动，键盘切换日期会滚动到目标格子。
- `Value()` / `SetValue(start, end)`，`OnChange(func(start, end time.Time))`；`Bounds`、`DisableDates` 与 Calendar 相同；`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：字段是 `button`，名字是标签，`value` 是显示的日期；打开后是名为标签的 `dialog`，里面是 Calendar 的 `grid`。

验证：`go run ./examples/components -section date_picker`，加 `-theme dark` 检查深色。
