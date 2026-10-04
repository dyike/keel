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
- `Size(kit.TimeFieldSizeXSmall / TimeFieldSizeSmall / TimeFieldSizeMedium / TimeFieldSizeLarge)` 同步调整框体、字号、时钟图标、分段宽度及内边距，适用于整串与分段模式。默认 Medium 使用主题控件高度；其余三档最小高度为 24、28、40dp。标签和错误文字保留表单标准字号；切换尺寸不提交草稿或改变值。整串与分段的双倍率尺寸验证已通过。
- `SetDisabled`、`SetError`。组件或祖先被禁用时丢弃草稿，恢复后显示已提交值。

Agent：每段是单独的 `textbox`，名字由本地化的“时 / 分 / 秒”和字段标签组成；`FocusID()` 指向小时段。上午/下午是带本地化“时段”名称的按钮。

验证：`go run ./examples/components -section time_field`，加 `-theme dark` 检查深色。

`SegmentKeys(true)` 启用快速分段编辑，并自动打开分段模式；传 false 恢复原有草稿编辑。与默认模式相比：

- 获得焦点时选中整段，左右箭头提交当前草稿并移到相邻段；首段左移、末段右移不会离开控件，Tab 仍可离开。
- 输入或粘贴有效的两位数字立即提交，并聚焦下一段；12 小时制最后一个数字段之后聚焦时段按钮。无效两位数不会自动提交或跳段。
- ↑ ↓、PageUp / PageDown 只在当前段范围内循环，不进位到其他段。12 小时段为 1–12，24 小时段为 0–23，分秒为 0–59。
- Backspace / Delete 将当前段重置为 0；12 小时段重置为 12，并保留当前上午/下午。
- 12 小时制的 a / p 设为上午/下午；时段按钮支持左右切段、上下切换，删除回到上午。

已验证连续输入、粘贴跳段、独立循环、删除重置、时段快捷键、禁用及双倍率尺寸；默认模式仍保留 Enter/移焦提交和进位行为。
