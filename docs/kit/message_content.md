# MessageContent

在同一条 Message 正文中混排多个气泡、附件和普通 View。直接传入的 Bubble 会继承消息对齐；任一直接气泡为 Ghost 时，头尾自动取消缩进。HeaderInset/FooterInset 的显式设置仍优先。

```go
content := kit.MessageContent(
    kit.Bubble(kit.Label("导出完成")),
    kit.Attachment("订单.csv", 2048),
    kit.Bubble(kit.Label("链接将在 24 小时后过期")).Variant(kit.BubbleGhost),
)
msg := kit.Message("助手", content).Header(kit.Label("刚刚"))
```

`SetItems` 复制新顺序并忽略 nil，空参数清空；`Items` 返回副本。应复用子项，每个实例只出现一次；自定义 View 需提供稳定 ID。存续气泡使用内部稳定副本渲染，动态重排保留其输入与焦点，原气泡的样式和对齐不会被修改。更改源气泡的变体会在下一帧更新样式和 Ghost 继承。

`Gap` 设置非负有限 dp，默认 `theme.SpaceMd`；`Style` 配置正文栈，nil 恢复默认。样式回调在默认值后执行，不应保留元素或添加子项。`SetDisabled` 禁用整个正文，祖先 Message 的禁用同样有效。

直接通过 Message 构造参数或 `Content(content)` 安装时，组件负责混排，不会再包一层 User 气泡。独立 Render 默认靠左，并使用接收方默认气泡颜色。普通 View 包裹后的气泡不会被递归检测；自动 Ghost 继承只针对直接 Bubble 子项。

运行：`go run ./examples/components -section message_content`，加 `-theme dark` 检查深色。
