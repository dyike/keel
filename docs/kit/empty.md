# Empty

`kit.Empty("暂无订单").Description("新建订单开始").Icon(kit.IconInbox).Action(action)` 垂直居中显示图标、标题、说明与操作区。Action 接受 el.View，每帧调用其 Render。暂可用 el.ViewFunc 包装 el.Div().OnClick 组合按钮。

SetTitle、SetDescription 更新文案。Empty 不新增语义角色，标题和说明分别作为 text，操作元素保留自身语义与键盘行为。窄容器文字换行。验证：`go run ./examples/components -section empty`，加 `-theme dark` 检查深色。

```go
kit.Empty("暂无订单").Action(el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Div().OnClick(createOrder).Child(el.Text("新建订单").TextColor(theme.Primary))
}))
```

`Media(view)` 在标题上方放置头像、图片、头像组或任意自定义内容，保留内容自身的尺寸、语义和交互。后一次调用替换前一次内容；传 nil 恢复 `Icon` 配置，`Icon(kit.IconNone)` 隐藏默认图标。媒体、标题和说明变化不会重建 Action 的输入状态或焦点。媒体仍需适配父容器宽度；滚动由父容器提供。

```go
kit.Empty("暂无成员").Media(kit.Avatar("Alex").Size(48)).Action(kit.Button("邀请", invite))
```

`Heading(view)` 和 `DescriptionContent(view)` 替换标题、描述的字符串展示，传 nil 恢复字符串；此时 SetTitle / SetDescription 更新的是回退文案。`Footer(view)` 在 Action 下方添加独立辅助内容，传 nil 移除。富内容的语义由内容自身提供，不重复宣布被替换的字符串。

`PartStyle(part, func(*el.DivEl))` 每帧在默认样式之后调整分区，支持 `EmptyPartRoot`、`EmptyPartHeader`、`EmptyPartMedia`、`EmptyPartTitle`、`EmptyPartDescription`、`EmptyPartContent`（Action）、`EmptyPartFooter`。可设置背景、边框、圆角、宽度、间距、对齐和继承字号/颜色；子内容显式样式优先。传 nil 恢复该分区默认样式，非法 part 忽略。回调不要保留元素引用；子分区 ID 由组件固定，以保留操作区状态。

```go
kit.Empty("没有结果").
    Heading(kit.Tag("没有结果")).
    Footer(kit.Button("查看帮助", help).Variant(kit.ButtonLink)).
    PartStyle(kit.EmptyPartRoot, func(e *el.DivEl) {
        e.Bg(theme.Subtle).Rounded(theme.RadiusLg)
    })
```

默认布局仍保留 Keel 的 Surface 背景和间距；Empty 不新增自动播报或焦点目标，内容的按钮与输入框保持自身行为。
