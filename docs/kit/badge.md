# Badge

计数或圆点，可以单独显示，也可以挂在子组件右上角。

```go
unread := kit.Badge(3).Child(kit.Button("通知", open))
kit.Badge(1).Dot().Tone(kit.ToneSuccess).Child(avatar)
```

- 计数小于等于 0 时隐藏。超过 `Max`（默认 99）时显示"99+"。
- 挂在子组件上时，角标画在角上，不改变子组件的尺寸和位置，计数变化也不会让布局跳动。
- `Tone` 设置颜色，默认 `ToneDanger`；`Value()` / `SetValue`。

Agent：角色 `badge`，名字是原始计数，`value` 是显示的文字（"99+"）或 `dot`。

验证：`go run ./examples/components -section badge`，加 `-theme dark` 检查深色。
