# ToggleGroup

一排开关按钮。默认单选，`Multiple()` 后可多选。

```go
align := kit.ToggleGroup("左对齐", "居中", "右对齐")
style := kit.ToggleGroup("B", "I", "U").Multiple().OnChange(func(on []string) { … })
```

- 单选模式下，再点一次已按下的选项会取消选择。
- `Value()` 按选项顺序返回已按下的选项（返回副本），`SetValue(values...)` 不触发回调；单选模式只保留第一个。`SetDisabled`。

Agent：容器角色 `group`，每个按钮是 `toggle`。

验证：`go run ./examples/components -section toggle_group`，加 `-theme dark` 检查深色。

`Variant(kit.ToggleGhost)` 设置未选中时透明无边框外观；`ToggleOutline` 设置透明背景加边框。`ToggleDefault` 保留原 Surface 背景和边框。选中均显示 Highlight 底色；Ghost 仅在聚焦时出现焦点边框。

`Size(...)` 接受 `ToggleSizeXSmall`、`ToggleSizeSmall`、`ToggleSizeMedium`、`ToggleSizeLarge`， 同步调整高度、字号、图标和水平留白，高度分别为 24/28/32/40dp，默认 Medium。运行中切换样式保留焦点与选中状态，不触发 OnChange。非法枚举值忽略。

`Segmented(true)` 连接相邻按钮，默认间距为 0，首尾外侧保留圆角，中间接缝使用单个描边；Ghost 保留无描边外观和每项焦点轮廓。单项组保留四个圆角，空组不绘制按钮。

`Gap(dp)` 设置有限非负间距；分段组设正值后各项恢复四角圆角。`ResetGap()` 恢复分段组 0、普通组 SpaceXs 的默认间距。`Segmented(false)` 恢复普通组；显式 Gap 不因模式切换而丢失。切换模式/间距不会改变选中值或焦点，不触发回调。默认仍为单选，可调用 Multiple 开启多选。
