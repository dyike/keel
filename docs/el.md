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
| `el.Input()` / `el.TextArea()` | 输入框，带默认边框样式，获得焦点时边框变蓝。单行框默认高 `theme.ControlHeight`，文字垂直居中，和 kit 的字段一致；点内边距也会聚焦 |
| `el.Widget(w)` | 嵌入任意 `core.Widget`，比如用 `core.Func` 包起来的一段 Gio 布局 |

输入框：

```go
el.Input().ID("q").Placeholder("搜索").Bind(&v.query).OnChange(func(s string) { v.refresh() }).OnSubmit(v.search)
```

点击空白处会让输入框失去焦点。

`Bind(&字符串)` 双向绑定：用户输入会写进变量，程序改了变量，下一帧输入框也跟着变。`Password()` 遮盖内容。`MaxLen(n)` 限制字符数，`Filter("0123456789")` 只接受这些字符（输入和粘贴都会过滤），`ReadOnly(true)` 允许选择复制但不能编辑。单行输入框设置 `OnKey` 后，↑ ↓ PageUp PageDown 先交给它处理，编辑器不再收到这几个键（单行框里它们本来只能把光标移到开头或结尾）；带 Shift 等修饰键的组合仍归编辑器，用来扩展选区。返回值不影响结果，这几个键总是被拿走。输入框位于 `Disabled(true)` 的子树里时不能编辑，`el.Widget` 嵌入的 Gio 代码也一样。

输入框外面再包一层框（带图标、按钮的搜索框）时，给外框设 `ID` 和 `FocusOnPress(输入框ID)`：点外框里没有子元素接住的地方会聚焦输入框，鼠标显示文字光标；外框不会因此变成按钮，也不进 Tab 顺序。

## 样式方法

所有元素共享同一套方法（`Styled[T]` 泛型实现，链式调用返回原来的类型）：

| 分类 | 方法 |
| --- | --- |
| 方向与对齐 | `Row()`、`Col()`（默认）、`Wrap()` 自动换行、`Grid(columns)` 等宽列网格、`ColSpan(n)` 跨列、`Gap(dp)`、`Justify(Start/Center/End/SpaceBetween/SpaceAround)`、`Items(Start/Center/End/Stretch)`、`Center()` |
| 伸缩 | `Grow()` 等于 CSS 的 `flex: 1`：初始尺寸按 0 算，分享剩余空间；其他元素空间不够时按比例收缩，`NoShrink()` 禁止收缩 |
| 尺寸 | `W(l)`、`H(l)`、`Size(l)`、`MinW/MinH/MaxW/MaxH(l)`、`WFull()`、`HFull()`；长度用 `el.Dp(40)`、随字号缩放的 `el.Sp(40)`、`el.Frac(0.5)`、`el.Full` |
| 间距 | `P`、`Px`、`Py`、`Pt`、`Pb`、`Pl`、`Pr`（内边距），`M`、`Mx`、`My`、`Mt`、`Mb`（外边距），单位 dp |
| 滚动与定位 | `ScrollX()` 横向滚动（需要约束宽度）、`ScrollY()` 纵向滚动（需要确定的高度），`StickToBottom()` 跟随到底，`ScrollToEndOn(v)` 在 v 变化时跳到底部；`Absolute()` + `Top/Right/Bottom/Left` 绝对定位，同时给左右会拉伸宽度 |
| 外观 | `Bg(c)`、`Border(dp, c)`、`Rounded(dp)`（用 `theme.RadiusSm/Md/Lg/Xl/Full`）、`Shadow(theme.ElevationSm/Md/Lg)` 阴影画在元素外面、不改变尺寸，`Opacity(a)` 设置整个子树 0–1 透明度（0 完全不绘制，仍保留布局和交互），`CursorPointer()`、`Hidden(b)`、`IsHidden()`（读取本元素声明的隐藏值，不包含祖先） |
| 文字（向下继承） | `TextColor(c)`、`TextSize(sp)`（用 `theme.TextXs` … `theme.TextHeading`）、`Bold()`、`Medium()`、`Mono()` 等宽字体（`theme.MonoFace`）、`LineHeight(倍数)`、`MaxLines(n)` |
| 状态变体 | `Hover(func(*el.Style))`、`Active(func(*el.Style))`：悬停、按下时的颜色变化，背景在 120ms 内渐变过去；开启减少动画（`theme.SetReducedMotion`，自动化模式默认开启）时直接切换 |
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

### 动作与键位表

要让用户改键，就不要在视图里写死按键，改成命名动作：

```go
core.Bind("editor.save", "mod+s")      // 启动时设默认键，可以给多个
core.LoadKeymap(userJSON)              // 叠加用户的 {"editor.save": ["ctrl+alt+s"]}

func (v *editor) Render(cx *el.Context) el.Element {
    cx.Action("editor.save", v.save)   // 绑定到这个动作的每个键都会触发
    ...
}
```

- 键位表是全局的，`core.Bind` 替换一个动作的全部按键，不传按键就是解绑；按键写错时返回错误，什么都不改。
- 改键后所有窗口重绘，下一帧起生效；`cx.Action` 每帧按当前键位表注册，不用重启。
- `core.Bindings(name)` 查一个动作的按键，`core.Keymap()` 返回全部，可以用来做快捷键设置页。
- 显示按键用 `kit.KbdFor(name)` 和 `Menu.ActionItem`，它们跟着键位表变。

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

`FocusStyle(func(*el.Style))` 是绘制样式，可改背景、边框色和文字色，不改变尺寸。普通元素通过 Tab、方向键、Space / Enter 或程序主动聚焦时，默认使用 2dp Primary 焦点边框。鼠标点击保留实际焦点和键盘操作能力，但不绘制焦点样式；点击回调中的同步 `cx.Focus` 也遵循此规则。输入框聚焦时始终沿用自身边框。文字色传递给未显式设置颜色的子元素，失焦后恢复。

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

`Scrollbars(el.ScrollbarAlways / el.ScrollbarHover / el.ScrollbarScrolling)` 设置单个滚动容器的显示策略，两轴共用；默认 Always，只有内容溢出才出现。Hover 在指针进入整个视口时显示；Scrolling 在偏移实际变化后显示，停止 900ms 后隐藏，鼠标拖动滚动条期间持续显示。程序定位、键盘滚动也会显示；停在边界且偏移不变不会重新计时。隐藏后不保留滚动条点击区域，内容仍能接收指针事件；滚轮和键盘滚动不受策略影响。`ScrollOffset` 的受控模式仍隐藏所有滚动条。当前是组件配置，没有自动读取系统滚动条偏好或渐隐动画。已验证横纵双倍率交互、空闲隐藏后的点击穿透、拖出视口时继续拖动和显示策略切换的窗口像素。

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

`Anchored(anchorID, content)` 使用本帧锚点位置，锚点可在主树或先声明的浮层中。Placement 的方向为 Bottom / Top / Left / Right，对齐为 Start / Center / End，默认 Bottom / Start；Offset 默认 4dp。指定方向放不下、对侧放得下时翻转，再将位置平移到 root 内；超出部分按 root 裁剪。MatchAnchorWidth 让浮层与锚点等宽（下拉框与触发器同宽），内容按这个宽度换行或截断。锚点不存在或隐藏时不绘制，并调用一次 OnDismiss。

非模态浮层之外、且不在锚点上的按下事件会请求关闭，并继续传给下面的元素。`.Modal()` 使锚定浮层拦截外部点击；`el.Modal(content)` 创建默认居中的模态浮层，自带遮罩和焦点约束，遮罩在绘制时读取 `theme.Scrim`，随运行时主题切换更新；`.Scrim(false)` 只隐藏遮罩颜色，仍拦截输入。模态期间背景不响应悬停和点击，Agent 快照也不列出被遮挡的主树及下层浮层。

`layer.Owner(id)` 把浮层生命周期绑定到本帧树中的启用元素。可用于模态组件：返回一个带 ID 的零尺寸 Absolute 元素作为所属标记，再给 Modal 设置 Owner；祖先禁用、隐藏或移除标记时会请求关闭。Owner 不改变定位，也不会把浮层自身的模态遮挡当成禁用。

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

## 局部主题

`cx.Themed(palette, view)` 用另一套调色板渲染一个视图：它渲染和绘制时都换用这套颜色，所以里面的 kit 组件和自己画的内容都跟着变，窗口其余部分仍用全局主题。适合浅色窗口里的深色侧栏、主题预览。返回的盒子会拉伸子元素，要铺满颜色就给它设背景。

```go
nord, _ := theme.Named("nord")
cx.Themed(nord, sidebar).Bg(nord.Bg)
```

用 `cx.Cache` 缓存的元素按全局主题版本失效，局部主题里的内容不要跨主题复用缓存。

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

- 布局是 flexbox 的子集：支持 wrap 和简单 grid；尚无 `align-self`、内容尺寸的最小值（min-content）。收缩按内容宽度比例分配。
- `ScrollY` 里的子元素每帧都布局（看不见的不绘制）。内容不变的部分用 `cx.Cache` 跳过重建和重排；几百行以上用 `kit.VirtualList` 或 `kit.Table`，只布局可见行。
- 没有过渡动画的封装，需要自己用 `Now` / `Animating` 计算。
- 浮层只在 `el.Root` 中完整支持，`el.Embed` 按嵌入约束尽力支持。
- 浮层不支持跨窗口。


在 `Decorate` 内可用 `cx.LayoutSize(element)` 读取同一棵树中元素的最终宽高（dp），包括被裁剪的行。布局前不可读取。虚拟列表可用 `cx.ScrollTo(id, offset)` 在下次绘制时设置纵向偏移，按新的内容尺寸裁剪，用于内容变化后保持锚点；首次绘制前和只读布局时不生效。

`cx.AfterEnabled(id, key, delay, fn)` 把定时器绑定到指定元素：元素可见且未禁用时才运行；元素或祖先禁用、隐藏或被模态层遮挡后暂停，恢复时重新等待完整 delay。与 `After` 一样每帧声明，省略声明会取消。

`cx.Enabled(id)` 查询最近声明的元素是否可接收输入，包含祖先禁用和模态层阻挡；找不到 ID 时返回 false。在 `Render` 中查询的是上一轮声明，与焦点查询的时机一致。


`Wrap()` 从左到右排布，宽度不足时另起一行。`Gap` 同时作用于行和行内元素；`Grow/Flex` 在各自行内分配剩余宽度，`Justify` 对齐每行，`Items` 对齐同一行内的不同高度元素。无宽度约束时不会换行。

`Grid(columns)` 按行填充指定数量的列，`Gap` 设置行列间距。列默认等宽，但会先满足子元素的固定宽度及最小宽度；若所有最小宽度之和超过可用空间，保留最小宽度并溢出。每行按最高元素确定高度，自动高度的元素默认拉伸到行高。`ColSpan(n)` 让子元素跨列，限制在 1 到父网格列数，放不下时从新行开始；跨度最小宽度分配到覆盖的列。隐藏和绝对定位元素不占网格单元。不支持跨行或命名区域。`Row`、`Col`、`Wrap`、`Grid` 会切换布局模式。

验证：`go run ./examples/components -section layout`，调整窗口宽度检查换行和网格。

`PinLeft(offset)` / `PinRight(offset)` 将元素绘制在最近 `ScrollX` 视口对应边缘的 offset dp 处，保留布局占位。固定元素最后绘制；普通兄弟元素裁剪到两侧固定元素之间，裁剪同时约束点击和语义区域。两侧宽度超过视口时左侧优先。没有横向滚动祖先时保持普通布局，用于表格冻结列等场景。

`cx.ClickModifiers()` 仅在指针点击/双击回调中返回该事件的 Shift、Ctrl、Command 等修饰键；回调外为零。键盘事件直接使用 `KeyEvent.Modifiers`。

`OnContextMenu(fn)` 在次键按下时调用，不吞掉主键操作；回调可用 `ClickModifiers`。键盘入口通过 `OnKey` 声明，例如 Shift+F10。锚定浮层在锚点或其祖先禁用、隐藏后关闭，不把浮层自身对背景的输入阻挡视作禁用。

拖动结束时 `DragEvent.Canceled` 区分取消与正常释放。需要在松手后提交变更的组件应在取消时丢弃暂存结果。

`cx.ViewportSize()` 在 Render 阶段返回根视口可用宽高（dp），用于限制命令面板等窗口内浮层的高度，避免键盘滚动目标位于窗口之外。

`cx.Countdown(id, key, duration, paused, fn)` 声明保留剩余时间的一次性倒计时。显式 paused、所属元素不可见/禁用/被模态遮挡时暂停，恢复后继续剩余时间；改 duration 重启，省略声明取消。与 `AfterEnabled` 恢复后重新等待完整延迟的语义不同，通知倒计时用 Countdown，悬停提示延迟继续用 AfterEnabled。

`element.Reveal(fraction)` 按 0–1 比例揭示自然高度，保留子元素完整排版，同时裁剪绘制与输入区域；0 时不占高度且不能获得焦点。用于折叠动画，动画时间仍由组件根据 `cx.Now()` 驱动。NaN 按 0，越界值限制到 0–1。

### 鼠标按下监听

`OnMousePress(button, fn)` 观察左键、右键或中键按下，使用 `pointer.ButtonPrimary/Secondary/Tertiary`。监听覆盖交互子元素，但不阻止它们接收事件，也不增加 Tab 停靠点；禁用容器会禁用监听。0 清除监听，非法按键值忽略，多键同时按下不触发。它与 `OnContextMenu(fn)` 共用一个处理器，后者等同于选择右键，最后设置者生效。键盘操作继续使用 `OnKey` 或子组件回调。

锚定浮层可用 `el.Anchored(...).Arrow(true)` 绘制 6dp 指示箭头，跟随实际弹出方向；Offset 测量到箭头尖端。箭头取面板纯色背景，未设置时使用主题 Surface，渐变、边框和阴影不延伸到箭头。Modal 不显示箭头。

### 显式 Tab 顺序

`TabStop(false)` 跳过顺序遍历，保留鼠标/程序聚焦；`TabIndex(n)` 按升序排列，相同值保持树顺序，负值跳过，默认 0。显式配置出现时，el root 处理 Tab/Shift+Tab，在当前模态或 TrapFocus 浮层内循环；否则使用 Gio 原生顺序。输入框也参与排序。禁用、隐藏和未绘制的节点跳过。

范围限单个 el root，不跨独立 Embed 或原生 Gio 控件。直接调用 Router.MoveFocus 绕过此规则，应使用正常 Tab 事件；控件显式消费 Tab 时保留其操作行为。

`WrapFit()` 与 Wrap 一样支持换行；自动宽度时按每行内容收紧，适合按钮胶囊等需要贴合内容的容器。显式或拉伸宽度仍使用常规行对齐及 Grow 分配。调用 Wrap() 恢复填满可用行宽的默认行为。

### 受控滚动与上一帧尺寸

`ScrollX/ScrollY` 可配合 `ScrollOffset(x, y)` 使用绝对 dp 偏移；绘制时按内容边界限制，禁用时仍显示指定位置。指定偏移时停用默认滚动手势并隐藏滚动条，容器可自行使用 OnDrag；省略即可恢复普通滚动；非有限值忽略。应由应用状态持续提供目标位置。

`cx.LastSize(id)` 在 Render 中读取同 root 上次绘制的元素尺寸（dp），首次或被裁掉时返回零；包含禁用帧的几何更新。`cx.LayoutSize(element)` 用于 Decorate 中读取当前布局尺寸。窗口尺寸变化后，依赖 LastSize 的布局通常需再绘制一帧收敛。

`cx.PixelScale()` 返回当前每 dp 的物理像素数，可在 Render 中用于与布局一致的像素舍入，未设置时为 1。

### 自定义滚动事件

`OnScroll(xRange, yRange, fn)` 接收 dp 单位的 ScrollEvent，范围使用 `el.ScrollRange{Min: ..., Max: ...}`。零范围不接收该轴，超出范围的位移由 Gio 路由给外层；子滚动区域优先。nil 移除回调，禁用/隐藏祖先阻止事件。范围按像素尺度换算，非有限端点按 0 处理，端点限制在 ±1,000,000dp。

此事件不暴露滚轮/触控板类型或手势结束相位。可用于受控 ScrollOffset 容器；处理时更新应用状态，再由下一帧应用偏移。

### 指定内容底边对齐

横向容器使用 `Items(el.ContentBottom)`，可把子项对齐到指定后代的底边。子项用 `.ContentBottom(target)` 指定当前渲染树中的正常流后代；不指定、目标被隐藏或不在子树中时，回退到子项自身底边。这是几何对齐，不是字体基线。

```go
body := el.Div().Child(header, content, footer).ContentBottom(content)
row := el.Div().Row().Items(el.ContentBottom).Child(avatar, body)
```

布局使用当前帧尺寸计算对齐线以上和以下所需空间，支持 Row 和 Wrap 的每一行。目标不接受绝对定位节点；容器显式限高时仍遵守限高。此模式不用于纵向容器或 Grid。

`el.Input().SelectOnFocus(true)` 在获得焦点时选中全部内容。`CaptureKeys(names...)` 让单行输入的 OnKey 提前接收指定的无修饰键；这些键由回调完全负责，返回 false 也不会交还编辑器，带修饰键的快捷键不受影响。`cx.SelectInput(id, start, end)` 在下一次 Bind 同步后设置 rune 选区，不改变焦点或文字；端点由编辑器限制到有效范围，缺失或禁用输入忽略。上述接口用于 TimeField 的快速分段编辑，已验证中文 rune 选区、范围限制及带修饰键的编辑行为。

`el.Input().TransformEdit(func(before, after el.InputEdit) el.InputEdit)` 在编辑归一化时同时提供编辑前后的文本及 rune 选区，适合格式掩码判断删除方向。它与 Transform 互斥，后配置者生效；撤销重做保存归一化后的文本与选区，程序 Bind 更新仍清空历史。已通过掩码删除和撤销重做回归。

`cx.InputSelection(id)` 返回最近的编辑器文本和 rune 选区。`cx.InputAction(id, el.InputCopy / InputCut / InputPaste / InputSelectAll)` 将编辑命令排入该输入下一次绘制；缺失/禁用输入忽略，只读拒绝剪切/粘贴，密码输入拒绝命令复制/剪切。Paste 走异步系统文本剪贴板，继续由编辑器完成过滤和 Transform；命令本身不改变焦点，菜单调用方可用 cx.Focus 恢复输入焦点。已通过菜单剪切、焦点恢复和受限输入回归。

`ContainerContentSize(element)` 可在 `Decorate` 中读取容器布局后的内容像素尺寸，数值在该容器自身的最小/最大尺寸限制及 `Reveal` 之前计算；文本、输入和 widget 叶子返回零。

`ElementBounds(root, target)` 在布局后返回目标相对根元素的边框位置，包含离屏元素，不叠加滚动偏移；隐藏或不属于该树的目标返回 false。可在 `Decorate` 中结合 `ScrollTo` 实现离屏定位。
