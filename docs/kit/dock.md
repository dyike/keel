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


## 面板状态与工厂

只保存位置继续用 `Layout/SetLayout`。要在新 Dock 中重建面板，使用 `Snapshot/Restore`，并按类型注册工厂：

```go
factory := func(s kit.DockPanelState) (kit.DockPanel, error) {
    input := kit.Input("搜索")
    var query string
    if len(s.State) > 0 {
        if err := json.Unmarshal(s.State, &query); err != nil {
            return kit.DockPanel{}, err
        }
    }
    input.SetValue(query)
    return kit.DockPanel{View: input, SaveState: func() (json.RawMessage, error) {
        return json.Marshal(input.Value())
    }}, nil // Restore 填入原来的 ID、Kind 和 Title
}
if err := d.RegisterPanel("search", factory); err != nil { /* 处理错误 */ }
state, err := d.Snapshot()
// json.Marshal(state) 保存；读回后 json.Unmarshal 到 kit.DockState。
if err == nil { err = anotherDock.Restore(state) } // anotherDock 也须注册 search
```

首次添加面板时设置 `DockPanel.Kind: "search"` 和 `SaveState`；工厂接收实例 ID、类型、标题和 JSON 数据，可以恢复同类型的多个实例。返回值的非空 ID/Kind 必须与快照匹配，View 不得为 nil；标题为空时继承快照，否则采用工厂标题。应用负责数据版本迁移，工厂也应返回新的 `SaveState`，才能继续保存编辑后的值。

`DockState` 版本为 1，内部 `DockLayout` 仍是版本 2。快照按 ID 排序，复制布局和 JSON，包括隐藏、分离的面板。注册表属于单个 Dock，空类型、nil 工厂和重复类型返回错误。没有 Kind 的既有静态面板可按 ID 原样复用，但不能携带 SaveState 或数据；要跨新实例恢复，所有面板都应有已注册的 Kind。

恢复先检查整份清单、JSON 和布局，再调用工厂。未知类型、重复 ID、布局引用缺失的面板、清单中没有布局位置的面板、非法树或工厂错误都会返回错误，保留当前 Dock。恢复成功后，面板集合以快照为准，当前多出来的面板被移除；中心空视图、外观、禁用状态、注册表和回调保留。工厂应只构造视图，外部副作用和已创建资源由应用管理，Dock 无法替应用回滚。

这些操作在 UI 线程调用，不触发 `OnLayoutChange`。面板数据编辑不会触发布局回调，应用应在保存工作区或关闭窗口时显式调用 `Snapshot`。分离面板恢复到 Dock 内，不自动开关应用窗口；旧窗口关闭回调不会影响恢复后的新实例。

## 独立外观

```go
d.Skin(&kit.DockSkin{
    Header: func(e *el.DivEl) { e.Bg(theme.Bg) },
    Body: func(e *el.DivEl) { e.P(theme.SpaceLg) },
    Tab: func(e *el.DivEl, selected bool) {
        if selected { e.Bg(theme.Primary).TextColor(theme.PrimaryText) }
    },
    Separator: func(e *el.DivEl) { e.Bg(theme.Primary) },
})
```

`Panel` 配置标签组外框，`Header/Body/Tab/Separator` 分别配置标题栏、正文、标签与内外分隔条。回调每帧应用在新元素上，可改颜色、边框、字号和面板留白；保留元素身份、子内容和事件处理，分隔条保持 4dp 几何。配置对象可以共享并在 UI 线程更新，`Skin(nil)` 恢复默认外观，不改布局或面板内容。皮肤不进入布局 JSON，也不修改全局主题。它只管样式；面板自己的标签、工具栏和菜单见下一节。

组件库示例的“保存工作区 / 恢复工作区”可以验证搜索词随布局恢复，“切换 Dock 外观”用于检查皮肤。原生窗口、浅深色视觉和多窗口生命周期仍需真机验收。

## 面板的标签、工具栏和菜单

`DockPanel` 的可选字段让每个面板决定自己在 Dock 里的样子和行为：

```go
d.Panel(kit.DockPanel{
    ID: "files", Title: "文件", View: files,
    Icon:    kit.IconFolder,                                     // 标签上的图标
    Toolbar: kit.Button("", refresh).Name("刷新").Icon(kit.IconRetry).Variant(kit.ButtonGhost).Size(24),
    Menu:    func(m *kit.MenuView) { m.Item("全部折叠", "", collapseAll) },
    NoClose: true, // 菜单里没有“关闭”
}, kit.DockLeft)
```

- `Icon` 显示在标题前；`Tab(selected)` 完全替换标签内容（例如带状态点），`Title` 仍是标签的无障碍名字。
- `Toolbar` 在面板是当前标签时显示在标题栏右侧、菜单按钮之前。
- `Menu` 往面板菜单里加项，排在 Dock 自带的移动、拆分、最大化、关闭之前。
- `NoClose` 去掉“关闭”；`NoZoom` 去掉“最大化”，双击标签和 `Zoom(id)` 也不再最大化它。
- `NoPadding` 去掉正文留白，适合终端、画布这类贴边绘制的面板。

## 收起侧栏

```go
toolbar.Child(d.RegionButton(kit.DockLeft).Render(cx), d.RegionButton(kit.DockBottom).Render(cx))
```

`RegionButton(side)` 返回一个切换按钮，侧栏展开时为选中状态，可放进标题栏或工具栏；按钮建一次后复用。`SetRegionOpen(side, open)` / `RegionOpen(side)` / `ToggleRegion(side)` 是对应的程序接口，`ToggleRegion` 会触发 OnLayoutChange。收起只是隐藏整块区域，面板的标签、拆分和尺寸都保留，再展开原样回来；收起状态存在布局的 `LeftClosed/RightClosed/BottomClosed` 里。中心区不能收起。收起最大化面板所在的侧栏会先退出最大化。
