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

构造和 `SetRoots` 会深复制节点。更新标签或子节点后应重新调用 `SetRoots`，修改原始节点不会改变组件。选中和展开状态按 ID 保留，移除节点后清理相应状态；`SetValue` 指定不存在的 ID 会清空选择。nil 节点忽略，空 ID、重复 ID 或循环引用会在改变旧树之前 panic。虚拟行使用节点 ID 保持身份，`SetValue` 会滚动到目标节点。
