# Alert

`kit.Alert("保存失败").Danger().Description("网络不可用").OnClose(fn)` 显示行内提示。默认 Info，支持 Success、Warning、Danger；SetKind、SetTitle、SetDescription 可程序更新，Tone 保留为兼容入口。

左侧等级条和图标随主题取色，标题加粗，描述为 13sp Muted。只有设置 OnClose 才显示关闭按钮；点击或 Tab 聚焦后 Space/Enter 关闭，先隐藏再回调。Visible 查询状态，SetVisible 恢复或隐藏时不调用回调。

Agent 容器角色 alert，名字为标题，value 为等级。描述和关闭按钮单独可读。窄容器文字换行。运行 `go run ./examples/components -section alert -theme dark`；省略 theme 查看浅色。
