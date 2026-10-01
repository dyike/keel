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
