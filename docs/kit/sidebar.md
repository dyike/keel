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
- 每个可用项都能用 Tab 聚焦，↑ ↓ / Home / End / PageUp / PageDown 移动并滚动到目标，回车或空格激活。`Disabled` 或 `SetItemDisabled(id, bool)` 禁用单项；`SetDisabled` 禁用整个侧栏及插槽操作。
- `SidebarItem.Children` 定义嵌套导航。分支整行点击展开/收起，不触发导航回调；→ 展开，← 收起或移到父项。`SetExpanded` / `Expanded` 管理展开状态，`SetValue` 会展开目标祖先并滚动到选中项。
- `Header(el.View)` / `Footer(el.View)` 位于导航滚动区之外，适合工作区切换和账户信息。插槽可在 Render 中读取 `Collapsed()`，自行切换图标版内容。
- `Height(dp)` 显式设置侧栏高度，导航内容溢出时只滚动中间区域；默认按内容定高，上限为窗口高度。嵌入应用外壳时可传 `cx.ViewportSize()` 返回的高度。
- 底部按钮可以收起侧栏，收起后宽 56dp：只显示图标，名称改由 Tooltip 显示，角标变成圆点。
- `Value()` / `SetValue(id)`、`SetBadge(id, n)`、`Collapsed()` / `SetCollapsed`、`Width(dp)`（默认 220）。

选项会递归复制。整个 Sidebar 的 ID 必须非空且唯一；新增 Section 包含空值或重复 ID 时整组不加入。父项禁用不隐式禁用子项。

Agent：容器角色 `navigation`；叶项是 `link`，分支是带展开布尔值的 `button`，`selected` 表示当前页；收起按钮名为"收起侧栏""展开侧栏"。

验证：`go run ./examples/components -section sidebar`，加 `-theme dark` 检查深色。
