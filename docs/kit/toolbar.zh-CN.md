# Toolbar

[English](toolbar.md) | 简体中文

一排命令按钮，放不下的按钮会自动移进"更多"菜单。

```go
bar := kit.Toolbar(
    kit.ToolbarItem{Label: "新建", Icon: kit.IconPlus, Action: create},
    kit.ToolbarItem{Separator: true},
    kit.ToolbarItem{Label: "导出", Action: export},
)
```

- 每一项可以是按钮或分隔线。设置 `Icon` 时显示图标，同时设置 `IconOnly` 则只显示图标，名字作为提示；禁用的按钮不响应。
- 溢出判断用的是父容器给的宽度，所以要把工具栏放在有宽度约束的位置，比如整行，或者一个设置了宽度的容器里。父容器按内容定宽时，所有按钮都会显示。
- 命令区域只占一个 Tab 停靠点，← → 在可用按钮和“更多”之间移动，首尾循环，Home / End 跳到首尾，回车或空格执行。“更多”菜单关闭后返回触发按钮；左右附加操作保留各自的 Tab 停靠点。
- `Leading(el.View)` / `Trailing(el.View)` 固定左右区域，中间命令根据剩余宽度溢出。`Size(dp)` 设置统一高度，默认 32，最小 24。
- `SetDisabled` 禁用整个工具栏和附加操作；`SetItemDisabled(index, bool)` 禁用单项。图标按钮可通过悬停或键盘焦点看到提示。
- `SetItems` 替换按钮并关闭旧菜单；传入切片会复制，`Items()` 也返回副本。同名命令的禁用状态独立。

Agent：容器角色 `toolbar`，按钮单独列出；"更多"按钮打开的是普通的 `menu`。

验证：`go run ./examples/components -section toolbar`，加 `-theme dark` 检查深色。

首帧尚未测出命令区宽度时，操作先放进“更多”，测量后自动展开能容纳的按钮。命令区裁剪自己的绘制和命中，避免首次显示或窗口缩放时覆盖右侧插槽；首帧菜单操作有回归测试。

## 任意位置的自定义组

`ToolbarItem{Label: "缩放", Content: zoomSelect, Width: 150}` 在该位置放入交互式视图，视图内部可以组合多个控件。`Content` 优先于 Action/Icon，Separator 仍优先；Label 用于组语义和溢出菜单。`Width` 指组宽度，合法范围为大于 0 且不超过 4096dp，未设置或非法时使用 160dp。

放不下的组以 Label 进入“更多”，点击后在工具栏下方的模态浮层展示 Content。可用 `OverflowContent` 提供另一种布局；浮层限制在窗口内，内容过高时滚动，Esc 或点击外部关闭。组重新放得下、被禁用或 SetItems 替换时关闭浮层。自定义视图的状态由应用持有；在工具栏与浮层间切换会改变元素树路径，输入选区等框架内部状态不保证保留。

自定义组的子控件保留各自的 Tab 停靠点和方向键行为，工具栏的左右/Home/End 导航仍只管理命令按钮。整组与整个工具栏的禁用会传递给子控件。示例在搜索与复制之间加入缩放选择器；自动测试覆盖双倍率溢出、嵌套 Select、Esc、禁用、宽度变化与替换清理；原生视觉未验收。
