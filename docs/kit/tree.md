# Tree

English | [简体中文](tree.zh-CN.md)

Expandable, collapsed tree, only building visible rows.

```go
tree := kit.Tree(&kit.TreeNode{ID: "ui", Label: "ui", Children: []*kit.TreeNode{
    {ID: "kit", Label: "kit"},
}}).Height(260).OnChange(open)
```

- keyboard:
  - ↑ ↓ Move, Home / End jump to the beginning and end;
  - → Expand the current node and enter the first child node when expanded;
  - ← Collapse the current node, and return to the parent node when collapsed;
  - Press Enter to activate.
- Clicking the arrows expands or collapses; double-clicking a node expands or collapses it and activates it.
- Node IDs must be unique within the entire tree. `Value()` returns the ID of the selected node; `SetValue(id)` will expand all its ancestors to ensure that it is visible and does not trigger a callback. `Expanded`, `SetExpanded`, `SetRoots`, `SetDisabled`. `Plain()` removes the border and background, the usage is the same as List.

Agent: container role `tree`, each node is `treeitem`, `value` is expanded / collapsed (empty when there are no child nodes), `selected` means selected.

Verify: `go run ./examples/components -section tree`, add `-theme dark` to check the dark theme.

Construct and `SetRoots` make a deep copy of the node. `SetNodeLabel` can be used to update a single label, and `SetChildren` can be used to update child nodes; modifying the original node will not change the component. The selected and expanded state is retained by ID, and the corresponding state is cleared after removing the node; `SetValue` specifying a non-existent ID will clear the selection. nil nodes are ignored, and empty IDs, duplicate IDs, or circular references will panic before changing the old tree. The dummy row uses the node ID to maintain identity, and `SetValue` scrolls to the target node.

`TreeNode.Disabled` / `SetNodeDisabled(id, on)` Disable selection, expansion, activation and dragging of individual nodes, keyboard and range selection skip disabled items. Disables are not passed recursively to descendants: if the parent node is expanded, enabled child nodes are still operable; procedural assignment and expansion allow operation of disabled items.

`MultiSelect()` enables Ctrl/Cmd to add and subtract nodes, Shift to select consecutively in the current expansion order, and Ctrl/Cmd+A to select currently visible and enabled nodes. `SelectedIDs()` returns a copy of the selection containing the collapsed descendants, in the order of the entire tree; `SetSelectedIDs` expands the ancestors of the selected node, without triggering `OnSelectionChange`. Removed IDs are cleared and the remaining selections remain after `SetRoots`.

`MoveNode(id, parent, index)` inserts the entire subtree into the child list of the specified parent node. The empty parent indicates the root level, and index is the final position. It is equal to the original length of the target to indicate appending. Unknown ID, out-of-bounds position, self or descendant target returns an error and does not modify the tree. Expand the target parent node after moving, retaining the node ID, subtree, and selection. `Roots()` returns a deep copy snapshot that can be used to save the updated structure.

`Reorderable(func(id, parent string, index int))` enables drag sorting and submits after releasing the button. It does not submit when dragging is cancelled. The program `MoveNode` does not trigger the callback.

- The drop point is shown when dragging: the top half of the row is inserted in front of it, the bottom half is inserted behind it, and the front and back are marked with a main color line.
- Lines with child nodes (including lazy loading) are divided into three sections: the upper quarter is inserted in the front, the lower quarter is inserted in the back, the middle is moved into this directory (appended to the end), and the entire line is highlighted.
- Pause in the middle of the collapsed directory for 0.6 seconds and automatically expand.
- Dragging it within 32dp of the upper and lower edges of the list will automatically scroll. The closer you are to the edge, the faster it will be. If the pointer does not move, it will continue to scroll, and the landing point will change accordingly.
- It cannot be moved to itself or its descendants, nor can it be placed in a disabled node or disabled directory; the drop point will not be displayed in these locations, and it will not be submitted if you let go.
- Moving into an empty directory (a leaf with no child nodes) still requires the use of commands provided by the program interface or application.


`RenderItem(func(TreeItemContext) el.View)` Customize the content after the expansion arrow and add icons, statuses and buttons. The context contains the node ID/Label, the current expanded row Index, Depth, Expanded/Selected/Disabled/HasChildren, Loading/Error, and Toggle/Retry for UI events to call. The outer layer retains the hierarchical indentation, selection background and semantics; clicking the arrow only switches to expansion, and the sub-button does not select or activate the row. nil callback or nil content restores the default label; reuses stateful sub-views to avoid calling contextual actions in rendering. Disabled means that the node or the entire tree is disabled, and the outer container is disabled by the event system.

`RowHeight(dp)` sets a unified virtual line height, minimum 20dp, 0 returns to 28dp; custom content needs to be placed within this height. `Indent(dp)` sets the increment of each layer, the default is 16dp, 0 removes the level indent. Height and indent ignore negative and non-finite values. `ScrollTo(cx, id)` Expands the ancestor and exposes the target, does not change the selection, and does not send expansion/selection notifications. Unknown ID returns false.

`SetChildren(id, children...)` Atomic replacement of a node's sublist; enters a deep copy, allowing nil, empty IDs, duplicate IDs from other branches, circular references, and unknown parent nodes to return an error, and the old tree remains unchanged. Preserve existing selections and expansions, clean up removed nodes; unlike SetRoots, local updates will not re-expand collapsed selected branches. `Node(id)` Returns a deep copy snapshot of the node and its descendants, or nil for unknown IDs. `SetNodeLabel(id, label)` returns whether the node is found.

`OnExpand(func(id string, expanded bool))` Notifies after the user changes the expansion state through arrows, arrow keys, double-click, or custom Toggle; the program SetExpanded, automatically expand ancestors when selected, and ScrollTo do not notify. The callback replaces the data; the component then determines whether it needs to be loaded based on the current model.

Load on demand:

```go
tr := kit.Tree(&kit.TreeNode{ID: "folder", Label: "Directory", Lazy: true})
tr.OnLoad(func(id string, token uint64) {
    go func() {
        children, err := loadChildren(id) // Application provides data source
        core.Update(func() {
            if err != nil {
                tr.SetChildError(id, token, err.Error())
                return
            }
            accepted, applyErr := tr.SetChildResults(id, token, children...)
            // accepted=false and applyErr=nil: the request has expired; applyErr means the node data is invalid.
            _ = accepted
            _ = applyErr
        })
    }()
})
```

Lazy nodes still display expand arrows when they have no children. OnLoad is triggered when the user or SetExpanded opens it, and there are no repeated requests during the loading of the same node; the results can be submitted synchronously or returned through core.Update. After success, the Lazy mark is cancelled, and the empty result becomes a leaf node; existing content is retained during reloading. The default row displays the loading animation, error text, and retry button, and the custom row displays itself through context.

`ReloadNode(id)` Retry/refresh for expanded and available nodes, returning whether to start the request. `NodeLoading` / `NodeError` query status. Collapse of the node, disabling the node/whole tree, SetRoots, and replacing OnLoad all invalidate the old token; SetChildren only invalidates the target subtree request and does not affect other branches. Re-expand the Lazy node to request again; after replacing OnLoad, you need to explicitly reopen or ReloadNode. SetChildResults retains the loading state when returning an invalid data error, and the application can correct data re-crossing or setting errors. Only one successful/error completion is accepted per token.

Local updates currently still require traversing and validating the entire tree to ensure unique IDs across branches; the renderer only calls rows near the viewport. Tree uses a uniform row height and scrolls using a minimum exposure strategy. Dragging supports automatic scrolling and hover expansion.
