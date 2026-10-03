# Alert

`kit.Alert("保存失败").Tone(kit.ToneDanger).Description("网络不可用").OnClose(fn)` 显示行内提示。默认 ToneInfo；Tone 支持 ToneNeutral、ToneInfo、ToneSuccess、ToneWarning、ToneDanger。SetTone、SetTitle、SetDescription 可程序更新。

左侧等级条和图标随主题取色，标题加粗，默认描述为 13sp Muted。

- `Size(AlertSizeXSmall/AlertSizeSmall/AlertSizeMedium/AlertSizeLarge)` 设置四档字号、图标与内边距，默认 Medium 保留原布局；非法值忽略。
- `Icon(IconName)` 替换等级图标，`IconNone` 隐藏图标及其间距。
- `Content(el.View)` 替换描述，可传 Markdown 文档、富文本或操作按钮；传 nil 恢复 Description。正文控件保留自己的键盘与点击行为，并继承 Alert 的禁用状态。
- `Banner(true)` 使用满宽、直角、无边框的染色横幅，不显示单独标题行；显示 Content 或 Description，没有正文时以标题作消息。关闭按钮和 Agent 名称仍使用标题。传 false 恢复行内提示。

```go
notice := kit.Alert("维护通知").Banner(true).
    Description("今晚进行例行维护").Icon(kit.IconCalendar).
    Size(kit.AlertSizeSmall)
// 富文本由应用组合：notice.Content(markdown.New("**注意**：请先保存工作"))
```

只有设置 OnClose 才显示关闭按钮；点击或 Tab 聚焦后 Space/Enter 关闭，先隐藏再回调。Visible 查询状态，SetVisible 恢复或隐藏时不调用回调。

Agent 容器角色 alert，名字为标题，value 为等级。描述和关闭按钮单独可读。窄容器文字换行。运行 `go run ./examples/components -section alert -theme dark`；省略 theme 查看浅色。

SetDisabled(true) 禁止关闭并向按钮传递 disabled 语义；恢复启用后可以重新聚焦和关闭，不会因测量或主题切换丢失显隐状态。
