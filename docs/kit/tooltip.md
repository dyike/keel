# Tooltip

给任意视图加一句简短提示。

```go
copy := kit.WithTooltip(kit.Button("", doCopy).Icon(kit.IconCopy), "复制（⌘C）")
```

- 什么时候显示：指针停留 `kit.TooltipDelay`（500ms）后显示；键盘焦点进入被包裹的视图时立即显示。
- 什么时候隐藏：指针移开、焦点离开、按 Esc。按 Esc 隐藏后，要等指针和焦点都离开一次，才会再次显示。
- 显示位置：默认在上方居中，放不下时翻到下方。
- 提示本身不能获得焦点，也不响应点击。
- `SetText` 用来更新提示文字，比如复制后改成"已复制"。

Agent：角色是 `tooltip`，名字是提示文字。

验证：`go run ./examples/components -section tooltip`。

悬停延迟绑定到实际可见且启用的目标，移开、禁用或模态遮挡会停止等待；恢复后重新等待完整延迟。`SetDisabled(true)` 同时禁用提示和目标区域。提示宽度会受当前窗口限制。多个 Tooltip 的延迟独立计算，不沿用上一个目标的等待时间。
