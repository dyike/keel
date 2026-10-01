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
