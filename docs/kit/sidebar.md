# Sidebar

应用的导航侧栏：分组、选中项、角标，可以收起成只显示图标。

```go
nav := kit.Sidebar().
    Section("工作台", kit.SidebarItem{ID: "inbox", Label: "收件箱", Icon: kit.IconInbox, Badge: 12}).
    Section("", kit.SidebarItem{ID: "settings", Label: "设置", Icon: kit.IconUser}).
    OnChange(navigate)
```

- 36dp 导航行、16dp 图标、14sp 文案；选中项使用中性底色和强调图标，计数为低对比数字，底部展开状态显示折叠文案。
- 有标题的分组显示标题；没有标题的分组之间用分隔线隔开。
- 每一项都能用 Tab 聚焦，↑ ↓ 在项之间移动，回车或空格选中。
- 底部按钮可以收起侧栏，收起后宽 56dp：只显示图标，名称改由 Tooltip 显示，角标变成圆点。
- `Value()` / `SetValue(id)`、`SetBadge(id, n)`、`Collapsed()` / `SetCollapsed`、`Width(dp)`（默认 220）。

Agent：容器角色 `navigation`；每一项是 `link`，`selected` 表示当前页；收起按钮名为"收起侧栏""展开侧栏"。

验证：`go run ./examples/components -section sidebar`，加 `-theme dark` 检查深色。
