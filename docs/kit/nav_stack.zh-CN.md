# NavStack

[English](nav_stack.md) | 简体中文

用于列表／详情流程的全高导航栈。只渲染当前页；下层页面在退出栈之前保留控件状态和滚动位置。放入 `el.Root` 或高度有界的容器。

```go
nav := kit.NavStack()
nav.Page(kit.NavPage{ID: "home", Title: "首页", Create: func() el.View {
    return kit.Button("详情", func() { nav.Push("detail") })
}})
nav.Page(kit.NavPage{ID: "detail", Title: "详情", Create: func() el.View {
    return kit.Input("备注")
}})
// window.Open(window.Options{Content: el.Root(nav)})
```

## API 与状态

- `Page(NavPage)` 注册 ID、标题和可选的 `Create func() el.View`。第一个启用页面成为初始根页面。空 ID、重复 ID 会忽略；禁用页面不能进入或恢复。
- `Push(id) bool`、`Pop() bool` 是导航命令。成功时通知 `OnChange(func([]string))`；拒绝时返回 false。不能 Pop 根页面，标题栏的上一页按钮调用 Pop。
- `Value() []string` 返回从根到当前页的路径副本，回调也收到副本。
- `SetValue([]string)` 静默恢复路径，保留与现有路径相同的前缀。任一 ID 未注册或已禁用时拒绝整个赋值。空路径清空栈，同一 ID 可以作为不同栈条目重复进入。
- `SetDisabled(bool)` 禁用导航及后代交互，禁用期间仍可静默恢复路径。
- `NavPage.Fill` 让页面自行负责滚动；否则栈提供带内边距的滚动容器。每个条目首次显示时调用一次 Create。Pop 释放该条目，之后再次 Push 会创建新状态；应用保存的外部引用仍由应用管理。

导航栈不是无限缓存。当前页下方的页面仍占用内存，应使用 Pop 或 SetValue 移除不再需要的条目。重复 Render 不会重新创建页面，也不会让隐藏页动画。Create 可以返回 nil，表示空目的页。

## 交互与语义

标题为 `group` 命名，其下暴露当前页面语义。隐藏页面不会出现在自动化快照中，也不接收输入。本地化的上一页按钮在默认主题度量下至少为 44 dp；Tab 可聚焦，Enter／Space 激活。父级禁用状态会穿过保留的页面根节点传递。

组件不拦截 Escape、浏览器历史记录或系统返回事件，由宿主决定弹层关闭与路由处理；宿主可调用 Pop 或使用 SetValue 恢复路径，安全区边距也由宿主提供。此初始组件尚未实现模态页面和滑动返回。

## 验证

`ui/kit/nav_stack_test.go` 覆盖输入、键盘激活、滚动保留、公共前缀恢复、路径副本、非法／禁用导航及父级禁用。`ui/window/kit_nav_stack_test.go` 在 320／600 dp 下检查实际语义快照和交互。组件库的 `nav_stack` 展示三级导航及草稿保留。
