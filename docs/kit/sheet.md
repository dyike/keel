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

`Footer(views...)` 设置独立页脚，复制传入切片，支持多个操作按钮与换行；正文继续独立滚动，页脚不随正文滚动。无正文时页脚靠面板底部，空参数清除页脚，nil 子项忽略。页脚适合少量操作，需为标题和页脚留出足够面板高度。

`Keyboard(bool)`、`Overlay(bool)`、`OverlayClosable(bool)`、`CloseButton(bool)` 分别控制 Esc、遮罩颜色、外部点击关闭和标题关闭按钮，默认均开启。隐藏遮罩仍保持模态阻挡和焦点约束；关闭按钮不受 Keyboard/OverlayClosable 限制。配置可在打开期间更新，程序 SetValue(false) 始终可用。

`MarginTop(dp)` 给面板顶部预留空间，例如 `MarginTop(32)` 避开标题栏。四个方向均在剩余窗口区域内布局；顶部面板从预留区下沿滑入，底部面板仍贴底，左右面板缩短高度。绘制和点击区域同步裁剪，动画不会覆盖预留区。默认 0，负数和非有限值忽略；超出窗口高度时面板完全裁剪，仍可按 Esc 关闭。遮罩和模态阻挡继续覆盖整个窗口，点击顶部预留区按 OverlayClosable 处理。

`PanelStyle(func(*el.DivEl))` 在默认外观之后配置面板配色、文字颜色、边框、圆角、阴影、内边距和内容间距；`PanelStyle(nil)` 恢复默认。回调每帧接收新元素，不应保存引用。面板身份和 Size/窗口约束在回调后设置，正文与操作继续通过 Body/Footer 提供。显式设置颜色的子组件保持自身颜色，未设置的文字继承面板颜色。

```go
details.PanelStyle(func(panel *el.DivEl) {
    panel.Bg(theme.Surface).Border(1, theme.Border).P(24).Gap(theme.SpaceLg)
})
```
