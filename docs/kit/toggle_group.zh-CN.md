# ToggleGroup

[English](toggle_group.md) | 简体中文

一排开关按钮。默认单选，`Multiple()` 后可多选。

```go
align := kit.ToggleGroup("左对齐", "居中", "右对齐")
style := kit.ToggleGroup("B", "I", "U").Multiple().OnChange(func(on []string) { … })
```

- 单选模式下，再点一次已按下的选项会取消选择。
- `Value()` 按选项顺序返回已按下的选项（返回副本），`SetValue(values...)` 不触发回调；单选模式只保留第一个。`SetDisabled`。

Agent：容器角色 `group`，每个按钮是 `toggle`。

验证：`go run ./examples/components -section toggle_group`，加 `-theme dark` 检查深色。

所有变体都无边框。`ToggleGhost` 和 `ToggleOutline` 未选中时背景透明，`ToggleDefault` 使用 `Subtle` 浅色填充；分段组的未选中项同样用 `Subtle` 填充，项之间留 1dp 间隙；选中项使用 `Highlight`。只有键盘或程序聚焦时显示内嵌焦点环。

`Size(...)` 接受 `ToggleSizeXSmall`、`ToggleSizeSmall`、`ToggleSizeMedium`、`ToggleSizeLarge`， 同步调整高度、字号、图标和水平留白，高度分别为 24/28/32/40dp，默认 Medium。运行中切换样式保留焦点与选中状态，不触发 OnChange。非法枚举值忽略。

`Segmented(true)` 将相邻按钮组成一组，默认用 1dp 背景间距分隔，只保留首尾外侧圆角，不绘制外框或接缝。每项仍保留键盘焦点环。单项组保留四个圆角，空组不绘制按钮。

`Gap(dp)` 设置有限非负间距；数值不是分段组默认的 1dp 时，每项恢复四角圆角。`ResetGap()` 恢复分段组默认 1dp、普通组默认 SpaceXs。`Segmented(false)` 恢复普通组；显式 Gap 不因模式切换而丢失。切换模式或间距不改变选中值、焦点，也不触发回调。

逐项配置使用 `Item(value, toggle)`：

```go
g := kit.ToggleGroup("star", "inbox").Multiple().Variant(kit.ToggleOutline).
    Item("star", kit.Toggle("", false).Icon(kit.IconStar)).
    Item("inbox", kit.Toggle("收件箱", false).Icon(kit.IconInbox))
g.SetValue("star")
```

value 仍是构造时的选项值，与显示文案分离。纯图标项的 Agent 名称回退到 value。Item 保存配置快照，源 Toggle 后续修改需再次调用 Item 才生效；源 Toggle 的 Value 和 OnChange 不参与组状态。组回调仍返回选项值，按原选项顺序排列。

显式逐项 Size/Variant 覆盖组配置，包括显式 Medium/Default；未配置则继承组值。逐项禁用与组禁用叠加，不清除已有选中值，也不阻止程序 SetValue。`Item(value, nil)` 清除覆盖，未知 value 忽略。替换配置保留该项身份与焦点。连接分段组建议统一高度，混合尺寸仍按各项高度绘制，不自动拉齐。
