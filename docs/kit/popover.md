# Popover

在触发元素旁边显示一块非模态面板，用于筛选条件、简短表单这类内容。

```go
filters := kit.Popover(form).Width(280)
filters.Trigger(kit.Button("筛选", filters.Toggle).Variant(kit.ButtonSecondary))
```

- 触发元素自己负责打开：把 `Toggle` 传给它的点击回调。Popover 只在触发元素外面包一层不可交互的锚点，不会多出 Tab 停靠点，键盘行为完全由触发元素决定。
- 以下操作会关闭面板：再次点击触发元素、在面板和触发元素之外按下鼠标、按 Esc。面板外的那次点击会继续传给下面的元素。
- 面板不会移动焦点。如果内容里有输入框，用户需要自己点进去或用 Tab 进入。
- `Value()` 返回是否打开；`SetValue(bool)` 用程序打开或关闭，不触发回调；`OnChange(fn)` 只在用户操作时调用。
- `Placement(side, align)` 设置面板相对触发元素的位置，默认 `el.Bottom, el.Start`；放不下时自动翻到对侧。
- 需要 `el.Root`（见 [el · 浮层](../el.md#浮层e4--e5)）。

Agent：面板的角色是 `dialog`，里面的元素单独列出；页面其余部分仍然可见。

验证：`go run ./examples/components -section popover`，加 `-theme dark` 检查深色。
