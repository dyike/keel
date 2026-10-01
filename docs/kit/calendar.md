# Calendar

按月显示的日历，可选一天或一个日期范围。

```go
cal := kit.Calendar().Bounds(time.Now(), time.Time{}).DisableDates(isWeekend)
span := kit.Calendar().Range().OnChange(func(start, end time.Time) { … })
```

- 键盘：Tab 进入时落在当前焦点日；方向键按天或按周移动，PageUp / PageDown 按月移动，Home / End 跳到本周首尾，Enter / Space 选中。
- 范围模式：第一次点击确定一端，第二次点击确定另一端，先后顺序不限。
- `Bounds(min, max)` 限制可选范围，零值表示这一侧不限；`DisableDates(fn)` 禁用某些日期。
- `Value()` 返回 `(start, end)`，单日模式下两者相同；`SetValue` 不触发回调，并跳到 start 所在的月份。`SetMonth`、`SetDisabled`。
- 星期名、每周从哪天开始、月份标题、日期格式都来自 `ui/locale`：中文从周一开始，英文从周日开始。

Agent：日历是 `grid`，每一天是 `gridcell`，名字是日期（如 2026-10-08），`selected` 表示已选或在范围内；不可选的日期报告 `disabled`。

验证：`go run ./examples/components -section calendar`，加 `-theme dark` 检查深色。
