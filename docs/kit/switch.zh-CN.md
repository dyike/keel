# Switch

[English](switch.md) | 简体中文

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

Tooltip 可通过外部组合提供。

滑块位置使用 180ms 平滑过渡（`kit.SwitchDuration`），快速反向切换从当前显示位置衔接。首次显示和减少动画时直接显示目标位置；轨道颜色、Value、回调和 Agent checked 状态立即更新。动画使用帧时钟，不创建定时器；禁用不改变已有值，程序 SetValue 仍可更新并触发位置过渡。

`FocusRing(false)` 隐藏键盘/程序焦点轮廓，true 恢复默认。它不移除焦点或 Tab 停靠点，也不影响 Space/Enter；用于外层已有焦点提示的场景。焦点轮廓画在轨道外圈，不包括标签；鼠标点击沿用原有不显示焦点环的策略。自定义控件要把轮廓画在某个部件上时，可以给可聚焦元素设透明的 `FocusStyle`，再用 `cx.FocusVisible(id)` 判断是否显示。

`TabStop(false)` 从 Tab/Shift+Tab 遍历中跳过开关，仍可鼠标或 `cx.Focus(s.FocusID())` 聚焦并使用键盘。`TabIndex(n)` 按升序排列停靠点，相同值保持树顺序；默认 0，负值跳过。显式设置后，el root 接管 Tab 遍历；没有配置时保持 Gio 原生顺序。禁用、隐藏和当前未绘制的控件跳过；模态/TrapFocus 浮层内独立循环，首次焦点也遵循排序。

```go
notify.TabIndex(2)
secondary.TabStop(false)
```

排序范围是同一个 el root，不能跨多个独立 Embed 或原生 Gio 控件排序。直接调用 Gio Router.MoveFocus 不经过该规则；应用应发送正常 Tab 键事件。编辑器显式处理 Tab 的行为优先于全局遍历。
