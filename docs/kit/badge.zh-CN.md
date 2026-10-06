# Badge

[English](badge.md) | 简体中文

数字、圆点或图标徽标，可单独显示或挂在子组件角上。

```go
unread := kit.Badge(3).Child(kit.Button("通知", open))
kit.Badge(1).Dot().Tone(kit.ToneSuccess).Child(avatar)
```

- 数字和圆点模式下，计数小于等于 0 时隐藏。超过 `Max`（默认 99）时显示"99+"。
- 挂在子组件上时，角标画在角上，不改变子组件的尺寸和位置，计数变化也不会让布局跳动。
- `Icon(IconName)` 切换为图标模式，与计数无关，即使 count 为 0 也显示。`Icon(IconNone)` 恢复数字模式；Dot 切换为圆点并清除图标。图标模式位于右下角，带 Surface 色边框；数字和圆点位于右上角。
- `Size(dp)` 设置数字/图标高度，默认 18，接受 12–128；推荐 12/18/24。圆点按 8/18 比例缩放，数字超长时宽度随文字增长。非法值忽略。
- `Color(color.NRGBA)` 指定底色，前景自动按底色选择黑白。`Tone` 设置主题语义颜色（默认 ToneDanger），同时清除 Color 覆盖。固定自定义色不自动随主题变化。
- `Name(string)` 设置可访问名称，适合图标状态，例如“已验证”；空字符串恢复原始计数名称。`Value()` / `SetValue` 操作计数，不改变图标模式。

Agent：角色 `badge`，默认名字是原始计数，可用 Name 覆盖；`value` 是显示数字（"99+"）、`dot` 或 `icon`。

验证：`go run ./examples/components -section badge`，加 `-theme dark` 检查深色。
