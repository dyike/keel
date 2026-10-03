# Switch

立即生效的开关，比如"接收通知"。

```go
notify := kit.Switch("接收通知", true).OnChange(func(on bool) { save(on) })
```

- 点击、Space、Enter 切换；`Value()` / `SetValue(bool)`，`SetValue` 不触发回调；`SetDisabled`。

Agent：角色 `switch`，`checked` 表示开关状态。

验证：`go run ./examples/components -section switch`，加 `-theme dark` 检查深色。

`Size(kit.SwitchSmall)` 使用 28×16dp 轨道，默认 `SwitchMedium` 为 36×20dp；滑块分别为 12/16dp。标签字号继续继承父元素。`LabelSide(el.Left/el.Right)` 设置标签位置，默认右侧；标签和轨道共用一个可点击、可聚焦的控件，切换样式保留焦点。非法枚举值忽略。

`Color(color.NRGBA)` 只覆盖选中轨道，`ClearColor()` 恢复主题色。自身禁用且选中时，自定义色的 alpha 减半；未设置自定义色时保留原禁用配色。祖先禁用沿用 el 的禁用表现。主题颜色应在 Render 时传入，以跟随主题切换。

```go
notify.Size(kit.SwitchSmall).LabelSide(el.Left).Color(theme.Success)
```

滑块当前立即切换位置；动画、独立焦点环开关和 Tab 顺序配置仍待补齐。Tooltip 可通过外部组合提供。
