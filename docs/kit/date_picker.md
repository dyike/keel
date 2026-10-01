# DatePicker

点击后弹出日历的日期字段。

```go
due := kit.DatePicker("交货日期").Placeholder("选择日期")
trip := kit.DatePicker("出差日期").Range()
```

- 点击、Enter 或 ↓ 打开日历，焦点落在已选日期，没有选择时落在今天。选好日期（或范围选完两端）后自动关闭；Esc 或点击外部关闭且不改变选择。
- `Value()` / `SetValue(start, end)`，`OnChange(func(start, end time.Time))`；`Bounds`、`DisableDates` 与 Calendar 相同；`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：字段是 `button`，名字是标签，`value` 是显示的日期；打开后是名为标签的 `dialog`，里面是 Calendar 的 `grid`。

验证：`go run ./examples/components -section date_picker`，加 `-theme dark` 检查深色。
