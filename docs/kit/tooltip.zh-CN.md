# Tooltip

[English](tooltip.md) | 简体中文

给任意视图加一句简短提示。

```go
copy := kit.WithTooltip(kit.Button("", doCopy).Icon(kit.IconCopy), "复制（⌘C）")
```

- 什么时候显示：指针停留 `kit.TooltipDelay`（500ms）后显示；键盘焦点进入被包裹的视图时立即显示。
- 什么时候隐藏：指针移开、焦点离开、按 Esc。按 Esc 隐藏后，要等指针和焦点都离开一次，才会再次显示。
- 显示位置：默认在上方居中，放不下时翻到下方。
- 提示本身不能获得焦点，也不响应点击。
- `SetText` 用来更新提示文字，比如复制后改成"已复制"。

Agent：角色是 `tooltip`，名字是提示文字；富内容和动作键位作为子元素列出。

验证：`go run ./examples/components -section tooltip`。

悬停延迟绑定到实际可见且启用的目标，移开、禁用或模态遮挡会停止等待；恢复后重新等待完整延迟。`SetDisabled(true)` 同时禁用提示和目标区域。提示宽度会受当前窗口限制。多个 Tooltip 的延迟独立计算，不沿用上一个目标的等待时间。

`Content(view)` 替换显示内容，支持多行文字、图标等；原 text 保留为语义名称，必须非空。传 nil 恢复纯文字。富内容中的控件禁用，不会取得焦点或执行点击；需要交互的预览请用 HoverCard。

`Action(name)` 显示 `core.Bind` 中该动作的第一个快捷键，自动跟随改绑；未绑定时不显示，不负责注册或执行动作。`Placement(side, align)` 设置方向和对齐，`Offset(dp)` 设置间距（默认 4dp，忽略非有限值）；定位仍受窗口边缘避让约束。
