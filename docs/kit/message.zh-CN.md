# Message

[English](message.md) | 简体中文

对话中的一条消息：头像、内容、操作栏。

```go
kit.Message("我", text).User()                  // 用户消息：靠右的气泡，不显示头像
kit.Message("AI 助手", answerDoc).Actions(copy) // 其他人：头像 + 整宽内容 + 操作栏
```

- 非用户消息占满宽度，适合长篇 Markdown 回答。头像显示作者名字的首字母。
- `Actions(views...)` 显示在内容下方，比如复制、重试按钮。流式输出还没结束时，通常先不加操作。

Agent：每条消息是 `article`，名字是作者，内容和操作按钮单独列出。

验证：`go run ./examples/components -section message`，加 `-theme dark` 检查深色。

`SetState(MessageSending, "")` 显示发送中；`SetState(MessageFailed, reason)` 显示失败原因。`OnRetry(fn)` 在失败时显示重试按钮，点击先转为发送中，再调用一次业务回调；业务完成后设置 `MessageReady`，再次失败则设置 `MessageFailed`。

用户和助手消息都支持 `Actions`，传入切片会被复制，空槽位会被忽略。操作栏是否在流式输出期间显示由应用决定。`SetDisabled(true)` 禁止消息内的操作、反应和重试。

```go
msg.Reactions(kit.MessageReaction{Name: "有帮助", Count: 2}).
    OnReaction(func(index int, active bool) { /* 保存到服务端 */ })
```

反应使用可切换按钮，点击更新当前用户的选中状态与计数，再调用回调。`Reactions` 复制数据；后续服务端结果可再次调用它覆盖。未提供回调的反应只读。Agent 可读取 `article` 的 `sending` / `failed` 状态，以及反应按钮的选中状态。

`Avatar(view)` 替换头像，传 nil 隐藏；`DefaultAvatar()` 恢复默认：助手显示姓名首字母，用户消息不显示头像。用户消息显式配置头像时放在右侧。头像按正文容器底边对齐，头部在上方，状态、操作、反应及尾部继续排在正文列下方；定位使用当前帧布局，较高的头像会撑开上方空间。正文被隐藏或不存在时，以正文列底边作为回退。

`Header(view)` 放在正文上方，`Footer(view)` 放在状态、操作和反应区之后；传 nil 清除。头尾支持任意 View 和可交互控件，默认继承小号、弱化文字样式，用户消息靠右。`Content(view)` 独立替换正文，nil 清空正文但保留其他分区。消息禁用状态覆盖所有插槽。

```go
msg.Avatar(kit.Avatar("Alice").Size(32)).
    Header(kit.Label("Alice · 10:24")).
    Footer(kit.Button("回复", reply))
```

插入、删除头像或头尾时，正文的身份保持稳定，输入内容和焦点不变。普通用户气泡和显式 Bubble 的头尾默认使用 `theme.SpaceLg` 水平缩进；普通助手正文保持无缩进。完整消息行可用 [MessageGroup](message_group.zh-CN.md) 分组。

`Bubble(surface)` 安装显式气泡，避免 User 消息再包一层气泡。组件以副本渲染，气泡对齐跟随消息；修改原气泡的变体会在下一帧体现，不会反向修改原实例。传 nil 清空正文；`Content(view)` 恢复普通正文模式（User 自动包气泡）。

```go
surface := kit.Bubble(answer).Variant(kit.BubbleGhost)
msg.Bubble(surface).Header(kit.Label("系统消息")).Footer(kit.Label("刚刚"))
msg.HeaderInset(true).FooterInset(false)
msg.ResetContentInsets() // 恢复自动规则
```

显式 Ghost 气泡自动取消头尾缩进；`HeaderInset`、`FooterInset` 分别覆盖继承规则。普通 Content 内部的自定义 View 不参与 Ghost 检测。该入口接受单个 Bubble；多个气泡与附件混排使用 [MessageContent](message_content.zh-CN.md)，直接气泡共同参与 Ghost 继承。

`Alignment(el.Start/el.End)` 独立控制左右位置，覆盖 User 的默认靠右布局；`ResetAlignment()` 恢复默认。头像、头尾、状态、操作和反应跟随位置，显式气泡也跟随。User 仍决定默认气泡色和是否显示默认头像，改对齐不会改变这些设置。其他 Align 值忽略。普通正文靠右时按自身宽度布局，撑满宽度的控件仍占满正文列。

`PartStyle(part, func(*el.DivEl))` 在每帧默认布局之后配置分区样式，nil 恢复默认，非法 part 忽略。可选分区：Root（外层）、Stack（正文列）、Avatar、Header、Content、Footer、Status、Actions、Reactions，对应常量均以 `MessagePart` 开头。头尾样式在自动缩进及显式覆盖后执行，可进一步调整内边距。

```go
msg.PartStyle(kit.MessagePartRoot, func(e *el.DivEl) {
    e.P(theme.SpaceLg).Bg(theme.Subtle).Rounded(theme.RadiusMd)
}).PartStyle(kit.MessagePartStack, func(e *el.DivEl) {
    e.Gap(theme.SpaceSm)
})
```

回调只修改当帧元素，不应保留它或追加子项。组件保留内部 ID、外层 article 的名称/状态及整体禁用；分区可额外禁用自身，不能绕过祖先禁用。Content 样式作用于正文容器，气泡表面仍由 Bubble 的样式接口控制。

自定义 `MessagePartRoot` 的 Items 可覆盖默认底边对齐；`MessagePartStack` 的 ContentBottom 可选择另一后代作为对齐目标。改变这些布局规则由应用负责。默认头像仍为 28dp，助手自动显示头像、User 自动包主色气泡的约定保留，与上游默认无槽位不同。
