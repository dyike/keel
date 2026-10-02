# StatusBar

`kit.StatusBar().Left(status).Right(position)` 创建固定 24dp 状态栏，顶部 1dp 边框，默认 12sp Muted 文本。左右内容分组，中间空间由 Grow 撑开，长文本限制单行。父视图决定状态栏位置。

Left/Right 接受 el.View，替换各自的视图组，每帧调用子视图的 Render。简单内容可用 el.ViewFunc 包装：

```go
status := el.ViewFunc(func(cx *el.Context) el.Element {
    return el.Text("就绪").TextColor(theme.Muted)
})
```

## 溢出菜单

窄窗口里放不下所有内容时，用 `Add`（左组）和 `AddRight`（右组）加入可以收起的项：

```go
bar := kit.StatusBar().Left(ready).
    Add(kit.StatusItem{Label: "main 分支", Action: switchBranch, Priority: 3}).
    AddRight(
        kit.StatusItem{Label: "行 12，列 4", Action: gotoLine, Priority: 2},
        kit.StatusItem{Label: "UTF-8", Action: pickEncoding},
    )
```

- 放不下时，`Priority` 低的先收进右侧的 “…” 菜单；同优先级时排在后面的先收。收起一个宽项后空出的位置，会让之前收起的窄项回来。
- 窗口变宽后，收起的项自动回到状态栏。
- `Left`、`Right` 传入的视图始终显示，不会收起，文字过长时截断。
- `View` 为空时状态栏显示 `Label`；有 `Action` 时它是可点的按钮。菜单里总是显示 `Label`，点击运行 `Action`。
- 状态栏要按宽度记住每一项的尺寸，所以要保存在视图里复用，不要每帧新建。
- `Hidden()` 返回当前收起的项的 `Label`。

Agent 角色 status，子元素单独列出；“…” 是名为“更多”的按钮。

验证：`go run ./examples/components -section status_bar -theme dark`，省略 theme 查看浅色。
