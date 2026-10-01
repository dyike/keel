# GroupBox

给一组视图提供标题、边框和内边距，不拦截子组件交互。

```go
group := kit.GroupBox("账户", kit.Tag("已验证"))
return group.Render(cx)
```

`GroupBox(title, children ...el.View)` 返回 `*GroupBoxView`。`SetTitle` 更新标题；`SetChildren` 复制视图切片并保留视图实例，每帧调用子视图 Render，nil 子项跳过。子视图的可变状态由调用方持有。无展开/收起、焦点、禁用等容器行为。

Agent 角色 group，名称为标题，子组件仍单独可见。布局使用主题 Surface 和 Border，跟随浅深色。验证：`go run ./examples/components -section group-box -theme dark`，支持 light。
