# GroupBox

[English](group_box.md) | 简体中文

`kit.GroupBox("通知设置").Description("选择接收方式").Child(views...)` 在内容框外显示标题和说明，内容区带边框、圆角与 Surface 背景。

构造函数只接受标题。Child 接受 el.View 并追加到内容区，SetChildren 批量替换内容。每帧调用子视图的 Render，保留子视图交互。SetTitle 修改标题。Agent 角色 group，标题为名字，说明与内容单独列出。组件本身无键盘操作。

`Variant` 接受 GroupBoxSurface（默认，保留 Surface 背景和边框）、GroupBoxNormal（无背景/边框）、GroupBoxFill（Subtle 背景）和 GroupBoxOutline（透明底色加边框）。各外观均保留内容内边距。

`Footer(el.View)` 在内容框外添加底部说明或操作，和标题左侧对齐、间距 8dp；默认继承小号 Muted 文字，传 nil 移除。标题、描述和 footer 不受正文样式影响。

`TitleStyle(func(*el.TextEl))` 修改标题字号、颜色、字重、间距等；`ContentStyle(func(*el.DivEl))` 在外观默认值之后修改正文背景、边框、圆角、内边距与布局。回调每帧执行，可读取当前主题；只修改传入元素，不保留元素引用。传 nil 恢复默认。正文容器保持稳定身份，删除标题或更新 footer 不会重建内部输入状态。

```go
box := kit.GroupBox("通知设置").Variant(kit.GroupBoxFill).
    Child(kit.Switch("邮件通知", true)).
    Footer(el.ViewFunc(func(*el.Context) el.Element {
        return el.Text("设置仅用于当前设备")
    })).
    ContentStyle(func(e *el.DivEl) { e.P(24).Rounded(theme.RadiusLg) })
```

验证：`go run ./examples/components -section group_box -theme dark`，省略 theme 查看浅色。
