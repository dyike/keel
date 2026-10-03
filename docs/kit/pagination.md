# Pagination

分页条。

```go
pages := kit.Pagination(total, 20).OnChange(func(p int) { load(p) })
start, end := pages.Bounds() // 当前页的数据范围 [start, end)
```

- 页码从 1 开始。页数多时，显示首页、末页和当前页附近的页码，中间用省略号代替。
- 第一页时"上一页"禁用，最后一页时"下一页"禁用。
- `Value()` / `SetValue(p)`（会限制在合法范围内，不触发回调）、`SetTotal(n)`、`Pages()`。
- "共 N 条""上一页""下一页"来自 locale。

Agent：容器角色 `navigation`，`value` 为"当前页/总页数"；页码是名为数字的按钮。

验证：`go run ./examples/components -section pagination`，加 `-theme dark` 检查深色。

窄容器下总数、页码与前后页按钮自动换行，末页按钮仍可点击。页数和最后一页范围的计算避免整数加法溢出。

`Compact(true)` 只显示前后页图标按钮，隐藏总数和页码；Agent 仍报告当前页/总页数。false 恢复完整布局，切换不改变当前页。

`VisiblePages(n)` 设置数字按钮数量上限，不包含省略号和前后页。正数限制在 3–101，以便始终保留首页、当前页和末页；中间窗口随当前页移动。0 恢复原有窗口策略，负数忽略。默认沿用 Keel 原布局（少于八页时全显示），与上游默认五个按钮不同。

`Size(dp)` 设置按钮高度，正数限制在 16–128dp，0 恢复默认 28dp；负数和非有限值忽略。使用连续尺寸，可选 20/24/28/36dp 对应不同密度。`SetDisabled(true)` 禁用整条分页；祖先禁用也会生效。程序调用 SetValue/SetTotal 仍更新状态且不触发 OnChange。

```go
pages.VisiblePages(9).Size(36)
compact := kit.Pagination(total, 20).Compact(true).Size(24)
```

页码按钮使用稳定身份，窗口移动和紧凑模式切换时，仍存在的按钮保留键盘焦点。总页数缩小时当前页会收敛到末页。
