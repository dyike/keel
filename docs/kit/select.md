# Select

从列表中选一项。

```go
status := kit.Select("状态", "待付款", "已付款", "已发货").OnChange(func(s string) { … })
city := kit.Select("城市", cities...).Searchable()
```

- 点击、Enter、Space 或 ↓ 打开列表。列表中 ↑ ↓ 移动（首尾循环），Home / End 跳到首尾，Enter 选中，Esc 关闭；关闭后焦点回到下拉框。
- `Searchable()` 在列表顶部加搜索框，打开时焦点在搜索框里，回车选第一个匹配项。在搜索框里按 ↓ 进入列表。
- `Hint(s)` 是未选择时显示的文字，默认用 locale 的"请选择"。
- `Value()` / `SetValue`、`SetOptions`（原选择不在新选项里时清空）、`SetDisabled`、`SetError`。需要 `el.Root`。

Agent：下拉框角色 `select`，`value` 为当前选项；打开后列表是 `listbox`，每项是 `option`，`selected` 表示当前选中项。

验证：`go run ./examples/components -section select`，加 `-theme dark` 检查深色。

`SetEntries(...SelectOption)` 支持独立 `Value`、`Label`、`Group` 和 `Disabled`。连续同组前显示标题，空 Label 使用 Value。Value 必须非空且唯一，非法数据在修改前 panic；`Entries()` 返回副本。旧的 `Select(label, options...)` / `SetOptions` 仍以文字作为值。所有入口复制数据，动态替换后清理已移除的选择。`SetOptionDisabled(value, on)` 控制单个选项；点击和键盘会跳过禁用项，程序赋值仍允许它。

`Multiple()` 启用多选，点击、空格或回车切换当前项且保持下拉框打开，Esc 或外部点击关闭。`Values()` 返回按选项顺序排列的独立副本，`SetValues` 替换选择且不触发回调，`OnValuesChange` 接收用户修改后的副本。`Value()` 是主值，多选应使用 `Values()`；`SetValue` 会替换成单个值。显示使用 Label，回调使用 Value。

搜索匹配标签或值。列表虚拟化只构建视口附近的行，打开时滚动到当前选项；方向键绕过标题与禁用项，Home/End 跳到首尾，PageUp/PageDown 翻页。键盘焦点保留在选项容器上，避免长列表回收行后失去焦点；关闭后回到字段。示例包含分组、禁用项和一万条多选选项。

## 自定义展示与菜单

`RenderItem(func(*el.Context, SelectItemContext) el.Element)` 自定义可见行内容；上下文含 Entries 索引、选项副本、选中和键盘活动状态。返回 nil 使用默认文字，外层仍负责 option 语义、禁用、勾选和选择。内容应为展示元素，交互操作放在独立控件中。`RenderValue` 接收已选条目的副本，自定义关闭时的展示；占位提示保持默认，返回 nil 恢复标签。Agent 的值继续使用存储值，不跟随自定义绘制文字。

`TitlePrefix("城市：")` 仅在有选择时添加前缀，最长占字段一半宽度并截断。`Empty(view)` 替换无匹配内容，nil 恢复默认；空内容区域可滚动。`Match(func(SelectOption,string) bool)` 替换搜索匹配，收到原始查询；nil 恢复标签/值的包含匹配。闭包依赖的数据变化后重新调用 Match 使缓存失效。

`Clearable(true)` 显示独立清空按钮，清除选择、查询和错误并关闭菜单，焦点回字段；多选回调收到 nil，主值改变时 OnChange 收到空字符串。空选择不重复发回调，程序 SetValue/SetValues 仍静默，禁用继承会阻止按钮。

`MenuWidth(dp)` 设置菜单宽度，0 跟随字段；`MenuMaxHeight(dp)` 设置含搜索/留白的高度预算，0 默认 240dp，正数最小 64dp，窗口可用空间优先。`Size(dp)` 缩放字段、文字、间距和默认行高，推荐 28/36/48，正数最小 20；`RowHeight(dp)` 独立设定统一选项/分组高度，正数最小 16，0 跟随 Size。自定义内容须适配该行高，超出部分由虚拟列表裁剪。上述接口忽略负数和非有限值。

`Appearance(false)` 移除字段背景和边框，保留标签、错误文本、键盘、选项和禁用行为。菜单仍使用主题浮层外观。
