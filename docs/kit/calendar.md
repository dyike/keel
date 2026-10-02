# Calendar

按月显示的日历，可选一天或一个日期范围。

```go
cal := kit.Calendar().Bounds(time.Now(), time.Time{}).DisableDates(isWeekend)
span := kit.Calendar().Range().OnChange(func(start, end time.Time) { … })
```

- 键盘：Tab 进入时落在当前焦点日；方向键按天或按周移动，PageUp / PageDown 按月移动（月末超出目标月份时落在该月最后一天），Home / End 跳到本周首尾，Enter / Space 选中。
- 范围模式：第一次点击确定一端，第二次点击确定另一端，先后顺序不限。
- `Bounds(min, max)` 限制可选范围，零值表示这一侧不限；`DisableDates(fn)` 禁用某些日期。
- `Value()` 返回 `(start, end)`，单日模式下两者相同；`SetValue` 不触发回调，并跳到 start 所在的月份。`SetMonth`、`SetDisabled`。
- 星期名、每周从哪天开始、月份标题、日期格式都来自 `ui/locale`：中文从周一开始，英文从周日开始。

Agent：日历是 `grid`，每一天是 `gridcell`，名字是日期（如 2026-10-08），`selected` 表示已选或在范围内；不可选的日期报告 `disabled`。

验证：`go run ./examples/components -section calendar`，加 `-theme dark` 检查深色。

`Months(n)` 显示 1–12 个连续月份，宽度不足时换行；跨月面板不重复生成日期单元格 ID。点击已显示的后续月份不会跳走首月。点击顶部月份标题进入年月选择器，年份支持直接输入，点击月份完成跳转；月份名称来自 `locale.MonthNames`。

键盘跳过禁用日期；超出 Bounds 时从边界向内寻找可选日期，避免焦点落到禁用单元格后丢失。为避免“全部禁用”的谓词使 UI 无休止查找，每次导航最多检查 366 天，找不到则保持原焦点。Bounds 的反向端点会交换；超出边界的月份翻页按钮禁用。

范围选择使用草稿：第一次点击只设置待定端点，`Value()` 仍返回原先提交的范围；第二次点击完成后才回调。范围内任何一天不可选，都拒绝提交并从本次点击的日期重新起选，同时显示提示。`RangePending()` 查询草稿，`CancelRange()`、取消按钮或焦点日期上的 Esc 放弃草稿，不改变值或触发回调。禁用、修改 Bounds/DisableDates、程序 SetValue 都会取消草稿。

单月网格的标准宽度为 252dp；更窄的容器中七列等宽收缩，星期标题与日期共用列宽。224dp、1× / 2× 下每一天和末列点击都有回归；多月视图仍按月份换行。
