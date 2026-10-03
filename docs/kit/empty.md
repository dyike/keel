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
