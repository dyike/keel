# Toolbar

一排命令按钮，放不下的按钮会自动移进"更多"菜单。

```go
bar := kit.Toolbar(
    kit.ToolbarItem{Label: "新建", Icon: kit.IconPlus, HasIcon: true, Action: create},
    kit.ToolbarItem{Separator: true},
    kit.ToolbarItem{Label: "导出", Action: export},
)
```

- 每一项可以是按钮或分隔线。`HasIcon` 时显示图标，同时设置 `IconOnly` 则只显示图标；禁用的按钮不响应。
- 溢出判断用的是父容器给的宽度，所以要把工具栏放在有宽度约束的位置，比如整行，或者一个设置了宽度的容器里。父容器按内容定宽时，所有按钮都会显示。
- 命令区域只占一个 Tab 停靠点，← → 在可用按钮和“更多”之间移动，首尾循环，Home / End 跳到首尾，回车或空格执行。“更多”菜单关闭后返回触发按钮；左右附加操作保留各自的 Tab 停靠点。
- `Leading(el.View)` / `Trailing(el.View)` 固定左右区域，中间命令根据剩余宽度溢出。`Size(dp)` 设置统一高度，默认 32，最小 24。
- `SetDisabled` 禁用整个工具栏和附加操作；`SetItemDisabled(index, bool)` 禁用单项。图标按钮可通过悬停或键盘焦点看到提示。
- `SetItems` 替换按钮并关闭旧菜单；传入切片会复制，`Items()` 也返回副本。同名命令的禁用状态独立。

Agent：容器角色 `toolbar`，按钮单独列出；"更多"按钮打开的是普通的 `menu`。

验证：`go run ./examples/components -section toolbar`，加 `-theme dark` 检查深色。

首帧尚未测出命令区宽度时，操作先放进“更多”，测量后自动展开能容纳的按钮。命令区裁剪自己的绘制和命中，避免首次显示或窗口缩放时覆盖右侧插槽；首帧菜单操作有回归测试。
