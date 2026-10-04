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

构造和 `SetRoots` 会深复制节点。更新单个标签可用 `SetNodeLabel`，更新子节点可用 `SetChildren`；修改原始节点不会改变组件。选中和展开状态按 ID 保留，移除节点后清理相应状态；`SetValue` 指定不存在的 ID 会清空选择。nil 节点忽略，空 ID、重复 ID 或循环引用会在改变旧树之前 panic。虚拟行使用节点 ID 保持身份，`SetValue` 会滚动到目标节点。

`TreeNode.Disabled` / `SetNodeDisabled(id, on)` 禁用单个节点的选择、展开、激活和拖动，键盘与范围选择跳过禁用项。禁用不递归传给后代：若父节点已展开，启用的子节点仍可操作；程序赋值和展开允许操作禁用项。

`MultiSelect()` 启用 Ctrl/Cmd 增减节点、Shift 按当前展开顺序连选、Ctrl/Cmd+A 选择当前可见且启用的节点。`SelectedIDs()` 返回包含折叠后代的选区副本，按整棵树的顺序排列；`SetSelectedIDs` 展开所选节点的祖先，不触发 `OnSelectionChange`。移除的 ID 会清理，剩余选区在 `SetRoots` 后保留。

`MoveNode(id, parent, index)` 将整棵子树插入指定父节点的子列表，空 parent 表示根层，index 为最终位置，等于目标原长度表示追加。未知 ID、越界位置、自身或后代目标返回错误，不修改树。移动后展开目标父节点，保留节点 ID、子树和选区。`Roots()` 返回深复制快照，可用于保存更新后的结构。

`Reorderable(func(id, parent string, index int))` 启用拖动排序，松手后才提交，取消拖动不提交；程序 `MoveNode` 不触发回调。

- 拖动时显示落点：行的上半部分插到它前面，下半部分插到后面，前后用一条主色线标出。
- 有子节点（含懒加载）的行分三段：上四分之一插前面，下四分之一插后面，中间**移入**这个目录（追加到末尾），整行高亮。
- 在折叠的目录中间停 0.6 秒，自动展开。
- 拖到列表上下边缘 32dp 内会自动滚动，离边缘越近越快，指针不动也会继续滚，落点跟着变。
- 不能移到自己或自己的子孙里，也不能放到禁用节点或禁用目录里；这些位置不显示落点，松手不提交。
- 移入空目录（没有子节点的叶子）仍需用程序接口或应用提供的命令。


`RenderItem(func(TreeItemContext) el.View)` 自定义展开箭头之后的内容，可加入图标、状态和按钮。上下文包含节点 ID/Label、当前展开行 Index、Depth、Expanded/Selected/Disabled/HasChildren、Loading/Error，以及供 UI 事件调用的 Toggle/Retry。外层保留层级缩进、选择背景和语义；点击箭头只切换展开，子按钮不连带选择或激活行。nil 回调或 nil 内容恢复默认标签；复用有状态子 View，避免在渲染中调用上下文动作。Disabled 表示节点或整树禁用，外层容器禁用由事件系统执行。

`RowHeight(dp)` 设置统一虚拟行高，最小 20dp，0 恢复 28dp；自定义内容需放入这个高度内。`Indent(dp)` 设置每层增量，默认 16dp，0 去掉层级缩进。高度和缩进忽略负数及非有限值。`ScrollTo(cx, id)` 展开祖先并露出目标，不改变选择，也不发送展开/选择通知，未知 ID 返回 false。

`SetChildren(id, children...)` 原子替换一个节点的子列表；输入深复制，允许 nil，空 ID、与其他分支重复的 ID、循环引用和未知父节点返回错误，旧树保持不变。保留仍存在的选择与展开，清理被移除节点；与 SetRoots 不同，局部更新不会重新展开已折叠的选中分支。`Node(id)` 返回节点及其后代的深复制快照，未知 ID 返回 nil。`SetNodeLabel(id, label)` 返回是否找到节点。

`OnExpand(func(id string, expanded bool))` 在用户通过箭头、方向键、双击或自定义 Toggle 改变展开状态后通知；程序 SetExpanded、选中时自动展开祖先和 ScrollTo 不通知。回调可替换数据；组件随后根据当前模型判断是否需要加载。

按需加载：

```go
tr := kit.Tree(&kit.TreeNode{ID: "folder", Label: "目录", Lazy: true})
tr.OnLoad(func(id string, token uint64) {
    go func() {
        children, err := loadChildren(id) // 应用提供数据来源
        core.Update(func() {
            if err != nil {
                tr.SetChildError(id, token, err.Error())
                return
            }
            accepted, applyErr := tr.SetChildResults(id, token, children...)
            // accepted=false 且 applyErr=nil：请求已过期；applyErr 表示节点数据无效。
            _ = accepted
            _ = applyErr
        })
    }()
})
```

Lazy 节点没有子项时仍显示展开箭头。用户或 SetExpanded 打开它时触发 OnLoad，同一节点加载中不重复请求；可同步提交结果，也可通过 core.Update 回传。成功后取消 Lazy 标记，空结果成为叶节点；已有内容在重新加载期间保留。默认行显示加载动画、错误文本和重试按钮，自定义行通过上下文自行展示。

`ReloadNode(id)` 为已展开且可用的节点重试/刷新，返回是否启动请求。`NodeLoading` / `NodeError` 查询状态。折叠该节点、禁用节点/整树、SetRoots、替换 OnLoad 都使旧 token 失效；SetChildren 只使目标子树请求失效，不影响其他分支。重新展开 Lazy 节点可再次请求；替换 OnLoad 后需显式重新打开或 ReloadNode。SetChildResults 返回无效数据错误时保留加载状态，应用可修正数据重交或设置错误。每个 token 只接受一次成功/错误完成。

局部更新目前仍需遍历并验证整棵树，以保证跨分支 ID 唯一；渲染器只调用视口附近的行。Tree 使用统一行高，滚动采用最小露出策略。拖动支持自动滚动和悬停展开。
