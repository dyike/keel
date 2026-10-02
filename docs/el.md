# 元素与视图（ui/el）

`ui/el` 是 GPUI 风格的写法：界面是一个视图（普通的 Go struct），每一帧调用它的 `Render`，返回一棵用链式样式搭起来的元素树。状态就是 struct 的字段，事件回调直接改字段，下一帧自动重画。

```go
type Counter struct{ n int }

func (c *Counter) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Items(el.Start).Child(
        el.Text(fmt.Sprintf("点了 %d 次", c.n)).TextSize(20).Bold(),
        el.Div().Px(16).Py(8).Rounded(6).Bg(theme.Primary).TextColor(theme.OnColor).
            CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.PrimaryHover) }).
            OnClick(func() { c.n++ }).
            Child(el.Text("+1")),
    )
}

window.Open(window.Options{Title: "计数", Content: el.Root(&Counter{})})
window.Main()
```

和直接写 Gio 比：

| | Gio | el |
| --- | --- | --- |
| 交互状态 | 每个组件自己声明 `widget.Clickable` 等字段 | 框架按元素位置（或 `ID`）自动保存，不用声明 |
| 布局 | 层层嵌套 `layout.Flex{}.Layout(gtx, layout.Rigid(...))` | flexbox：内外边距、间距、伸缩、对齐、百分比尺寸、滚动、绝对定位 |
| 样式 | 每次绘制手写 | 每个元素都能链式设置，悬停、按下有样式变体 |
| 事件结果 | 处理顺序取决于布局顺序 | 先处理事件再渲染，同一帧就画出 |
| Agent 语义 | 手动声明 | 自动：有 `OnClick` 的是按钮，文字是文本，`Input` 是输入框 |

现成组件在 [ui/kit](kit.md)，它们都是 el 视图。

## 一帧里发生了什么

```
1. 分发事件   上一帧登记过的点击 → 调用对应元素的 OnClick
2. 渲染       view.Render(cx) → 元素树（按当前状态，每帧新建）
3. 布局       flexbox 引擎算出每个元素的位置和尺寸
4. 绘制       背景、边框、文字；登记点击区域；生成语义信息
5. 回收       这一帧没出现的元素，状态删掉
```

Render 每帧都会调用，要保持便宜：只根据状态搭树，不做 I/O、不做大计算。耗时的事放 goroutine，完成后用 `core.Update` 改状态。线程规则和其他模块相同，见[架构 · 线程规则](architecture.md#线程规则)。

## 元素

| 构造 | 说明 |
| --- | --- |
| `el.Div()` | 盒子，唯一能有子元素的元素。默认子元素从上到下排列 |
| `el.Text(s)` | 文字，按可用宽度自动换行 |
| `el.Input()` / `el.TextArea()` | 输入框，带默认边框样式，获得焦点时边框变蓝 |
| `el.Widget(w)` | 嵌入任意 `core.Widget`，比如用 `core.Func` 包起来的一段 Gio 布局 |

输入框：

```go
el.Input().ID("q").Placeholder("搜索").Bind(&v.query).OnChange(func(s string) { v.refresh() }).OnSubmit(v.search)
```

点击空白处会让输入框失去焦点。

`Bind(&字符串)` 双向绑定：用户输入会写进变量，程序改了变量，下一帧输入框也跟着变。`Password()` 遮盖内容。`MaxLen(n)` 限制字符数，`Filter("0123456789")` 只接受这些字符（输入和粘贴都会过滤），`ReadOnly(true)` 允许选择复制但不能编辑。单行输入框设置 `OnKey` 后，↑ ↓ PageUp PageDown 先交给它处理，编辑器不再收到这几个键（单行框里它们本来只能把光标移到开头或结尾）；带 Shift 等修饰键的组合仍归编辑器，用来扩展选区。返回值不影响结果，这几个键总是被拿走。输入框位于 `Disabled(true)` 的子树里时不能编辑，`el.Widget` 嵌入的 Gio 代码也一样。

## 样式方法

所有元素共享同一套方法（`Styled[T]` 泛型实现，链式调用返回原来的类型）：

| 分类 | 方法 |
| --- | --- |
| 方向与对齐 | `Row()`、`Col()`（默认）、`Gap(dp)`、`Justify(Start/Center/End/SpaceBetween/SpaceAround)`、`Items(Start/Center/End/Stretch)`、`Center()` |
| 伸缩 | `Grow()` 等于 CSS 的 `flex: 1`：初始尺寸按 0 算，分享剩余空间；其他元素空间不够时按比例收缩，`NoShrink()` 禁止收缩 |
| 尺寸 | `W(l)`、`H(l)`、`Size(l)`、`MinW/MinH/MaxW/MaxH(l)`、`WFull()`、`HFull()`；长度用 `el.Dp(40)`、`el.Frac(0.5)`、`el.Full` |
| 间距 | `P`、`Px`、`Py`、`Pt`、`Pb`、`Pl`、`Pr`（内边距），`M`、`Mx`、`My`、`Mt`、`Mb`（外边距），单位 dp |
| 滚动与定位 | `ScrollX()` 横向滚动（需要约束宽度）、`ScrollY()` 纵向滚动（需要确定的高度），`StickToBottom()` 跟随到底，`ScrollToEndOn(v)` 在 v 变化时跳到底部；`Absolute()` + `Top/Right/Bottom/Left` 绝对定位，同时给左右会拉伸宽度 |
| 外观 | `Bg(c)`、`Border(dp, c)`、`Rounded(dp)`、`CursorPointer()`、`Hidden(b)` |
| 文字（向下继承） | `TextColor(c)`、`TextSize(sp)`、`Bold()`、`MaxLines(n)` |
| 状态变体 | `Hover(func(*el.Style))`、`Active(func(*el.Style))`：悬停、按下时的颜色变化 |
| 交互 | `OnClick(fn)`、`OnDoubleClick(fn)` |
| 结构 | `ID(s)`、`Child(...)`、`Children(slice)`、`When(cond, func(*T))` |
| Agent 语义 | `Role(s)`、`Name(s)`、`Value(s)`、`Selected(b)` |

布局规则按 flexbox 的直觉来，和 CSS 有几处不同：

- `Div` 默认纵向排列、子元素拉满宽度（像块级元素），`Row()` 改成横向。
- 横向排列的子元素宽度取内容宽度；空间不够时一起按比例收缩，文字随之换行。
- `ScrollY` 的视口是内边距盒，内边距跟着内容滚动。
- 文字样式像 CSS 一样向子元素继承。

## 状态和 ID

元素的内部状态（悬停、按下、滚动位置、输入框内容和光标）由框架保存，键是元素在树上的路径：每一层取它的 `ID`，没有就取它在兄弟中的序号。

**会增删、重排的列表项要给 `ID`**，否则状态会按位置错配（比如删掉第一行，第二行的输入框内容跑到第一行）：

```go
el.Div().Children(el.Map(v.rows, func(i int, r Row) el.Element {
    return el.Div().ID(r.ID).OnClick(func() { v.open(r) }).Child(el.Text(r.Name))
}))
```

元素某一帧没有出现，它的状态就被删除。隐藏再显示的输入框会清空；要保留内容，用 `Bind` 把内容放在视图的字段里。

## 视图与组合

视图是任何实现了 `Render(*el.Context) el.Element` 的类型。拆分界面就是拆分 struct，在父视图的 Render 里调用子视图的 Render：

```go
type Page struct{ list *OrderList; form *OrderForm }

func (p *Page) Render(cx *el.Context) el.Element {
    return el.Div().Row().Gap(16).Child(
        el.Div().W(el.Dp(300)).Child(p.list.Render(cx)),
        el.Div().Grow().Child(p.form.Render(cx)),
    )
}
```

`el.ViewFunc(func(cx *el.Context) el.Element { … })` 将函数适配成 View，适合传给 kit 的内容插槽。插槽持有 View，每帧调用 Render；主题色在函数内读取，不要在构造视图时保存 Element。

`cx.Shortcut("mod+s", fn)` 在视图渲染期间绑定快捷键。

### 缓存不变的部分

长列表、聊天记录里大部分内容每帧都不变。`cx.Cache(key, build)` 在 key 不变时直接复用上一帧的元素和布局，`build` 不会被调用：

```go
for _, m := range v.msgs {
    m := m
    list.Child(cx.Cache(msgKey{m.id, m.version}, func() el.Element { return v.message(m) }))
}
```

key 必须能比较，而且元素的样子只由 key 决定：外观会变时 key 也要变（比如带上版本号）。某一帧没用到的缓存项会被删掉。`ui/markdown` 就是这样缓存写完的块的。

主题切换时 `theme.Apply` 会使 `cx.Cache` 的元素在下次访问时重建。自建缓存需要包含 `theme.Revision()`；主题色应在 Render 或缓存构建函数内读取，固定颜色不会自动转换。

### 复制到剪贴板

`el.WriteClipboard(text)` 在回调里调用，当前帧写入系统剪贴板。

## 焦点、按键与禁用（E1 / E2）

```go
el.Div().ID("save").Focusable(true).
    OnClick(save).
    FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
    Child(el.Text("保存"))
```

`Focusable(true)` 让元素接受点击焦点，并按绘制顺序参与 Tab / Shift+Tab 导航，与 `Input`、`TextArea` 共用原生焦点顺序。隐藏、移除和完全滚出绘制区域的节点不参与导航。带 OnClick 的元素默认可聚焦；Focusable(false) 显式退出 Tab 顺序。

聚焦元素收到无修饰键的 Space / Enter 时，在匹配的按键释放事件中调用一次 `OnClick`；失去焦点后不保留待激活按键。`OnKey(func(el.KeyEvent) bool)` 接收按下和释放事件，从聚焦元素向有处理器的祖先冒泡。返回 `true` 会停止冒泡并取消默认激活。`KeyEvent` 是 el 自己的结构体，包含 string 类型的 Name、KeyPress/KeyRelease 状态和 key.Modifiers。Tab 保留原生导航行为；全局快捷键继续使用 `cx.Shortcut`。

`FocusStyle(func(*el.Style))` 是绘制样式，可改背景、边框色和文字色，不改变尺寸。普通元素默认使用 2dp Primary 焦点边框；输入框沿用自身边框。文字色传递给未显式设置颜色的子元素，失焦后恢复。

`cx.Focus("save")` 在本帧绘制后请求焦点，也支持带 ID 的 `Input` / `TextArea`。ID 应在当前 root 内唯一；重复时选择第一个已绘制的匹配目标。目标不存在、隐藏或完全在视口外时保留原焦点；`cx.Focus("")` 清除焦点。只能在 Render 或其事件回调里调用。

当前 `OnKey` 冒泡源是可聚焦的普通元素；输入框编辑按键仍由 Gio editor 处理，不通过这条冒泡链。焦点陷阱留给浮层阶段。可运行 `go run ./examples/components -section focus` 验证接口。

`cx.Focused(id)` 读取当前焦点，查询支持普通元素和输入框。`Disabled(true)` 自上而下禁止子树的点击、悬停和按键，释放当前焦点，禁止程序聚焦，并将 Agent 语义标记为 disabled；解除禁用后可重新聚焦。`DisabledStyle(func(*el.Style))` 设置禁用外观，默认文字为 Muted。显式禁用与测量时没有输入源分开处理，连续测量不会清空交互状态。

## 时间与减少动画（E3）

`cx.Now()` 返回本帧时间；动画只从它计算相位。`cx.Animating()` 请求下一帧，不创建 goroutine。`el.ReducedMotion()` 查询应用偏好；`theme.SetReducedMotion(true)` 在帧锁内切换。当前没有原生系统偏好桥接，默认 false。

`cx.After(key, duration, callback)` 声明一次性定时器，key 必须可比较且在当前 root 内唯一。存活期间每次 Render 声明同一个 key；某帧没有声明即取消，与调用位置和声明顺序无关。duration 变化时重新计时。触发后持续声明不会重复执行，省略一帧再声明才会重新启动。回调在帧锁内、当前树绘制完成后通过 core.Call 执行，并通知所有窗口重绘。

```go
type noticeTimerKey struct { ID string }
if visible {
    cx.After(noticeTimerKey{noticeID}, 3*time.Second, func() { visible = false })
}
return el.Div().Hidden(!visible).Child(el.Text("已保存"))
```

同类组件用自身稳定 ID 组成 key。不要在 `cx.Cache` 的构建函数里声明 `After`：缓存命中时不会执行构建函数，未再次声明的定时器会被取消。应把 After 放在每次执行的 Render 路径上，再单独缓存元素树。

## 滚动状态与锚定

`ScrollX()` 让子内容横向延展并裁剪到视口，可与 `ScrollY()` 组合。支持水平滚轮和触控板水平手势；横纵轴分别消费对应滚动量。`cx.ScrollStateX(id)` 返回偏移、视口宽度、内容宽度（dp）；`cx.ScrollIntoViewX(id, left, right)` 最小滚动以显示目标区间。横纵滚动条支持拖动滑块、点击轨道定位，并随浅深色主题切换。滚动容器添加 `Focusable(true)` 后可用 Tab 聚焦，再用方向键、PageUp/PageDown、Home/End 滚动；双轴容器的 Shift+PageUp/PageDown、Shift+Home/End 操作横轴。组件自身 `OnKey` 优先处理（如表格行选择）。示例：`go run ./examples/components -section scrollable`。

`cx.ScrollState(id)` 返回带 ID 的 `ScrollY` 元素上一帧的滚动偏移、可视高度、内容高度，单位 dp；第一次绘制之前三个值都是 0。虚拟列表用它决定构建哪些行。`cx.ScrollIntoView(id, top, bottom)` 以最小的滚动量让内容中 `[top, bottom]` 这一段可见，在下一次绘制时生效。

`KeepBottomOn(version)`：`version` 变化的那一帧，保持到底部的距离不变，在上方插入内容（比如加载更早的聊天记录）时画面不会跳动。在插入内容的同一个回调里递增 version，用法和 `ScrollToEndOn` 一样。

`Flex(w)` 和 `Grow` 一样占用剩余空间，但按权重分配：`Flex(2)` 分到的是 `Flex(1)` 或 `Grow` 的两倍。

## 拖动

```go
el.Div().ID("track").W(el.Dp(240)).H(el.Dp(20)).OnDrag(func(e el.DragEvent) {
    v.value = clamp(e.X / e.W) // 按下、移动、松开都会调用
})
```

`DragEvent.Kind` 是 `DragStart`（按下）、`DragMove`（按住移动，指针移出元素也继续报告）、`DragEnd`（松开或取消）。`X`、`Y` 是相对元素左上角的 dp，`W`、`H` 是元素尺寸，所以 `X/W` 就是水平方向的比例。按下时会聚焦可聚焦的元素。禁用的元素收不到拖动。

## 浮层（E4 / E5）

在 Render 中调用 `cx.Overlay(key, layer)` 声明浮层。key 必须可比较，并在当前 root 内唯一。打开状态由视图保存，打开期间每次 Render 都声明；某帧省略就关闭。不要在 Cache 的构建函数里声明浮层。后声明的浮层在上层，Esc 只请求关闭最上层。

```go
if v.open {
    cx.Overlay("filters", el.Anchored("filter-button",
        el.Div().W(el.Dp(280)).P(16).Bg(theme.Surface).Child(
            el.Text("筛选条件"),
            el.Input().ID("query").Bind(&v.query),
        ),
    ).Placement(el.Bottom, el.Start).OnDismiss(func() { v.open = false }))
}
```

`Anchored(anchorID, content)` 使用本帧锚点位置，锚点可在主树或先声明的浮层中。Placement 的方向为 Bottom / Top / Left / Right，对齐为 Start / Center / End，默认 Bottom / Start；Offset 默认 4dp。指定方向放不下、对侧放得下时翻转，再将位置平移到 root 内；超出部分按 root 裁剪。MatchAnchorWidth 将最小宽度设为锚点宽度。锚点不存在或隐藏时不绘制，并调用一次 OnDismiss。

非模态浮层之外、且不在锚点上的按下事件会请求关闭，并继续传给下面的元素。`.Modal()` 使锚定浮层拦截外部点击；`el.Modal(content)` 创建默认居中的模态浮层，自带遮罩和焦点约束，遮罩在绘制时读取 `theme.Scrim`，随运行时主题切换更新；`.Scrim(false)` 只隐藏遮罩颜色，仍拦截输入。模态期间背景不响应悬停和点击，Agent 快照也不列出被遮挡的主树及下层浮层。

`.TrapFocus()` 将 Tab / Shift+Tab 限制在浮层内，出现时聚焦第一个可聚焦元素；同帧 `cx.Focus(id)` 可指定浮层内的目标。关闭后恢复先前焦点，原目标已经移除时清除焦点。未开启焦点约束的非模态浮层不移动焦点。OnDismiss 在帧锁内执行，只通知调用方更新打开状态，不会替调用方保存 open。

`el.Modal(content).Placement(side, align)` 把模态内容贴在 root 的某条边上，而不是居中，用于侧边抽屉：`Placement(el.Right, el.Start)` 贴右边、顶端对齐。

Esc 交给最上层**设置了 OnDismiss** 的浮层。没有 OnDismiss 的浮层（例如通知栈）不会吞掉 Esc，下面的对话框照常关闭。

`cx.FocusWithin(id)` 查询该元素或它的任一子孙是否在上一帧获得焦点，Tooltip 用它在键盘聚焦时显示提示。

`cx.Hovered(id)` 查询最近处理的指针位置是否位于元素内，禁用或被模态层遮挡的元素返回 false。普通带 ID 的元素也可查询，不必添加点击回调。HoverCard 可组合锚点和卡片的 Hovered 结果。

完整浮层能力要求 `el.Root`。`el.Embed` 使用嵌入时的最大约束，通过 `op.Defer` 延后绘制，属于尽力支持；其可用空间不一定等于窗口大小。浮层不跨窗口。无输入源的帧统一按只读帧处理，包括测量和父组件禁用。它们复用真实的 store/cache，保留输入内容和滚动位置；不分发事件、不触发关闭回调、不增减浮层生命周期、不推进定时器，也不清理状态或覆盖焦点恢复记录。

验证：`go run ./examples/components -section overlay`，加 `-theme dark` 检查深色；切换浮层、打开模态、编辑输入框，并用 Tab / Shift+Tab / Esc 检查焦点。

## 做成可复用的组件

组件就是返回 `el.Element` 的函数，参数是它需要的数据和回调：

```go
func button(label string, onClick func()) el.Element {
    return el.Div().ID(label).Px(16).Pt(10).Pb(6).Rounded(6).
        Bg(theme.Primary).TextColor(theme.OnColor).TextSize(14).
        CursorPointer().Hover(func(s *el.Style) { s.Bg(theme.PrimaryHover) }).
        OnClick(onClick).Child(el.Text(label))
}
```

需要自己的状态、而且状态要跨帧保存的组件，写成视图（struct + Render）。

## 放进窗口和嵌入 Gio

- **整个窗口用 el**：`window.Options{Content: el.Root(view)}`。`Root` 占满窗口，窗口不再加边距和外层滚动；页面要滚动时，给根 `Div` 加 `ScrollY()`。浮层（对话框、菜单）用 `cx.Overlay` 声明，不需要 `window.Options.Overlay`。
- **在 el 里放 Gio 代码**：`el.Widget(w)` 嵌入任意 `core.Widget`，比如用 `core.Func` 包起来的一段 Gio 布局。
- **把 el 放进 Gio 布局**：`el.Embed(view)` 得到一个按内容定尺寸的 `core.Widget`。

## Agent 能看到什么

不用额外写代码：

| 元素 | Agent 看到的 |
| --- | --- |
| `Text` | `text`，名字就是文字 |
| 带 `OnClick` 的 `Div` | `button`，名字是里面的文字 |
| `Input` | `textbox`，名字是 `Name` 或占位文字，值是当前内容 |
| `.Role("tab").Selected(true)` | `tab`，选中 |
| `.Role("progressbar").Name("导入").Value("40%")` | `progressbar`，值 40% |

滚动容器外面的内容不会出现在元素列表里。

## 已知限制

- 布局是 flexbox 的子集：没有换行（wrap）、网格、`align-self`、内容尺寸的最小值（min-content）。收缩按内容宽度比例分配。
- `ScrollY` 里的子元素每帧都布局（看不见的不绘制）。内容不变的部分用 `cx.Cache` 跳过重建和重排；几百行以上用 `kit.VirtualList` 或 `kit.Table`，只布局可见行。
- 没有过渡动画的封装，需要自己用 `Now` / `Animating` 计算。
- 浮层只在 `el.Root` 中完整支持，`el.Embed` 按嵌入约束尽力支持。
- 浮层不支持跨窗口。
