# NavStack

[English](nav_stack.md) | 简体中文

用于列表／详情流程的全高导航栈。每个展示层只渲染当前页；下层页面在退出栈之前保留控件状态和滚动位置。放入 `el.Root` 或高度有界的容器。

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
- `NavPage.Modal` 在根页面之上 Push 时开启模态展示，之后的普通 Push 留在该模态展示内。根页面的 Modal 标记不生效。SetValue 也可恢复模态路径。
- `Dismiss() bool` 关闭最近的模态展示及其内部压入的页面，只通知一次。Esc、遮罩和 44 dp 关闭按钮调用 Dismiss；返回按钮和滑动调用 Pop，只退一页。嵌套模态从内向外关闭，底层页面保留状态并禁用交互；关闭后恢复到仍挂载的原控件焦点。
- `SwipeBack(false)` 关闭默认的左边缘滑动返回。从内容区左侧 24 dp 内开始右拖，至少达到内容宽度的 30%（最低 64 dp），且水平位移明显大于竖向位移。短距离、反向或取消手势不导航；竖向滚动和交互子控件保留输入。支持触摸和主指针拖动，内容随拖动移动，取消后弹回；减少动画时直接复位。Fill 页面中的自有滚动控件可以优先接管手势。
- `Value() []string` 返回从根到当前页的路径副本，回调也收到副本。
- `SetValue([]string)` 静默恢复路径，保留与现有路径相同的前缀。任一 ID 未注册或已禁用时拒绝整个赋值。空路径清空栈，同一 ID 可以作为不同栈条目重复进入。
- `SetDisabled(bool)` 禁用导航及后代交互，禁用期间仍可静默恢复路径。
- `NavPage.Fill` 让页面自行负责滚动；否则栈提供带内边距的滚动容器。每个条目首次显示时调用一次 Create。Pop 释放该条目，之后再次 Push 会创建新状态；应用保存的外部引用仍由应用管理。

导航栈不是无限缓存。当前页下方的页面仍占用内存，应使用 Pop 或 SetValue 移除不再需要的条目。重复 Render 不会重新创建页面，也不会让隐藏页动画。Create 可以返回 nil，表示空目的页。

## 交互与语义

标题为 `group` 命名，其下暴露当前页面语义。隐藏页面不会出现在自动化快照中，也不接收输入。本地化的上一页按钮在默认主题度量下至少为 44 dp；Tab 可聚焦，Enter／Space 激活。父级禁用状态会穿过保留的页面根节点传递。

普通页面不拦截 Esc。浏览器历史、系统返回和安全区边距仍由宿主处理。模态面板占用可用高度，宽度最多 640 dp 且不超出视口；焦点循环包含保留页面根中的控件。

## 验证

`ui/kit/nav_stack_test.go` 覆盖输入、键盘激活、滚动保留、公共前缀恢复、路径副本、非法／禁用导航及父级禁用。`ui/window/kit_nav_stack_test.go` 在 320／600 dp 下检查实际语义快照和交互。组件库的 `nav_stack` 展示三级导航、草稿保留及模态导航。`nav_stack_mobile_test.go` 覆盖模态关闭、嵌套展示、焦点恢复、滑动完成／取消、子控件输入及竖向滚动。窗口测试覆盖 320／390／600 dp 和 1／1.6 字体比例。这些是本地事件及窗口测试，不代表真机或输入法验收。
