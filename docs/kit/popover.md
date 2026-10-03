# Popover

在触发元素旁边显示一块非模态面板，用于筛选条件、简短表单这类内容。

```go
filters := kit.Popover(form).Width(280).Offset(12)
filters.Trigger(kit.Button("筛选", filters.Toggle).Variant(kit.ButtonSecondary))
```

- 触发元素自己负责打开：把 `Toggle` 传给它的点击回调。Popover 只在触发元素外面包一层不可交互的锚点，不会多出 Tab 停靠点，键盘行为完全由触发元素决定。
- 以下操作会关闭面板：再次点击触发元素、在面板和触发元素之外按下鼠标、按 Esc。面板外的那次点击会继续传给下面的元素。
- 面板不会移动焦点。如果内容里有输入框，用户需要自己点进去或用 Tab 进入。
- `Value()` 返回是否打开；`SetValue(bool)` 用程序打开或关闭，不触发回调；`OnChange(fn)` 只在用户操作时调用。
- `Placement(side, align)` 设置面板相对触发元素的位置，默认 `el.Bottom, el.Start`；放不下时自动翻到对侧。
- 需要 `el.Root`（见 [el · 浮层](../el.md#浮层e4--e5)）。

Agent：面板的角色是 `dialog`，里面的元素单独列出；页面其余部分仍然可见。

验证：`go run ./examples/components -section popover`，加 `-theme dark` 检查深色。

面板宽高受窗口约束，长内容可滚动。`Width(0)` 恢复按内容宽度；负数和非有限宽度会被忽略。`SetDisabled(true)` 关闭面板并禁用触发区域，程序打开和 Toggle 也不会绕过禁用。父容器禁用或锚点消失时，浮层会关闭。

内容中可以继续放 Menu、Select 等浮层组件。父面板先登记，内部浮层显示在上方；Esc 从最里面逐层关闭。示例的“选择预设”可验证这条路径。

`Offset(dp)` 设置触发元素与面板的间距，默认 4dp；0 表示贴合，负数允许重叠，NaN/无穷值忽略。打开时修改会在下一帧重新定位，不触发 OnChange。靠近窗口边缘仍会翻转或限制位置，实际间距可能受可用空间约束。

`Appearance(false)` 移除默认背景、边框、圆角、阴影和内边距；默认开启。定位、滚动和关闭行为保留。

```go
filters.Appearance(false).PanelStyle(func(panel *el.DivEl) {
    panel.Bg(theme.Surface).Border(1, theme.Primary).Rounded(theme.RadiusMd).P(12)
})
```

`PanelStyle` 每帧在默认外观和 Width 之后调用，nil 移除自定义样式；不要保留元素引用。面板 ID、dialog 角色、窗口尺寸上限和滚动由组件最后设置。打开期间修改外观保留内容状态和焦点；回调中读取主题颜色可随主题切换更新。

`RightClick(true)` 让触发区域的右键按下直接调用 Toggle，再次右键关闭；false 恢复默认行为。左键和键盘仍执行触发元素自己的回调，不额外增加 Tab 停靠点。触发元素不要再注册调用 Toggle 的右键回调，以免执行两次。禁用自身或父容器时右键也不会打开面板。

```go
info := kit.Popover(details).RightClick(true)
// 左键和键盘也能打开，作为右键操作的替代入口。
info.Trigger(kit.Button("详情", info.Toggle))
```

只希望鼠标右键打开时，可为触发按钮提供不同的左键回调；键盘替代入口由应用提供。其他鼠标按键使用 MouseButton 配置。

`MouseButton(pointer.ButtonPrimary/Secondary/Tertiary)` 选择左键、右键或中键按下时自动切换面板。默认 0，由触发元素自己打开；`MouseButton(0)` 恢复该模式。非法值忽略，同时按下多个按键不触发。`RightClick(true/false)` 分别等同于选择右键/恢复手动模式，最后一次设置生效。

```go
info.MouseButton(pointer.ButtonTertiary) // gioui.org/io/pointer
```

监听不会吞掉触发元素自身的事件。选择左键自动打开时，不要再将 Toggle 绑定到按钮左键回调，否则会在按下和释放时各切换一次。自动监听不增加 Tab 停靠点，键盘替代入口仍由触发元素提供。

`Arrow(true)` 显示指向触发器的箭头，默认关闭。箭头深 6dp，面板相应增加 6dp 间距，Offset 测量到尖端。方向跟随实际定位翻转，沿面板边缘按 Start/Center/End 对齐并避开圆角；窄面板会缩小箭头。窗口边缘可能裁剪箭头。箭头独立于内容滚动，点击箭头不会关闭面板。

箭头使用 PanelStyle 设置的纯色背景，未设置时取当前主题 Surface；不单独绘制阴影或边框，也不采样渐变背景。Appearance(false) 不会自动关闭箭头，可独立调用 Arrow(false)。
