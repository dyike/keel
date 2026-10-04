# Stepper

显示多步流程的进度。

```go
steps := kit.Stepper("填写订单", "确认付款", "发货").Navigable()
steps.SetValue(1) // 到第二步
```

- 当前步之前的为已完成（显示对勾），当前步加粗，之后的为未开始。
- `Navigable()` 允许点击已完成的步骤回到那一步，这时调用 `OnChange`。
- `Value()` 返回当前步的序号，等于步骤数时表示全部完成；`SetValue` 不触发回调。

Agent：容器角色 `list`，每一步是 `step`，`value` 为 `done` / `current` / `upcoming`。

验证：`go run ./examples/components -section stepper`，加 `-theme dark` 检查深色。

步骤多于可见宽度时提供横向滚动，保留步骤顺序和连接线。可导航的已完成步骤支持键盘聚焦；`SetDisabled` 禁止滚动和步骤修改。构造时复制标签切片。末项滚动后点击、禁用与 1× / 2× 有回归测试。

`Vertical()` 改为竖向排列与竖向滚动，连接线沿标记圆心排列。`Size(dp)` 设置标记直径，常用 20、24（默认）、32dp；图标与连接线长度跟随缩放，文字使用主题字号档。零、负数、NaN、无穷值不改变当前尺寸。

需要图标、单项禁用或多行说明时使用 `SetEntries`：

```go
steps := kit.Stepper().Vertical().Size(32).Navigable()
steps.SetEntries(
    kit.StepperItem{Label: "订单", Icon: kit.IconReceipt},
    kit.StepperItem{Label: "付款", Icon: kit.IconLock, Disabled: true},
    kit.StepperItem{Label: "发货", Icon: kit.IconInbox},
)
steps.SetValue(2)
```

- `StepperItem.Icon` 替换数字/完成对勾；未提供图标时保留原有标记。
- `StepperItem.Disabled` 或 `SetItemDisabled(index, on)` 禁止该步点击和键盘聚焦，不改变其完成状态，不阻止 `SetValue` 程序跳转。越界索引忽略。
- `StepperItem.Content` 可提供多行说明等展示内容；`Label` 仍是 Agent 名称。内容不应嵌套按钮或输入框。
- `SetEntries` 和 `Entries()` 都复制条目切片，内容 View 不深拷贝。替换条目时保留当前索引并按新长度收敛，不触发 `OnChange`。
- 自定义条目遵守配置的导航范围；单项、整个组件与祖先禁用均阻止导航。

`TextCenter(true)` 使横向步骤的文字/富内容位于标记下方并居中，连接线沿相邻标记圆心排列。各列平分可用宽度，最小宽度为标记直径的三倍，不足时横向滚动；自定义内容可以自行设定内部排版。竖向模式保留标记在内容左边，只设置文字居中。false 恢复原布局。`Horizontal()` 可从竖向切回横向。

`Navigation(kit.StepperNavigationNone / StepperNavigationCompleted / StepperNavigationAll)` 分别为只读、只可返回已完成步骤、可选择任意未禁用步骤。默认 None，`Navigable()` 等同于 Completed。All 包含尚未开始的步骤；点击当前步骤不触发 OnChange。配置变化及 SetValue 均不触发回调。

已验证三种导航策略、当前项重复点击、Tab 跳过禁用项、动态方向切换、双倍率等宽列及窄窗口滚动；浅深色窗口像素回归确认连接线随完成状态着色。
