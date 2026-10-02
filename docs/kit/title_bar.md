# TitleBar

无边框窗口的标题栏，由应用自己绘制。

```go
bar := kit.TitleBar("编辑器").
    Leading(toggleSidebar).
    Trailing(searchBox, shareButton)

window.Open(window.Options{Title: "编辑器", Frameless: true, Content: el.Root(app)})
// app 的 Render 里把 bar.Render(cx) 放在最上面
```

- **拖动**：按住标题栏的空白处可以拖动窗口，标题文字所在的区域也算空白。拖动交给系统完成，所以窗口吸附、跨屏移动等行为和原生窗口一致。
- **窗口按钮**：
  - macOS：左侧三个圆形按钮，依次是关闭、最小化、缩放，颜色沿用系统惯例，指针移上去时显示符号。标题居中。
  - Windows、Linux：右侧依次是最小化、最大化 / 还原、关闭，关闭按钮悬停时显示红色。标题靠左。
  - 按钮通过 `core.CurrentWindow()` 调用所在窗口的 `Minimize`、`ToggleMaximize`、`Close`，最大化状态变化时自动切换成"还原"。
- **应用自己的内容**：`Leading` 的内容放在窗口按钮后面，`Trailing` 的内容放在右侧。这些内容都在拖动区域之外，可以正常点击和输入。
- 窗口不是 `Frameless` 时，系统标题栏还在，TitleBar 只绘制标题和应用内容，相当于一个页头，不显示窗口按钮，也不登记拖动区域。
- 高度为 `kit.TitleBarHeight`（38dp）。

Agent：标题栏的角色是 `banner`，名字是标题；窗口按钮名为"关闭""最小化""最大化"或"还原"，应用内容单独列出。

验证：`go run ./examples/frameless` 打开真正的无边框窗口，拖动标题栏、点击窗口按钮；`go run ./examples/components -section title_bar` 查看普通窗口里的页头样式。

窗口失焦时标题使用次要文字颜色，macOS 窗口按钮变灰；重新激活后恢复。前后插槽切片会被复制，空插槽忽略。

macOS 双击只作用于当前绘制的中间拖动区域，前后插槽和窗口按钮不包含在内。原生本地事件监视器在 Gio 启动第二次拖动前处理双击：系统设置为最小化时最小化，设置为无操作时不操作，其余情况切换缩放。标题栏隐藏、所属区域禁用或窗口销毁时移除登记。自动化窗口通过普通双击回调测试缩放；真实窗口的拖动仍交给 AppKit。

验证记录（2026-10-02）：离屏双击、拖动区域排除控件、禁用清理和失焦像素测试通过。当前环境运行原生示例时，Gio 在创建 DisplayLink 阶段崩溃；改动前版本同样失败。因此 AppKit 双击偏好与系统拖动仍需在可运行桌面窗口的环境复核。
