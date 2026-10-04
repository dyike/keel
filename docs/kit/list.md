# List

默认单选的虚拟列表，支持多选、单项禁用、稳定 ID 和拖动重排。

```go
contacts := kit.List(names...).Height(240).OnChange(show).OnActivate(open)
```

- 点击或 ↑ ↓ Home End PageUp PageDown 选择，双击或回车激活。
- 焦点在整个列表上，不在某一行上：行会随滚动被回收，焦点没法留在某一行。
- `Value()` 返回选中项的序号（没有时为 -1），`SetValue` 不触发回调；`SetItems` 替换内容，原选择不存在时清空；`Items()`、`SetDisabled`。
- `Height(dp)` 或 `Fill()` 设置高度。`Plain()` 去掉边框和背景，用于已经有边框的面板、侧栏里；聚焦时仍显示轮廓。

Agent：容器角色 `listbox`，每项是 `option`，`selected` 表示选中。

验证：`go run ./examples/components -section list`，加 `-theme dark` 检查深色。

构造和 `SetItems` 复制选项切片，`Items()` 返回副本。通过 `SetItems` 更新数据；修改传入或返回的切片不影响列表。选择仍以索引表示，替换后索引越界时清空，不触发用户回调。

`SetEntries(...ListItem)` 接收 `{ID, Label, Disabled}`，复制数据并按 ID 保留选择；空 ID、重复 ID 在修改前 panic。`Entries()` 返回副本。`SetItems` 仍使用索引身份，需要跨插入、删除、重排保持选择时用 `SetEntries`。`SetItemDisabled(index, on)` 禁用单项，鼠标、方向键、范围选择和全选会跳过它，回车也不会激活它。程序选择允许保留禁用项。

`MultiSelect()` 启用 Ctrl/Cmd 加选、Shift 范围选择和 Ctrl/Cmd+A 全选。`SelectedValues()` 返回按显示顺序排列的索引副本，`SetSelectedValues` 程序赋值，`OnSelectionChange` 接收选区副本；`Value()` 是活动项，`SetValue` 替换为单项选择。程序赋值不触发用户回调。

`Reorderable(func(from, to int))` 启用拖动重排，松手后才修改内部顺序并通知，取消拖动不修改顺序。`Move(from, to)` 提供不触发回调的程序重排。两种方式都按 ID 保留选区，回调索引分别指移动前、移动后的位置；当前拖动面向视口内的目标，不自动滚动到远处条目。

## 分组、搜索与自定义内容

`ListItem` 可设置 Group、Keywords 和 Icon。相邻且 Group 相同的条目形成一组，组标题占一行但不可选择；组内没有搜索结果时不显示组头。`RenderGroupHeader(func(cx, group) el.Element)` 替换标题内容，`RenderGroupFooter` 在每个非空组末尾插入一行。所有虚拟行（含组头/页尾）使用统一行高，`RowHeight(dp)` 可设为 20–512dp，默认 32dp。

`Searchable(true)` 显示搜索框，默认不显示；`SetQuery` / `Query` 控制查询。搜索不区分大小写，匹配 Label 和 Keywords，不匹配分组标题。Keywords 在 SetEntries 时复制，Entries 返回的关键词也是副本。`OnSearch` 接收查询变化，程序调用 SetQuery 也会通知，应用可据此获取远端数据；本地过滤始终生效。

搜索不会重编号：Value、选择/激活回调以及自定义行上下文中的 Index 都是 SetEntries 的源数据索引。隐藏选区保留，键盘导航、范围选择和全选只作用于可见可用项，隐藏的当前项不能被 Enter 激活。拖动按包含组头/页尾的可见布局定位目标，再更新源顺序；条目保留原 Group，相邻关系变化可能让同名组分段出现。

```go
list.Searchable(true).RenderItem(func(cx *el.Context, row kit.ListItemContext) el.Element {
    return el.Div().Row().Grow().Child(
        el.Text(row.Item.Label).Grow(),
        kit.Button("详情", func() { open(row.Item.ID) }).Size(24).Render(cx),
    )
})
```

`ListItemContext` 带有条目副本、源 Index、Selected 和 Disabled 状态。RenderItem 返回 nil 使用默认图标/标签。自定义内容的子按钮独立处理事件，不连带选择/激活列表行；行的其余背景仍支持原有选择与拖动。自定义视图的有状态实例应由应用按稳定 ID 保存，内容高度需适配统一行高。

## 加载更多

`OnLoadMore(fn)` 配合 `SetHasMore(true)` 在离底部两行以内请求下一页。组件先设为 loading 再调用 fn，同一份数据不会重复自动请求。结果在 UI 线程或 core.Update 中通过 SetEntries 交付，再 SetLoading(false)；末页设置 SetHasMore(false)。追加数据不会主动把视口拉回旧选择。

失败用 `SetLoadError(message)`，停止自动请求并显示重试按钮；点击后再调用加载函数。禁用时不发起请求，已发出的网络任务由应用取消。远端搜索或加载的结果过期校验也由应用管理，交付前应核对查询/请求标识；组件未提供异步 token API。SetQuery 和 SetEntries 重置本轮请求标记，过滤后不足一屏可继续加载。

## 状态内容

```go
files := kit.List().Searchable(true).
    InitialContent(kit.Label("输入文件名开始搜索")).
    EmptyContent(kit.Label("这个文件夹是空的")).
    NoMatchesContent(kit.Label("没有找到匹配的文件")).
    LoadingContent(kit.Skeleton()).
    ErrorContent(func(msg string, retry func()) el.View { /* 自定义错误和重试 */ })
```

- `EmptyContent`：列表没有数据、也没有搜索词时显示。
- `NoMatchesContent`：有搜索词但没有结果时显示。
- `InitialContent`：可搜索列表还没输入搜索词时，代替整个列表显示，输入后显示结果。
- `LoadingContent`、`ErrorContent`：替换加载更多时的转圈和出错时的文字加"重试"；`retry` 重新发起加载。
- 都传 nil 恢复默认。状态内容显示在列表下方，和默认文字的位置相同。

示例 `go run ./examples/components -section list` 展示搜索、分组、图标和行内操作。自动测试覆盖原有选择/拖动回归，以及搜索源索引、分组页尾、范围跳过隐藏项、子按钮隔离、加载去重/失败重试/禁用；未做真机视觉与拖动验收。
