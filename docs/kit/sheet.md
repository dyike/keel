# Sheet

贴着窗口某条边的模态面板，用于详情、设置、较长的表单。

```go
details := kit.Sheet(el.Right, "订单详情").Body(view).Size(400)
details.SetValue(true)
```

- 方向：`el.Right`、`el.Left`、`el.Top`、`el.Bottom`。
- 尺寸：`Size(dp)` 对左右方向是宽度，对上下方向是高度，默认 360；不会超过窗口。
- 打开时用 `kit.SheetSlide`（200ms）从边缘滑入。开启减少动画时直接出现在最终位置，自动化模式下默认就是这样。
- 关闭方式和 Dialog 相同：Esc、点击遮罩、标题栏的关闭按钮，关闭时调用 `OnClose(fn)`。
- 内容区可以滚动。焦点限制在面板内，关闭后焦点回到原来的位置。
- `Value()` / `SetValue(bool)` 读取或设置是否打开，`SetTitle` 修改标题。需要 `el.Root`。

Agent：角色是 `dialog`，名字是标题；关闭按钮的名字是"关闭"。

验证：`go run ./examples/components -section sheet`。

Sheet 的正文可以包含 Menu、Popover 等浮层；父层先登记，Esc 从最内层关闭，最后才关闭 Sheet。`SetDisabled(true)` 关闭并阻止重新打开，祖先禁用或隐藏也会关闭模态层，恢复启用不会自动重开。用户关闭回调至多执行一次。

滑入距离使用受窗口约束后的实际宽高。尺寸为 NaN 或无穷时忽略；关闭后立即重新打开会重新开始动画，不要求中间先渲染一次关闭状态。示例“订单操作”可验证嵌套菜单。
