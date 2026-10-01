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
- 整个工具栏只占一个 Tab 停靠点，← → 在按钮间移动，首尾循环，Home / End 跳到首尾，回车或空格执行。
- `SetItems` 替换按钮。

Agent：容器角色 `toolbar`，按钮单独列出；"更多"按钮打开的是普通的 `menu`。

验证：`go run ./examples/components -section toolbar`，加 `-theme dark` 检查深色。
