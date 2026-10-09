# Sidebar

[English](sidebar.md) | 简体中文

应用的导航侧栏：分组、选中项、角标，可以收起成只显示图标。

```go
nav := kit.Sidebar().
    Section("工作台", kit.SidebarItem{ID: "inbox", Label: "收件箱", Icon: kit.IconInbox, Badge: 12}).
    Section("", kit.SidebarItem{ID: "settings", Label: "设置", Icon: kit.IconSettings}).
    OnChange(navigate)
```

- 鼠标点击只显示选中背景；Tab、方向键导航显示焦点框。滚动条占用右侧独立留白，不覆盖行背景、角标或点击区域。
- 36dp 导航行、16dp 图标、14sp 文案；选中项使用中性底色和强调图标，计数为低对比数字，底部展开状态显示折叠文案。
- 有标题的分组显示标题；没有标题的分组之间用分隔线隔开。
- 每个可用项都能用 Tab 聚焦，↑ ↓ / Home / End / PageUp / PageDown 移动并滚动到目标，回车或空格激活。`Disabled` 或 `SetItemDisabled(id, bool)` 禁用单项；`SetDisabled` 禁用整个侧栏及插槽操作。
- `SidebarItem.Children` 定义嵌套导航。分支整行点击展开/收起，不触发导航回调；→ 展开，← 收起或移到父项。`SetExpanded` / `Expanded` 管理展开状态，`SetValue` 会展开目标祖先并滚动到选中项。
- `Header(el.View)` / `Footer(el.View)` 位于导航滚动区之外，适合工作区切换和账户信息。插槽可在 Render 中读取 `Collapsed()`，自行切换图标版内容。
- `Height(dp)` 显式设置侧栏高度，导航内容溢出时只滚动中间区域；默认按内容定高，上限为窗口高度。嵌入应用外壳时可传 `cx.ViewportSize()` 返回的高度。
- 底部按钮可以收起侧栏，收起后宽 56dp：只显示图标，名称改由 Tooltip 显示，角标变成圆点。
- `Icon` 可以不设（`IconNone`）：展开时只显示文字，收起时显示名称首字。
- `SidebarItem.IconView` 接受仅用于显示的自定义图标，在同一个 16dp 插槽内覆盖 `Icon`；收起后仍显示，颜色由应用控制。分支展开箭头紧跟名称，位于角标和独立尾部内容之前。
- `Filter(query)` 只显示名称包含 query 的项（不区分大小写），以及通向它们的父项，父项会临时展开；没有匹配项的分组连同标题一起隐藏。传空字符串恢复全部。选中项被过滤掉时仍保持选中。组件库应用 `go run ./examples/components` 用它做搜索。
- `Side(el.Right)` 将分隔线放到左边，并调整折叠箭头与收起状态的提示方向；默认 `el.Left`。应用仍需把侧栏放在主内容右边，组件不改变父布局顺序。`BorderWidth(dp)` 调整分隔线，0 隐藏。
- `Collapsible(false)` 隐藏内置折叠按钮；程序仍可调用 `SetCollapsed`。
- `SidebarItem.Suffix` / `SetSuffix(id, view)` 添加独立尾部内容，可以放按钮或开关；点击尾部不选择或展开导航项。收起时隐藏尾部内容，展开时最大宽度为侧栏配置宽度的一半，高度应适配 36dp 行。Badge 仍可同时显示。
- `SidebarItem.ContextMenu` / `SetContextMenu(id, menu)` 为条目添加右键菜单，支持 Menu 的子菜单、快捷键提示、链接及自定义内容。一个 Menu 实例只属于一个条目，其 Trigger 不使用；菜单外观与位置仍由 Menu 配置。隐藏、禁用或替换条目菜单会关闭旧菜单。右键不改变选中项。
- `Value()` / `SetValue(id)`、`SetBadge(id, n)`、`Collapsed()` / `SetCollapsed`、`Width(dp)`（默认 220）。

选项结构会递归复制；IconView、Suffix 和 ContextMenu 保留传入实例。整个 Sidebar 的 ID 必须非空且唯一；新增 Section 包含空值或重复 ID 时整组不加入。父项禁用不隐式禁用子项。

Agent：容器角色 `navigation`；叶项是 `link`，分支是带展开布尔值的 `button`，`selected` 表示当前页；收起按钮名为"收起侧栏""展开侧栏"。

验证：`go run ./examples/components -section sidebar`，加 `-theme dark` 检查深色。
