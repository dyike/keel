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

`WithTime()` 为单日期增加分段时间输入，默认精确到分钟；`TimeSeconds()` 开启秒，`TimeHour12(bool)` 覆盖 locale 的 12/24 小时制。`DefaultTime(9*time.Hour)` 设置当前时钟，负值或超过一天按一天折返；这些配置不触发回调。先配置精度，再 SetValue；启用时间前由 Calendar 接收的值只保留日期。

时间模式的 `Value`、`SetValue` 和 `OnChange` 包含时分秒，单日期两端相同，保留日期的时区。选择新日期保留当前时钟并立即回调，弹层保持打开；再选同一天关闭且不重复回调。时间通过 Enter、失焦或方向键提交；Esc/外部关闭保留已提交值并丢弃时间输入草稿。尚无日期时调整时钟不会生成日期或触发回调。清空保留时钟供下次选择使用。

默认显示日期加时间；自定义 Format 是完整布局，需要自行包含时间，例如 `2006-01-02 15:04:05`。范围日历仍只编辑日期，日期边界/禁用规则不限制时刻。日期预设默认保留当前时钟；设置 `DatePickerPreset.IncludeTime: true` 后使用 Start 的时分秒（按组件精度截断，午夜也可显式指定），仍先检查日期边界和禁用规则。纯日期和范围模式也存储这些时刻，通过 DateTimeValue 读取。需要起止时刻时组合两个单日期组件。

预设示例：`kit.DatePickerPreset{ID: "meeting", Label: "下午会议", Start: meetingTime, IncludeTime: true}`。新增 IncludeTime 字段后，使用位置参数构造 DatePickerPreset 的代码需要改成命名字段。范围日历仍只编辑日期，独立保存起止时刻。


`DateValue()` 始终返回日期部分；`SetDateValue(start,end)` 更换日期并保留两端时钟。`DateTimeValue()` / `SetDateTimeValue(start,end)` 存取完整日期时间，支持单日期和范围，不自动打开时间输入、不触发回调。范围的空 End 使用 Start，反向范围连同时刻排序；未配置时间精度时保留纳秒，配置分钟／秒精度后两端按该精度截断。空日期保持零值，清空保留时钟。

兼容行为：范围的 `Value` / `SetValue` 仍只读取／更换日期；单日期 WithTime 的 Value / SetValue 包含时间。用户选择日期、完成范围或选择预设时，OnChange 包含保存的时刻。默认范围文本只显示日期；需要时刻时指定完整 Format。`DefaultTime` 同时设置两端时钟。时间按日期所属时区的本地钟面组合，夏令时不存在／重复时刻沿用 Go time.Date 的处理规则。

`FirstWeekday(time.Monday)` 单独配置弹层日历的周起始日；`ResetFirstWeekday()` 恢复跟随当前 locale，非法值忽略。打开期间可修改，已选日期与范围草稿保留，表头、日期网格、Home/End 和焦点日期的滚动定位同时更新。
