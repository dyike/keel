# Tabs

支持下划线、胶囊、描边和分段外观的标签页。

```go
tabs := kit.Tabs().Add("基本", basicForm).Add("通知", notifySettings).OnChange(onTab)
```

- `Variant(TabsUnderline/TabsPill/TabsOutline/TabsSegmented)` 选择外观，默认保留原下划线样式。颜色随主题切换。
- `AddItem(TabItem{Title, Page, Icon, Content, Disabled})` 添加图标或富标签。Content 替换可见标题，应为展示元素；Title 仍用于 Agent 名称和溢出菜单。富标签在有限宽度内布局，文字截断需内容自行设置 MaxLines。
- `SetItem(i, item)` 更新条目并保留稳定身份；`SetItemDisabled(i, bool)` 单项禁用。禁用项不能点击、关闭或拖动，方向键及溢出菜单跳过它。禁用当前项自动选择下一个可用项；全部禁用时保留当前页，标签不进入 Tab 导航，页面内容仍保留可用。重新启用一个条目后恢复选择。程序修改不触发 OnChange；SetValue 不选择禁用项。
- 只渲染当前页。每一页都是应用自己持有的 View，切换后状态还在。
- Tab 键聚焦到当前标签，← → 切换，首尾循环，Home / End 跳到首尾。
- 标签放不下时，多出来的收进末尾的"更多"菜单，当前选中的标签始终保留在栏内，可继续用方向键操作；长标题会截断。和 Toolbar 一样，标签页要放在有宽度约束的位置。
- `Closable(fn)` 给每个标签加关闭按钮，`fn(i)` 决定关闭的含义，通常直接传 `tabs.Remove`。关闭按钮在标签旁边而不在标签里面，所以点击它不会先切换到这个标签。
- `Reorderable(fn)` 允许拖动可见标签重排，放开指针后提交；取消拖动不改变顺序。`Move(from, to)` 可程序重排任意标签，不触发回调；页面与编辑状态跟随稳定身份，选中页保持不变。拖动不自动打开溢出菜单。
- 标签聚焦时 Delete 调用关闭回调。关闭当前页后，焦点移到剩余的相邻标签；关闭其他页保持当前页。
- `Leading(el.View)` / `Trailing(el.View)` 放置固定的左右区域；`Size(dp)` 设置标签高度，默认 40，最小 24。`SetDisabled` 禁用标签、关闭按钮、拖动和当前页。
- `Value()` / `SetValue(i)`（不触发回调）、`Remove(i)`、`Len()`。

Agent：标签栏是 `tablist`，每个标签是 `tab`，`selected` 表示当前标签；当前页是以标签标题为名的 `tabpanel`。

验证：`go run ./examples/components -section tabs`，加 `-theme dark` 检查深色。
