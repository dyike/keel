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
- 节点 ID 在整棵树内必须唯一。`Value()` 返回选中节点的 ID；`SetValue(id)` 会展开它的所有祖先，保证它可见，不触发回调。`Expanded`、`SetExpanded`、`SetRoots`、`SetDisabled`。`Plain()` 去掉边框和背景，用法同 List。

Agent：容器角色 `tree`，每个节点是 `treeitem`，`value` 为 expanded / collapsed（没有子节点时为空），`selected` 表示选中。

验证：`go run ./examples/components -section tree`，加 `-theme dark` 检查深色。

构造和 `SetRoots` 会深复制节点。更新标签或子节点后应重新调用 `SetRoots`，修改原始节点不会改变组件。选中和展开状态按 ID 保留，移除节点后清理相应状态；`SetValue` 指定不存在的 ID 会清空选择。nil 节点忽略，空 ID、重复 ID 或循环引用会在改变旧树之前 panic。虚拟行使用节点 ID 保持身份，`SetValue` 会滚动到目标节点。

`TreeNode.Disabled` / `SetNodeDisabled(id, on)` 禁用单个节点的选择、展开、激活和拖动，键盘与范围选择跳过禁用项。禁用不递归传给后代：若父节点已展开，启用的子节点仍可操作；程序赋值和展开允许操作禁用项。

`MultiSelect()` 启用 Ctrl/Cmd 增减节点、Shift 按当前展开顺序连选、Ctrl/Cmd+A 选择当前可见且启用的节点。`SelectedIDs()` 返回包含折叠后代的选区副本，按整棵树的顺序排列；`SetSelectedIDs` 展开所选节点的祖先，不触发 `OnSelectionChange`。移除的 ID 会清理，剩余选区在 `SetRoots` 后保留。

`MoveNode(id, parent, index)` 将整棵子树插入指定父节点的子列表，空 parent 表示根层，index 为最终位置，等于目标原长度表示追加。未知 ID、越界位置、自身或后代目标返回错误，不修改树。移动后展开目标父节点，保留节点 ID、子树和选区。`Roots()` 返回深复制快照，可用于保存更新后的结构。

`Reorderable(func(id, parent string, index int))` 启用拖动到可见节点前后的位置，松手后才提交；跨层目标采用目标节点所在的父层，取消拖动不提交。程序 `MoveNode` 不触发回调。拖动不自动滚动或展开折叠目录；移入空目录可使用程序接口或应用提供的命令。
