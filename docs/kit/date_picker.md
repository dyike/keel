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

`Format(layout)` 使用 Go 的时间布局同时格式化单日期和范围两端，例如 `Format("2006-01-02")`；传空字符串恢复当前 locale 的日期格式。只改变显示，不修改日期、不触发回调，格式语法与 GPUI 的 chrono 格式不同。

`Clearable(true)` 在有日期时显示独立清空按钮；鼠标或键盘清空会关闭弹层、丢弃范围草稿、移除错误，并回调一次 `OnChange(time.Time{}, time.Time{})`，焦点返回日期触发器。清空不会误打开日历，禁用状态继承到清空按钮；`Clearable(false)` 隐藏按钮，程序 SetValue 不触发回调。

`Presets(...DatePickerPreset)` 在日历下方显示快捷按钮，窄布局自动换行并随弹层滚动。每项包含稳定 `ID`、`Label`、`Start`、`End`；单日期模式忽略 End，范围模式的空 End 表示同一天，反向范围自动排序。空 ID/标签被忽略，重复 ID 保留首项；组件复制输入切片，`Presets()` 清除全部预设。

预设遵守 Bounds 和 DisableDates，范围内部也不能跨越禁用日期；不可用项保持显示但禁用，点击时重新检查最新配置。选择预设会取消范围草稿、清除错误、关闭弹层并回调一次，焦点返回字段。自定义禁用函数按日期逐日检查，应保持快速且无副作用。日期是快照，“今天”“最近七天”等移动预设由应用更新。

```go
trip.Presets(kit.DatePickerPreset{
    ID: "week", Label: "最近七天",
    Start: today.AddDate(0, 0, -6), End: today,
})
```

`Size(dp)` 同步缩放字段最小高度、字号、日历图标、清空按钮和间距，建议 28／36／48dp；`Size(0)` 恢复主题高度与继承字号。负值和非有限值忽略。日历网格与预设按钮不随字段缩放。

`Appearance(false)` 去掉字段背景、边框、圆角和内边距，保留最小高度、标签、错误文案与键盘交互，弹层仍保留原装饰；`Appearance(true)` 恢复。无装饰字段不显示默认边框上的焦点／错误配色，可在外层自行绘制。打开弹层期间切换这些配置不会清除日期或关闭弹层。
