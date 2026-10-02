# Dock

像 IDE 一样，把工具面板停靠在中间内容的左、右、下三侧。

```go
d := kit.Dock(editor).
    Panel(kit.DockPanel{ID: "files", Title: "文件", View: fileTree}, kit.DockLeft).
    Panel(kit.DockPanel{ID: "term", Title: "终端", View: terminal}, kit.DockBottom).
    OnLayoutChange(func(l kit.DockLayout) { save(l) })
if !d.SetLayout(loaded) { /* 不支持的版本或非法布局，保留原布局 */ }
```

- 每个停靠区可以放多个面板，用标签切换；长标签栏可横向滚动；停靠区和中间内容之间可以拖动调整大小。
- 面板标题栏右侧的菜单可以把面板移到另一侧，或者关闭它；`SetVisible(id, true)` 会在原来的停靠区重新打开。
- 最大化：菜单里的"最大化"或双击标签，让这个面板铺满整个 Dock，中间内容和其他面板暂时隐藏；再点"还原"、双击标签或按 Esc 恢复。`Zoom(id)` / `Zoom("")` / `Zoomed()` 在程序里控制，最大化状态保存在 `DockLayout.Zoomed`。移走或关闭最大化的面板会自动还原。
- 面板会拿到停靠区的全部高度。Tree、Table、List 这类组件可以用 `Fill()` 撑满，并自带滚动；较长的普通内容要自己包一层 `ScrollY`。
- `DockLayout` 记录每个停靠区里有哪些面板及其顺序、当前标签、各区尺寸、已关闭的面板。它可以直接编码成 JSON 保存。版本为 2，旧的无版本布局和版本 1 迁移为单标签组；不认识的版本、重复跨区 ID、NaN/无穷尺寸会被原子拒绝，`SetLayout` 返回 false 且保留原布局。`SetLayout` 会忽略不认识的面板 ID，布局里没提到的面板保持原位。
- `SetDisabled` 禁用面板与布局操作。分隔条可用方向键按 10dp 调整，Home / End 到边界；取消拖动恢复原尺寸，普通点击分隔条不触发布局回调。
- 用户移动、关闭、切换面板或拖动调整大小之后，调用 `OnLayoutChange`。
- 窗口太窄时，左右两区会按比例缩小，给中间内容至少留出 120dp。

Agent：每个停靠区是以当前面板标题命名的 `region`，标签栏是 `tablist`，菜单按钮名为"更多 面板标题"，菜单项是"停靠到左侧""停靠到右侧""停靠到底部""关闭"。

验证：`go run ./examples/components -section dock`，加 `-theme dark` 检查深色。

重复注册面板 ID 会更新标题和 View，保留所在区域；空 ID、非法区域不会加入。布局快照及恢复输入都做副本隔离，未知面板的 `Visible` 为 false。

嵌套分割：`d.Split("search", "files", kit.DockPlacementBottom)` 把搜索放到文件组下方；还支持 Left / Right / Top。标题菜单提供“向右拆分”和“向下拆分”，把当前标签从同组其他标签中拆出。分隔条可拖动，或用方向键每次调整 5%，Home / End 到 5% / 95%。取消拖动或禁用会恢复拖动前比例。`Move` 移到指定区域的首个标签组，并选中移入的标签；空组自动合并。关闭的组保留在布局中，重新打开时回到原位。

`LeftTree` / `RightTree` / `BottomTree` 由 `DockNode` 表达嵌套：叶子用 Panels / Active，分割用 First / Second / Axis / Ratio。树存在时，其面板顺序优先于旧的平面列表；旧列表仍随布局更新。恢复会深拷贝树，拒绝循环、超过 32 层、重复面板、缺失分支或越界比例，未知面板会剔除并合并空分支。布局快照可以直接 JSON 往返。程序调用 `Split` / `Move` / `SetLayout` 不触发布局回调。

拖动标签时，标题栏显示插入位置；放到另一组的中间会合并，放到正文四边会拆分。编辑器的左、右、下边缘可恢复空停靠区；对应区域已有面板时合并到其中一组。半透明色块表示落点，松开才改布局并回调一次，焦点跟随移入的标签。Esc、指针取消、放到外部或禁用都撤销预览。布局内的坐标使用 dp，拖放命中按实际可见裁剪范围判断。

## 中心区文档

`Panel(p, kit.DockCenter)` 把面板放进中心区，当作文档。中心区和侧边区一样由标签组组成：

- 文档可以拆分（`Split`、标题菜单、拖到组的边缘）、拖动合并、最大化，布局通过 `CenterTree`、`Center`、`CenterActive` 一起保存。
- 有文档时，文档取代 `Dock(center)` 传入的中心视图；文档全部关闭或移走后，中心视图重新出现，适合放“没有打开的文档”这类空状态。
- 侧边面板的菜单多了“移到中间”，中心文档的菜单可以停靠到左、右、下。用过 `DockCenter` 的 Dock，中心区为空时，把标签拖到中间就会打开为文档。
- 中心区有文档时，只有紧贴边缘的 24dp 窄条用来恢复已经空掉的侧边区，其余位置的边缘用来拆分文档组。

## 跨窗口

```go
d.OnDetach(func(p kit.DockPanel, reattach func()) {
    window.Open(window.Options{Title: p.Title, Content: el.Root(p.View), OnClose: reattach})
})
```

- 设置 `OnDetach` 后，面板菜单多出“在新窗口打开”。把标签拖出 Dock 范围松开，也会分离出去。
- 分离的面板从 Dock 里移走，`Visible` 为 false，`Detached()` 列出它们，并调用一次 `OnLayoutChange`。
- 新窗口关闭时调用 `reattach`，面板回到原来的停靠区和标签组。要在界面回调里调用（`OnClose` 本来就是），其他 goroutine 用 `core.Update`。
- kit 不自己开窗口，窗口大小、标题栏等由应用决定。
- 恢复布局不会重新打开窗口：`SetLayout` 把分离的面板放回 Dock。
- 不设置 `OnDetach` 时，拖出 Dock 和以前一样是取消。

容器应有明确的宽高，通常直接用 `el.Root(d)`。嵌入普通页面时给外层指定宽高；Dock 内部按可用空间分配区域。

根视口中的首帧和每次缩放都先按当前视口分配停靠区，给中心保留 120dp；视口小于该值时，停靠区可缩到零，中心使用剩余空间。嵌入更小的容器时，首次绘制测得实际尺寸后会请求重绘并校正。此处“保留中心”不代表所有面板在手机宽度下都适合阅读。
