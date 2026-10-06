# DropdownButton

[English](dropdown_button.md) | 简体中文

带下拉菜单的按钮。

```go
kit.DropdownButton("导出", formats)              // 整个按钮打开菜单
kit.DropdownButton("保存", more).Split(save)     // "保存"执行 save，旁边的箭头打开菜单
```

- 菜单是普通的 `kit.Menu`，键盘、子菜单、禁用项的行为都和 Menu 一致。
- `Variant(kit.ButtonSecondary)` 等设置按钮外观，分体样式下两个部分使用同一外观。
- `SetDisabled(true)` 同时禁用按钮和箭头，并关闭已打开的菜单。
- 需要 `el.Root`。

Agent：整体样式下，按钮名字就是标题；分体样式下，主操作按钮名为标题，箭头按钮名为"标题 更多选项"。

验证：`go run ./examples/components -section dropdown_button`。

菜单参数为 nil 时使用空菜单，Split 主操作仍可用。`SetDisabled(true)` 同时禁用主按钮和菜单所在区域；即使外部持有 Menu 并调用 SetValue(true)，禁用的锚点也不会显示弹层。祖先禁用也遵循相同规则。

`Button(button)` 使用已有 Button 配置主操作，并启用分体模式；支持其文字、可访问名称、图标、富内容、配色、紧凑/描边、加载和点击回调。渲染复制配置，不修改原 Button；内部采用组件自己的稳定 ID，忽略传入 Button 的 ID。若同时设置 Split，以 Split 的动作优先。Button(nil) 恢复原普通/Split 用法。

`Size(dp)` 设置两部分高度，0 恢复内层按钮或默认高度；负数和非有限值忽略。不显式设置 Variant/Size 时，两部分继承内层按钮的变体和高度；图标、内容、配色回调等仅属于主操作。

`Loading(bool)` 覆盖主按钮的加载状态。普通模式下不能通过加载中的按钮打开菜单；分体模式下箭头仍可打开菜单。已经打开的菜单不因 Loading 关闭。`SetDisabled(true)` 则始终禁用两部分并关闭菜单。内层 Button 自身禁用仅影响主操作。

```go
kit.DropdownButton("保存", more).
    Button(kit.Button("保存", save).Icon(kit.IconDone).Loading(saving)).
    Size(40)
```

`Placement(side, align)` 设置顶层菜单方向（Top/Bottom/Left/Right）和对齐（Start/Center/End），默认 Bottom/Start；非法组合忽略。`Offset(dp)` 设置间距，默认 4dp，支持 0 和负值重叠，非有限值忽略。打开期间可更新，空间不足时沿用浮层翻转和窗口内限制。子菜单仍按 Right/Start、2dp 展开。

DropdownButton 的这两个方法直接配置传入的 Menu；普通模式锚定整按钮，分体模式锚定箭头。共用同一 Menu 的调用方也会看到配置变化。
