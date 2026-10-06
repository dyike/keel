# Toggle

[English](toggle.md) | 简体中文

保持按下状态的按钮，比如工具栏里的"加粗"。

```go
bold := kit.Toggle("加粗", false).Icon(kit.IconStar).OnChange(func(on bool) { … })
```

- 点击、Space、Enter 切换；`Value()` / `SetValue(bool)`；`SetDisabled`。

Agent：角色 `toggle`，`selected` 表示是否按下。

验证：`go run ./examples/components -section toggle`，加 `-theme dark` 检查深色。

`Variant(kit.ToggleGhost)` 设置未选中时透明无边框外观；`ToggleOutline` 设置透明背景加边框。`ToggleDefault` 保留原 Surface 背景和边框。选中均显示 Highlight 底色；Ghost 仅在聚焦时出现焦点边框。

`Size(...)` 接受 `ToggleSizeXSmall`、`ToggleSizeSmall`、`ToggleSizeMedium`、`ToggleSizeLarge`， 同步调整高度、字号、图标和水平留白，高度分别为 24/28/32/40dp，默认 Medium。运行中切换样式保留焦点与选中状态，不触发 OnChange。非法枚举值忽略。
