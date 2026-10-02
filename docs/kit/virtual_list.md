# VirtualList

大量等高行的滚动列表，只构建可视区附近的行，10 万行和 30 行的开销差不多。

```go
logs := kit.VirtualList(len(lines), 24, func(cx *el.Context, i int) el.Element {
    return el.Text(lines[i])
}).Height(300)
```

- `VirtualList` 每行高度相同（`rowHeight`，单位 dp）；自然高度内容使用下方的 `VariableList`。
- 可视区上下各多构建一屏，滚动的那一帧也不会露出空白。
- `Height(dp)` 设置可视高度（默认 320），`Fill()` 改为撑满父容器给的空间。
- `SetCount(n)` 更新行数；`ScrollTo(cx, i)` 以最小滚动量让第 i 行可见。列表还没显示时（比如在另一个标签页），会在第一次显示时再滚动。
- 每行包着一个带稳定 ID 的元素，行在窗口里移动时状态不会丢。

Agent：只列出可视区里的行。

验证：`go run ./examples/components -section virtual_list`，加 `-theme dark` 检查深色。

数据量减少时按新内容高度收回滚动范围；列表清空后再次填充可正常显示。


自然高度内容使用 [VariableList](variable_list.md)，支持稳定 key、行高缓存和阅读位置保持。
