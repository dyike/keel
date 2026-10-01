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

和 `ui/widget` + `ui/layout` 的写法比：

| | widget + layout | el |
| --- | --- | --- |
| 交互状态 | 每个组件自己声明 `widget.Clickable` 等字段 | 框架按元素位置（或 `ID`）自动保存，不用声明 |
| 布局 | `layout.Column`/`Row` 和少量参数 | flexbox：内外边距、间距、伸缩、对齐、百分比尺寸、滚动、绝对定位 |
| 样式 | 组件内部写死，靠 theme 变量 | 每个元素都能链式设置，悬停、按下有样式变体 |
| 事件结果 | 回调后下一帧才画出 | 先处理事件再渲染，同一帧就画出 |
| Agent 语义 | 组件作者手动声明 | 自动：有 `OnClick` 的是按钮，文字是文本，`Input` 是输入框 |

两套写法可以混用，见下文"和 ui/widget 混用"。新界面优先用 `el`。

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
| `el.Widget(w)` | 嵌入任意 `core.Widget`，比如 `widget.Table` |

输入框：

```go
el.Input().ID("q").Placeholder("搜索").Bind(&v.query).OnChange(func(s string) { v.refresh() }).OnSubmit(v.search)
```

点击空白处会让输入框失去焦点。

`Bind(&字符串)` 双向绑定：用户输入会写进变量，程序改了变量，下一帧输入框也跟着变。`Password()` 遮盖内容。

## 样式方法

所有元素共享同一套方法（`Styled[T]` 泛型实现，链式调用返回原来的类型）：

| 分类 | 方法 |
| --- | --- |
| 方向与对齐 | `Row()`、`Col()`（默认）、`Gap(dp)`、`Justify(Start/Center/End/SpaceBetween/SpaceAround)`、`Items(Start/Center/End/Stretch)`、`Center()` |
| 伸缩 | `Grow()` 等于 CSS 的 `flex: 1`：初始尺寸按 0 算，分享剩余空间；其他元素空间不够时按比例收缩，`NoShrink()` 禁止收缩 |
| 尺寸 | `W(l)`、`H(l)`、`Size(l)`、`MinW/MinH/MaxW/MaxH(l)`、`WFull()`、`HFull()`；长度用 `el.Dp(40)`、`el.Frac(0.5)`、`el.Full` |
| 间距 | `P`、`Px`、`Py`、`Pt`、`Pb`、`Pl`、`Pr`（内边距），`M`、`Mx`、`My`、`Mt`、`Mb`（外边距），单位 dp |
| 滚动与定位 | `ScrollY()` 纵向滚动（需要确定的高度），`StickToBottom()` 跟随到底，`ScrollToEndOn(v)` 在 v 变化时跳到底部；`Absolute()` + `Top/Right/Bottom/Left` 绝对定位，同时给左右会拉伸宽度 |
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

### 复制到剪贴板

`el.WriteClipboard(text)` 在回调里调用，当前帧写入系统剪贴板。

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

## 和 ui/widget 混用

- **在 el 里用旧组件**：`el.Widget(table)` 嵌入。表格、下拉框、单选、开关现在就是这样用的，见 `examples/orders`。
- **在旧布局里用 el**：`el.Embed(view)` 得到一个按内容定尺寸的 `core.Widget`。
- **整个窗口用 el**：`window.Options{Content: el.Root(view)}`。`Root` 占满窗口，窗口不再加边距和外层滚动；页面要滚动时，给根 `Div` 加 `ScrollY()`。
- **对话框**：`widget.Dialog` 照旧放在 `window.Options.Overlay`。

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

- 布局是 flexbox 的子集：没有换行（wrap）、网格、`align-self`、横向滚动、内容尺寸的最小值（min-content）。收缩按内容宽度比例分配。
- 没有虚拟列表：`ScrollY` 里的子元素每帧都布局（看不见的不绘制）。内容不变的部分用 `cx.Cache` 跳过重建和重排；几百行的表格用 `el.Widget(widget.Table)`。
- 没有动画和过渡效果。


主题切换时 `theme.Apply` 会使 `cx.Cache` 的元素在下次访问时重建。自建缓存需要包含 `theme.Revision()`；主题色应在 Render 或缓存构建函数内读取，固定颜色不会自动转换。


## 焦点、按键与禁用（E1 / E2）

```go
el.Div().ID("save").Focusable().
    OnClick(save).
    FocusStyle(func(s *el.Style) { s.BorderColor(theme.Primary) }).
    Child(el.Text("保存"))
```

`Focusable()` 让元素接受点击焦点，并按绘制顺序参与 Tab / Shift+Tab 导航，与 `Input`、`TextArea` 共用原生焦点顺序。隐藏、移除和完全滚出绘制区域的节点不参与导航。带 OnClick 的元素默认可聚焦；Focusable(false) 显式退出 Tab 顺序。

聚焦元素收到无修饰键的 Space / Enter 时，在匹配的按键释放事件中调用一次 `OnClick`；失去焦点后不保留待激活按键。`OnKey(func(el.KeyEvent) bool)` 接收按下和释放事件，从聚焦元素向有处理器的祖先冒泡。返回 `true` 会停止冒泡并取消默认激活。`KeyEvent` 是 el 自己的结构体，包含 string 类型的 Name、KeyPress/KeyRelease 状态和 key.Modifiers。Tab 保留原生导航行为；全局快捷键继续使用 `cx.Shortcut`。

`FocusStyle(func(*el.Style))` 是绘制样式，可改背景、边框色和文字色，不改变尺寸。普通元素默认使用 2dp Primary 焦点边框；输入框沿用自身边框。文字色传递给未显式设置颜色的子元素，失焦后恢复。

`cx.Focus("save")` 在本帧绘制后请求焦点，也支持带 ID 的 `Input` / `TextArea`。ID 应在当前 root 内唯一；重复时选择第一个已绘制的匹配目标。目标不存在、隐藏或完全在视口外时保留原焦点；`cx.Focus("")` 清除焦点。只能在 Render 或其事件回调里调用。

当前 `OnKey` 冒泡源是显式 `Focusable` 元素；输入框编辑按键仍由 Gio editor 处理，不通过这条冒泡链。焦点陷阱留给浮层阶段。可运行 `go run ./examples/components -section focus` 验证接口。


`cx.Focused(id)` 读取当前焦点，查询支持普通元素和输入框。`Disabled(true)` 自上而下禁止子树的点击、悬停和按键，释放当前焦点，禁止程序聚焦，并将 Agent 语义标记为 disabled；解除禁用后可重新聚焦。`DisabledStyle(func(*el.Style))` 设置禁用外观，默认文字为 Muted。显式禁用与测量时没有输入源分开处理，连续测量不会清空交互状态。Focus 保留为 FocusStyle 的兼容别名。

## 时间与减少动画（E3）

`cx.Now()` 返回本帧时间；动画只从它计算相位。`cx.Animating()` 请求下一帧，不创建 goroutine。`el.ReducedMotion()` 查询应用偏好；`theme.SetReducedMotion(true)` 在帧锁内切换。当前没有原生系统偏好桥接，默认 false。

`cx.After(duration, callback)` 是声明式的一次性定时器：存活期间每次 Render 都声明，触发后不会自动重启；某帧不再声明即取消。重新声明或修改 duration 创建新的计时周期。回调在帧锁内、完成当前树绘制后运行，并请求重绘。

重复组件需要稳定身份，使用 `cx.Scope(id)` 绑定返回树中的元素 ID；该节点隐藏或移除时自动取消。不同组件不要共享 ID。同一 scope 中按调用位置识别定时器，同一调用位置循环声明时按顺序区分，因此动态列表应给每项独立 Scope。

```go
scope := cx.Scope("notice")
scope.After(3*time.Second, func() { visible = false })
return el.Div().ID("notice").Hidden(!visible).Child(el.Text("已保存"))
```

没有 Scope 的 After 属于当前 root，停止声明时取消。Scope 用于定时器归属；全局快捷键仍在原始 cx 上注册。
