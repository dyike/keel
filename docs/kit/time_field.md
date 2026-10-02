# TimeField

输入一天中的时间。支持整串输入和独立的时、分、秒编辑。

```go
start := kit.TimeField("开始时间").Segmented()
start.SetValue(9*time.Hour + 30*time.Minute)
precise := kit.TimeField("结束时间").Seconds().Hour12(false)
precise.SetValue(17*time.Hour + 45*time.Minute + 30*time.Second)
```

- `Segmented()` 打开分段输入；`Seconds()` 同时启用秒段和秒精度。Tab / Shift+Tab 在各段之间切换，输入后按 Enter 或移出该段提交；非法数字恢复为已提交值。
- 分段模式中，↑ ↓ 调整当前段一个单位，PageUp / PageDown 调整十个单位。时分秒进位联动，越过午夜循环；一次按键只触发一次 `OnChange`。
- 分段格式默认跟随 `locale.Current().Clock12`：中文为 24 小时制，英文为 12 小时制。`Hour12(bool)` 可以显式覆盖。12 小时制显示本地化的上午/下午按钮；切换时段保留分秒，午夜和正午均显示 12。运行时切换格式会丢弃未提交文字，保留值。
- 不调用上述配置时保留整串 `HH:MM` 输入：接受 `9:30`、`0930`、`930`；↑ ↓ 调整分钟，PageUp / PageDown 调整小时。
- `Value()` 返回从午夜开始的 `time.Duration`；默认精确到分钟，`Seconds()` 精确到秒。`SetValue` 对 24 小时取余，不触发回调。
- `SetDisabled`、`SetError`。组件或祖先被禁用时丢弃草稿，恢复后显示已提交值。

Agent：每段是单独的 `textbox`，名字由本地化的“时 / 分 / 秒”和字段标签组成；`FocusID()` 指向小时段。上午/下午是带本地化“时段”名称的按钮。

验证：`go run ./examples/components -section time_field`，加 `-theme dark` 检查深色。
