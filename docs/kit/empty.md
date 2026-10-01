# Empty

空列表、搜索无结果时说明当前状态。它不会自动决定列表是否为空，由应用决定何时显示。

```go
state := kit.Empty("暂无订单", "创建订单后将在这里显示。")
return state.Render(cx)
```

`Empty(title, description)` 返回 `*EmptyView`。`SetTitle`、`SetDescription` 更新文字，空字符串不生成对应文本。说明在窄容器里换行；颜色每帧读取主题。纯展示，无点击、键盘、操作按钮。

Agent 角色 `empty`，名称是标题，说明单独可读。验证：`go run ./examples/components -section empty -theme dark`；去掉深色参数检查浅色。
