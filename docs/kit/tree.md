# Tree

可展开、折叠的树，只构建可见的行。

```go
tree := kit.Tree(&kit.TreeNode{ID: "ui", Label: "ui", Children: []*kit.TreeNode{
    {ID: "kit", Label: "kit"},
}}).Height(260).OnChange(open)
```

- 键盘：
  - ↑ ↓ 移动，Home / End 跳到首尾；
  - → 展开当前节点，已展开时进入第一个子节点；
  - ← 折叠当前节点，已折叠时回到父节点；
  - 回车激活。
- 点击箭头展开或折叠；双击节点时展开或折叠，并激活它。
- 节点 ID 在整棵树内必须唯一。`Value()` 返回选中节点的 ID；`SetValue(id)` 会展开它的所有祖先，保证它可见，不触发回调。`Expanded`、`SetExpanded`、`SetRoots`、`SetDisabled`。

Agent：容器角色 `tree`，每个节点是 `treeitem`，`value` 为 expanded / collapsed（没有子节点时为空），`selected` 表示选中。

验证：`go run ./examples/components -section tree`，加 `-theme dark` 检查深色。
