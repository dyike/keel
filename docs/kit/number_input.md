# NumberInput

带 − / + 按钮的数字输入框。

```go
qty := kit.NumberInput("数量").Range(1, 99).Step(1)
price := kit.NumberInput("单价").Range(0, 1e6).Step(0.5).Decimals(2)
```

- 输入过程中允许超出范围的中间状态，比如想输入 15，先打出的 1 可能小于下限；按回车或离开输入框时，限制到范围内并规整格式，无法解析、NaN、无穷大和溢出的文字恢复为原值。
- − / + 按钮按步长调整，到达边界时禁用。步长只用于增减，不把手输值吸附到步长倍数；草稿加一步只发出一次最终值回调。0.1 等十进制步长不会积累二进制加法误差。
- ↑ ↓ 按步长加减，PageUp / PageDown 一次 10 步。
- `Value()` / `SetValue`、`SetDisabled`、`SetError`；`Decimals(n)` 将实际值和显示值一起规整到 0–15 位小数，默认 `-1` 保留精度；非法位数不生效。范围端点优先：例如范围为 0.001–0.009，即使设两位小数，也显示精确边界，避免文字与值不符。

`SetValue` 拒绝非有限值；包含 NaN 或只有无穷大一个端点值的 `Range` 不生效。程序赋值不触发回调。组件或祖先禁用会放弃未提交草稿，恢复已提交值；普通失焦仍提交。

Agent：输入框角色 `textbox`，`value` 是显示的文字；两个按钮名为"减少 标签""增加 标签"。

验证：`go run ./examples/components -section number_input`，加 `-theme dark` 检查深色。

`StepBy(func(value float64, action kit.NumberStepAction) float64)` 根据当前有效草稿及方向计算正步长，方向为 `NumberStepActionIncrement` 或 `NumberStepActionDecrement`。每次按钮／键盘动作调用一次，PageUp／PageDown 将该次步长乘十，不逐步重新求值。返回零、负值或非有限值会取消本次动作并保留草稿；回调不应修改同一个 NumberInput。`StepBy(nil)` 恢复最近一次固定步长，合法的 `Step` 调用会替换动态策略。渲染、程序赋值和单纯输入不调用策略。

`Prefix(view)` / `Suffix(view)` 在文字前后放置货币符号、单位或操作按钮，位于 − / + 按钮内侧；传 nil 移除。动态增删不会改变编辑器身份或内容，子控件继承整体禁用状态。插槽与步进按钮占用固定内容宽度，窄窗口应避免放置过宽的自定义内容。

```go
price.StepBy(func(value float64, action kit.NumberStepAction) float64 {
    if value < 1 || value == 1 && action == kit.NumberStepActionDecrement {
        return 0.1
    }
    return 0.5
})
price.Suffix(kit.Button("帮助", showHelp))
```

应用需要接管增减时，设置 `OnStep(func(kit.NumberStepEvent))`。事件包含有效草稿 `Value`、方向 `Action` 和步数 `Count`（普通按钮／方向键为 1，PageUp／PageDown 为 10）。此模式不自动提交草稿、不调用动态策略、不触发 `OnChange`；回调可用 `SetValue` 更新显示，也可暂不更新。无效草稿使用最近已提交值，文字保持原样；普通回车／失焦提交行为不变。

`OnStep(nil)` 恢复先前固定／动态策略；合法 `Step` 或任何 `StepBy` 调用退出事件模式。范围边界仍禁止向外步进，禁用状态不会派发事件。回调里的 `SetValue` 沿用程序赋值规则，因此不会额外产生 `OnChange`。

`Size(dp)` 同步调整最小高度、文字字号、步进按钮和间距，建议 28／36／48dp；`Size(0)` 恢复主题高度与继承字号，负值和非有限值忽略。显式尺寸按当前主题默认高度缩放，插槽自定义内容仍可指定自己的字号或撑高容器。

`Appearance(false)` 去掉默认背景、边框、圆角和内边距，保留最小高度、标签／错误提示、按钮交互及输入焦点；`Appearance(true)` 恢复。无装饰模式也不显示原边框上的错误／聚焦配色，应用可在外层自行绘制。
