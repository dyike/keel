# Toggle

[English](toggle.md) | 简体中文

保持按下状态的按钮，比如工具栏里的"加粗"。

```go
bold := kit.Toggle("加粗", false).Icon(kit.IconStar).OnChange(func(on bool) { … })
```

- 点击、Space、Enter 切换；`Value()` / `SetValue(bool)`；`SetDisabled`。

Agent：角色 `toggle`，`selected` 表示是否按下。

验证：`go run ./examples/components -section toggle`，加 `-theme dark` 检查深色。

所有变体都无边框。`ToggleGhost` 和 `ToggleOutline` 未选中时背景透明，`ToggleDefault` 使用 `Subtle` 浅色填充；选中时统一使用 `Highlight`。只有键盘或程序聚焦时显示内嵌焦点环，且不改变布局。

`Size(...)` 接受 `ToggleSizeXSmall`、`ToggleSizeSmall`、`ToggleSizeMedium`、`ToggleSizeLarge`， 同步调整高度、字号、图标和水平留白，高度分别为 24/28/32/40dp，默认 Medium。运行中切换样式保留焦点与选中状态，不触发 OnChange。非法枚举值忽略。
