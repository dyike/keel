# Link

[English](link.md) | 简体中文

可点击的文字链接。

```go
kit.Link("查看详情", openDetail)
```

- 主色文字，悬停时变深。可以用 Tab 聚焦，回车或空格触发。
- `SetText`、`SetDisabled`；禁用时变为次要文字色，不响应操作。

Agent：角色 `link`。

验证：`go run ./examples/components -section link`，加 `-theme dark` 检查深色。
