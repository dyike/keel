# RadioGroup

从几个选项中选一个。

```go
pay := kit.RadioGroup("付款方式", "转账", "支票", "现金").OnChange(func(s string) { … })
size := kit.RadioGroup("尺寸", "S", "M", "L").Horizontal()
```

- 键盘行为和原生单选组一致：Tab 进入组时落在当前选中项（没有选中或选中项被禁用时落在第一项可用项），方向键移动并选中，首尾循环。整个组只占一个 Tab 停靠点。
- `SetOptionDisabled(value, bool)` 禁用单项，保留已有选择；方向键跳过禁用项，Home / End 到首末可用项。全禁用时不占 Tab 停靠点。
- `Item(value)` 返回可独立渲染的选项，适用于卡片布局；同一组的所有 Item 共享选择、键盘顺序和禁用状态。每个值在一帧内只渲染一次；外层自行声明 `radiogroup` 角色和名称。未渲染或祖先禁用的项不会成为方向键目标。
- `Size(dp)` 设置组内圆点直径，默认 18dp，正值限制在 12–64dp；`TextSize(sp)` 设置选项继承字号，正值限制在 8–128sp。两者传 0 恢复默认，负值/非有限值忽略；独立 `Item` 使用相同配置。
- `Content(value, view)` 替换对应选项的可见标签，可组合标题、说明、图标等展示内容。原选项字符串仍是值和可访问名称；内容不应包含按钮或输入框。传 nil 恢复文字，未知选项忽略；重排保留内容，移除选项会清理内容。内容每帧渲染，显式字号优先于组的继承字号。
- `Options()` 返回副本；传入选项也会复制，空值和重复值被移除。重排保留选项身份。
- `Value()` 返回选中项，没有选中时为空；`SetValue` 不触发回调；`SetOptions` 替换选项，原选择不在新选项里时清空；`SetDisabled`。

Agent：组的角色是 `radiogroup`，每个选项是 `radio`，`checked` 表示是否选中。

验证：`go run ./examples/components -section radio_group`，加 `-theme dark` 检查深色。

富标签示例：

```go
plans := kit.RadioGroup("套餐", "基础", "专业").Size(28).TextSize(20)
plans.Content("专业", el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Div().Child(el.Text("专业版").Bold(),
        el.Text("支持团队协作").TextSize(theme.TextSm).TextColor(theme.Muted))
}))
```

当前仍使用组内单个 Tab 停靠点，未提供组件级 TabStop/TabIndex。尺寸和字号配置作用于整个组，不提供逐项尺寸覆盖。
