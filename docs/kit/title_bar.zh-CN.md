# TitleBar

[English](title_bar.md) | 简体中文

无边框窗口的应用标题栏。macOS 可使用系统原生交通灯，应用负责标题和工具按钮。

```go
bar := kit.TitleBar("编辑器").
    Leading(toggleSidebar).
    Trailing(searchBox, shareButton)

window.Open(window.Options{Title: "编辑器", Frameless: true, NativeTrafficLights: true,
    TrafficLightLayout: &window.TrafficLightLayout{Height: kit.TitleBarHeight, Left: 12, Spacing: 20},
    Content: el.Root(app)})
// app 的 Render 里把 bar.Render(cx) 放在最上面
```

- **拖动**：按住标题栏的空白处可以拖动窗口，标题文字所在的区域也算空白。拖动交给系统完成，所以窗口吸附、跨屏移动等行为和原生窗口一致。
- **窗口按钮**：
  - macOS：上例启用 AppKit 原生关闭、最小化和全屏按钮，TitleBar 仅预留空间；系统管理按钮外观、悬停和操作。未启用 `NativeTrafficLights` 时使用旧的绘制按钮，第三个按钮调用缩放。标题居中。
  - Windows、Linux：右侧依次是最小化、最大化 / 还原、关闭，关闭按钮悬停时显示红色。标题靠左。
  - 绘制按钮通过 `core.CurrentWindow()` 调用所在窗口的 `Minimize`、`ToggleMaximize`、`Close`，最大化状态变化时自动切换成"还原"。
- **应用自己的内容**：`Leading` 的内容放在窗口按钮后面，`Trailing` 的内容放在右侧。这些内容都在拖动区域之外，可以正常点击和输入。
- 窗口不是 `Frameless` 时，系统标题栏还在，TitleBar 只绘制标题和应用内容，相当于一个页头，不显示窗口按钮，也不登记拖动区域。
- 高度为 `kit.TitleBarHeight`（38dp）。

Agent：标题栏的角色是 `banner`，名字是标题；窗口按钮名为"关闭""最小化""最大化"或"还原"，应用内容单独列出。

验证：运行 `go run ./examples/components -section title_bar -lang zh-CN`，点击 **打开无边框窗口**，体验窗口拖动、窗口按钮、标题栏搜索和命令面板。组件库中的入口为“布局与外壳 → TitleBar”。原独立 frameless 示例已合并到这里。

窗口失焦时标题使用次要文字颜色，macOS 窗口按钮变灰；重新激活后恢复。前后插槽切片会被复制，空插槽忽略。

macOS 双击只作用于当前绘制的中间拖动区域，前后插槽和窗口按钮不包含在内。原生本地事件监视器在 Gio 启动第二次拖动前处理双击：系统设置为最小化时最小化，设置为无操作时不操作，其余情况切换缩放。标题栏隐藏、所属区域禁用或窗口销毁时移除登记。自动化窗口通过普通双击回调测试缩放；真实窗口的拖动仍交给 AppKit。

原生交通灯出现在系统辅助功能树中，不会出现在 Keel 的离屏快照里。标题栏内、拖动区域外的 Gio 控件双击会直接交给 Gio，避免透明标题栏同时触发系统缩放；系统按钮仍走 AppKit 事件分发。2026-11-10 在 Apple silicon 真窗口中验证了搜索按钮双击不再缩放、侧栏拖动及折叠恢复；原生窗口测试检查 NSWindow 的标准按钮身份、操作和折叠/调整窗口后的固定位置。

连续点击也遵守相同边界：第 3 下及后续点击不会交回系统触发缩放。空白拖动区仅在第 2 下执行双击动作，控件区继续接收正常点击。AppShell 的两种折叠方式均在真实 macOS 窗口中复测了 3 次和 6 次连点，窗口尺寸保持不变。

已知验证缺口：2026-11-10 的完整真实窗口回归仍报告了程序调尺寸后的交通灯隐藏或位置重排，以及窗口水平居中偏移 2.5dp。连续点击检查通过不代表这两项已修复；本次不宣称完整原生窗口回归通过。
