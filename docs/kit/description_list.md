# DescriptionList

按顺序显示字段名和值。M1 使用上下排列的单列布局，长文字自然换行；多列布局等待 el grid 能力。

```go
info := kit.DescriptionList(kit.Description{Label: "订单号", Text: "SO-123"})
return info.Render(cx)
```

构造函数返回 `*DescriptionListView`。`SetItems(items...)` 替换条目并复制切片；空列表无条目，空值仍显示字段名。颜色跟随主题，无交互或键盘行为。

Agent 角色 descriptionlist，value 为条目数，每个字段名和值可单独读取。验证：`go run ./examples/components -section description-list -theme dark`，支持 light。
