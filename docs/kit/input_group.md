# Input Group

把文本框、图标、单位或操作按钮组合到同一个边框中。

```go
query := kit.Input("").Placeholder("客户或订单号").Clearable()
search := kit.InputGroup("查询订单", query).
    Prefix(kit.Icon(kit.IconSearch)).
    Suffix(kit.Button("查询", func() { lookup(query.Value()) }))
```

Input 实例负责文字、过滤、只读和回调，Input Group 负责边框、标签和前后附加内容。传入的 Input 只通过该组渲染，不能同时在别处渲染。附加内容宽度由自身决定；为长按钮和窄容器预留足够空间。

- `Prefix(el.View)` / `Suffix(el.View)` 放置附加内容，传 nil 移除。动态增删附加内容保留输入焦点和文字。
- `Value` / `SetValue` / `OnChange` 使用内部 Input 的状态；程序赋值不触发回调。格式、密码、清空和提交配置仍放在 Input 上。
- `SetDisabled` 禁用输入及所有附加操作，恢复后保留内容；输入自身的禁用状态也会禁用整组。只读输入仍允许使用附加按钮。
- `SetError` / `Error` / `FocusID` 可接入 Form。组标签同时作为输入的可访问名称；点击可见标签聚焦编辑器。空标签时接受 Form 字段名。
- Tab 依次访问编辑器和可聚焦附加操作，文字编辑快捷键与 Input 相同。

Agent：外层是 `group`，内部编辑器为带字段名称的 `textbox`；附加按钮保留自己的名称、角色和禁用状态。

验证：`go run ./examples/components -section input_group`；加 `-theme dark` 验证深色。

`Addon(id, alignment, view)` 可以在四个位置添加任意数量的附加内容：

| alignment | 位置 |
| --- | --- |
| `InputGroupInlineStart` | 编辑器左侧，在 Prefix 之后 |
| `InputGroupInlineEnd` | 编辑器右侧，在 Suffix 之前 |
| `InputGroupBlockStart` | 整个输入行上方 |
| `InputGroupBlockEnd` | 整个输入行下方 |

同一位置按添加顺序排列。ID 必须非空且保持稳定；同 ID 替换内容和位置，传 nil 删除。上下区域各占一行，可在 View 内用 Row/Wrap 组合计数、工具栏、按钮或菜单。文字、图标及空白处点击聚焦输入；交互控件保留自己的点击和焦点行为。

```go
message := kit.TextArea("").Rows(3)
composer := kit.InputGroup("备注", message).
    Addon("heading", kit.InputGroupBlockStart,
        el.ViewFunc(func(*el.Context) el.Element { return el.Text("填写订单备注") })).
    Addon("send", kit.InputGroupBlockEnd,
        kit.Button("保存", func() { save(message.Value()) }).Size(28))
```

附加按钮直接使用 `kit.Button`，可配置 Variant、Size、Icon、Name、Loading 和禁用状态；紧凑按钮可设 `.Variant(kit.ButtonGhost).Size(24)`。图标按钮需 `Name`。菜单、Popover 和 Tooltip 也可以作为附加 View 组合，沿用各自接口。

TextArea 与单行输入共用这些位置；上下面板随内容增高，文本行数由内部 TextArea 控制。`Rows` 设置最小高度；要随内容增高并设上限，用 TextArea 的 `AutoGrow(minRows, maxRows)`，见 [Input](input.md)。输入和附加内容外只有一层字段边框，聚焦框覆盖整组。只读允许附加动作，输入禁用、组禁用和祖先禁用则禁用全部附加操作。
