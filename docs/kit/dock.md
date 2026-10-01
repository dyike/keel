# Dock

像 IDE 一样，把工具面板停靠在中间内容的左、右、下三侧。

```go
d := kit.Dock(editor).
    Panel(kit.DockPanel{ID: "files", Title: "文件", View: fileTree}, kit.DockLeft).
    Panel(kit.DockPanel{ID: "term", Title: "终端", View: terminal}, kit.DockBottom).
    OnLayoutChange(func(l kit.DockLayout) { save(l) })
d.SetLayout(loaded) // 恢复上次保存的布局
```

- 每个停靠区可以放多个面板，用标签切换；停靠区和中间内容之间可以拖动调整大小。
- 面板标题栏右侧的菜单可以把面板移到另一侧，或者关闭它；`SetVisible(id, true)` 会在原来的停靠区重新打开。
- 面板会拿到停靠区的全部高度。Tree、Table、List 这类组件可以用 `Fill()` 撑满，并自带滚动；较长的普通内容要自己包一层 `ScrollY`。
- `DockLayout` 记录每个停靠区里有哪些面板及其顺序、当前标签、各区尺寸、已关闭的面板。它可以直接编码成 JSON 保存。`SetLayout` 会忽略不认识的面板 ID，布局里没提到的面板保持原位。
- 用户移动、关闭、切换面板或拖动调整大小之后，调用 `OnLayoutChange`。
- 窗口太窄时，左右两区会按比例缩小，给中间内容至少留出 120dp。

Agent：每个停靠区是以当前面板标题命名的 `region`，标签栏是 `tablist`，菜单按钮名为"更多 面板标题"，菜单项是"停靠到左侧""停靠到右侧""停靠到底部""关闭"。

验证：`go run ./examples/components -section dock`，加 `-theme dark` 检查深色。
