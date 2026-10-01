# Toggle

保持按下状态的按钮，比如工具栏里的"加粗"。

```go
bold := kit.Toggle("加粗", false).Icon(kit.IconStar).OnChange(func(on bool) { … })
```

- 点击、Space、Enter 切换；`Value()` / `SetValue(bool)`；`SetDisabled`。

Agent：角色 `toggle`，`selected` 表示是否按下。

验证：`go run ./examples/components -section toggle`，加 `-theme dark` 检查深色。
