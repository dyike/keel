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
