# VirtualList

大量等尺寸项目的滚动列表，只构建可视区附近的项目；支持纵向行和横向列。

```go
logs := kit.VirtualList(len(lines), 24, func(cx *el.Context, i int) el.Element {
    return el.Text(lines[i])
}).Height(300)
```

- `VirtualList` 每行高度相同（`rowHeight`，单位 dp）；自然高度内容使用下方的 `VariableList`。
- 沿滚动轴在可视区前后各多构建一屏，滚动的那一帧也不会露出空白。
- `Height(dp)` 设置可视高度（默认 320），`Fill()` 改为撑满父容器给的空间。
- `SetCount(n)` 更新行数；`ScrollTo(cx, i)` 以最小滚动量让第 i 行可见。列表还没显示时（比如在另一个标签页），会在第一次显示时再滚动。
- 每行包着一个带稳定 ID 的元素，行在窗口里移动时状态不会丢。

Agent：只列出可视区里的行。

验证：`go run ./examples/components -section virtual_list`，加 `-theme dark` 检查深色。

数据量减少时按新内容高度收回滚动范围；列表清空后再次填充可正常显示。


自然高度内容使用 [VariableList](variable_list.md)，支持稳定 key、行高缓存和阅读位置保持。

等高列表可设置 `ItemKey(func(i int) string)`，在插入、排序时按数据身份保留已构建行的元素状态。key 必须非空且唯一；默认使用索引，回调只在构建范围内调用。


`Horizontal(true)` 改为横向虚拟化，构造时 rowHeight 参数成为每项宽度；false 恢复纵向。`Width(dp)` 设置视口宽度，横向默认 320dp；`Height(dp)` 始终设置高度，横向时是交叉轴尺寸。`Fill()` 沿主轴填满父布局分配的空间，横向时放在 Row 布局内。主轴尺寸的显式设置取消 Fill，交叉轴尺寸设置保留 Fill。

切换方向保留当前首项附近的滚动位置和已构建项目身份，待处理的 ScrollTo 优先；`ScrollToEnd(cx)` 露出最后一项。所有定位接口跟随当前轴，读取精确横向滚动状态使用 `cx.ScrollStateX(list.ID())`，设置偏移使用 `cx.ScrollToX`。数据减少会收回超范围偏移。

```go
cards := kit.VirtualList(len(items), 120, func(cx *el.Context, i int) el.Element {
    return el.Text(items[i].Title)
}).Horizontal(true).Width(480).Height(80)
```

这是单轴虚拟化，不同时虚拟化二维网格。项目内容需适合固定槽位尺寸；切换方向不会自动更改应用内容的布局。
