# Menu

命令菜单，从触发元素旁边弹出。

```go
export := kit.Menu().Item("PDF", "", exportPDF).Item("CSV", "", exportCSV)
more := kit.Menu().
    Item("复制", "mod+c", copy).
    Separator().
    Sub("导出", export).
    Item("删除", "delete", remove)
more.SetItemDisabled("复制", !hasSelection)
more.Trigger(kit.Button("更多", more.Toggle).Variant(kit.ButtonGhost))
```

- `Item(label, shortcut, action)`：`shortcut` 使用 `core.ParseShortcut` 的写法，只用 `kit.Kbd` 显示，**不注册**快捷键；不需要时传空字符串。
- `Sub(label, menu)` 添加子菜单，`Separator()` 添加分隔线，`SetItemDisabled(label, bool)` 禁用或启用菜单项。
- 菜单是模态的：打开时页面的其余部分不响应点击，焦点限制在菜单内，打开后聚焦第一个可用项。点击外部或按 Esc 关闭，关闭后焦点回到触发元素。
- 键盘：
  - ↑ ↓ 在可用项之间移动，跳过禁用项和分隔线，首尾循环；
  - Home / End 跳到第一项或最后一项；
  - Enter / Space 执行当前项；
  - → 打开子菜单，← 或 Esc 只关闭当前这一层子菜单。
- 执行任意一项后，整个菜单（包括所有子菜单）都会关闭，然后再调用 action。
- 子菜单默认显示在右侧，放不下时翻到左侧。
- `Value()` / `SetValue(bool)` 读取或设置是否打开，`Toggle` 用作触发元素的点击回调，`Width(dp)` 设置最小宽度（默认 220）。

Agent：菜单容器的角色是 `menu`（子菜单的名字是它在父菜单中的标题），菜单项的角色是 `menuitem`；有子菜单的项 `value` 为 `submenu`；禁用的项报告 `disabled`。

验证：`go run ./examples/components -section menu`。
