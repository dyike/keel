# 常见问题

## 中文显示成方框

Gio 自带的 Go 字体没有中文，中文靠系统字体回退。自动回退有时会选到缺字的字体，比如粗体标题里的"组"显示成方框。

`theme.Face` 固定了字体优先级：苹方 → 冬青黑体 → 微软雅黑 → Noto Sans CJK。在 Linux 上显示方框，说明这几种都没装，装上 Noto CJK 字体即可（Debian/Ubuntu：`fonts-noto-cjk`）。要用别的字体，改 `theme.Face`，把字体族名放在最前面。

## Windows 上 g、y 的下半截被切掉

纯文字元素（没有自定义 Role、Name、选中状态）不再自己画裁剪区，由 Gio 的 Label 报告语义并按字形超出量留边，所以下伸部分不会被切。曾经的写法是按 macOS 苹方调出的固定下移量（字号 × 0.22），苹方行框下方空白多，看不出问题；Windows 的微软雅黑行框贴着字形，下移后 g、y 超出文字框，被文字框的语义裁剪区切掉。回归测试 `TestTextShiftKeepsDescendersInBox` 用 Go 字体在任何系统上复现这种紧凑行框。

## 按钮里的中文偏上

字体的逻辑行框包含 ascent/descent 和留白，可见字形不一定在行框中央。el 的文字按代表字形测量后调整绘制基线（`ui/el/paint.go` 的 `textShift`）：用同字号 Label 的实际行框和"国""A"的字形位置算出下移量，且不让 g 的下伸部分移出行框；数字角标按实际数字字形居中。先区分是容器位置错误，还是字形在容器内偏移，再改对应层。不要把某张截图测得的像素差直接加到所有组件上。

## 输入框光标比文字向下伸出

Gio 默认按字体的 ascent/descent 画光标，某些中文字体的 descent 包含较多空白。`ui/internal/editorstyle` 为两套输入组件统一绘制光标：位置使用编辑器的实际坐标，高度使用中文字形、拉丁大写和下伸字母的可见范围。不要单独平移光标来补偿，否则点击定位、多行和滚动会错位。

## 输入框文字贴着选区上沿，选区下方留白过大

这是逻辑行框和可见字形混用的问题。一次 2× 中文渲染中，字形占第 2–28 像素，Gio 默认选区占第 0–44 像素。只修光标高度不能修正选区；只移动文字又会让文字脱离点击、选词和输入法使用的坐标。

修复入口是 [`ui/internal/editorstyle`](../ui/internal/editorstyle/README.md)，`el.Input`、`el.TextArea`（以及基于它们的 `kit.Input`、`kit.TextArea`）都调用它。处理顺序：

1. 调用者先消费 `Editor.Update`，再用原编辑器完成排版，记录绘制操作。原生选区设为透明。
2. 用 `Editor.Regions` 获取已经考虑换行、双向文本和滚动的选区范围。横向范围直接沿用；每个区域的基线是 `Bounds.Max.Y - Baseline`。
3. 测量选中文字的可见字形上、下沿，相对基线绘制选区，并留 1dp 内边距。纯空白选区使用代表字形的高度；密码框测量掩码。文字或字体变化时重新测量。
4. 将选区裁剪到编辑器视口，先画背景，再回放原文字。保留原编辑器的文字位置、行间距、输入、选区范围和滚动行为。

再次遇到文字对齐问题时，按这个顺序排查：

- **先复现状态**：填写内容、聚焦、选择文字并滚动。空输入框截图不能验证选区和滚动。
- **分开检查坐标**：比较组件边界、逻辑行框、实际字形、光标和选区。容器只在一侧预留角标空间，会让整颗按钮偏移；这类问题应修容器。
- **使用真实字体和比例**：同时测中文、数字、英文下伸字母、混排和密码，覆盖 1× / 2×、空行、软换行、部分可见行。
- **验证实际像素**：用不同颜色绘制文字和选区，检查可见范围和中心，而不只检查返回尺寸或字体参数。测试必须提供 `input.Router.Source()`，否则上下文处于禁用状态，颜色会被淡化。
- **验证交互**：选中后滚动、替换、撤销和输入法组合事件都要保留。模拟文字替换时，`key.EditEvent` 要带选区范围，随后发送更新光标的 `key.SelectionEvent`。

回归入口：

```sh
go test ./ui/internal/editorstyle ./ui/el ./ui/kit -count=1
go run ./examples/components -section textarea
go run ./examples/components -section input
```

像素回归在 `ui/internal/editorstyle/selection_test.go`，光标和输入法回归在 `caret_test.go`。在 TextArea 示例点击“填入多行”后全选，检查每行选区，再选中滚动；在 Input 示例检查数字和中文。

## 自动高度设为四行，第四行却被裁掉

Gio 的 `LineHeightScale` 为 0 时会使用默认的 1.2 倍行高。若视口按 `行数 × 已测行高` 计算，再仅设置 `LineHeight`，文字排版仍会额外乘 1.2，最后一行便放不下。

`Field.AutoHeight` 使用测量后的行高，并显式设置 `LineHeightScale = 1`，使排版和视口采用同一行距。回归不能只比较输入框是否变高，还要检查四行的 `Editor.Regions` 是否全部落在视口内，覆盖空行和 1× / 2×；对应测试是 `TestTextAreaAutoHeightFitsEveryVisibleLine`。超过最大行数后应滚动，不能截断内容。

## 链接时出现 -lobjc 警告

```
ld: warning: ignoring duplicate libraries: '-lobjc'
```

Gio 和 `native/internal/sys` 都声明了链接 `libobjc`，链接器提示重复并自动去重。不影响结果，忽略即可。

## 界面没刷新，或者 -race 报数据竞争

多半是在 goroutine 里直接改了组件。回调以外的地方改组件，必须包进 `core.Update`：

```go
go func() {
    data := fetch()
    core.Update(func() { label.SetText(data) })   // 不要直接 label.SetText
}()
```

`go test -race` 和 `go run -race` 能发现这类问题。

## 程序卡住，所有窗口都不响应

有回调执行得太久。所有窗口共用一把锁（见[架构 · 线程规则](architecture.md#线程规则)），一个回调卡住，所有窗口都停。常见原因：

- 回调里做了网络请求、读大文件、`screen.Capture`：放进 goroutine。
- 回调里等 `core.Update` 的结果：`core.Update` 要等当前回调结束才会执行，互相等待就是死锁。

## 多窗口时按快捷键或点按钮后卡死

如果是自己新加的窗口方法：它在锁内调用了会等主线程的 Gio 方法（`Perform`、`Option`、`Run`），和另一个窗口形成循环等待。解决办法和原理见[架构 · 不能在锁内等待主线程](architecture.md#不能在锁内等待主线程)。`Raise`、`Close` 已经修复过这个问题。

## 刚打开的窗口马上关闭时崩溃

Gio v0.10.3 在 macOS 上，窗口还没完全创建就被关闭，会在 `cascadeTopLeftFromPoint` 处段错误。Keel 的 `Close`、`Raise` 已经会等到窗口画出第一帧再执行。如果你绕过 Keel，直接调用 Gio 的 `app.Window.Perform`，要自己等第一个 `FrameEvent`。

## 窗口快捷键没反应

- 只有该窗口有焦点时才生效。
- 输入框有焦点时，它会先处理编辑相关的组合键（⌘A、⌘C、⌘V、⌘X、⌘Z 等），同样的组合键注册成窗口快捷键不会触发。换一个组合。
- 快捷键字符串写错会在 `window.Open` 时 panic，不会静默失效。

## 全局快捷键没反应

- macOS 上程序必须在 `window.Main()` 里运行，快捷键事件由主线程的事件循环派发。
- Linux 上需要 X11 会话，Wayland 下只有 XWayland 程序有焦点时才能收到。
- 组合键被其他程序占用时 `hotkey.Register` 返回 `native.ErrConflict`，检查返回值。
- 回调确实触发了但界面没变：回调里改界面要用 `core.Update`。

## 权限一直是 false

- 在终端里 `go run` 时，macOS 通常把授权记在终端程序名下。去 系统设置 → 隐私与安全性 找终端（或 iTerm、VS Code），不是你的程序名。
- 授予屏幕录制权限后要重启程序。
- 每次重新编译，二进制签名都会变，macOS 可能认为是新程序，要重新授权。打包成签名的 `.app` 能避免这个问题。

## 关掉窗口后，main 里 window.Main() 之后的代码没执行

最后一个窗口关闭时进程直接 `os.Exit(0)`，`window.Main()` 不会返回，`defer` 也不会执行。清理工作放在最后一个窗口的 `OnClose` 里。

## 窗口不能隐藏

Gio 不支持隐藏再显示窗口。需要"关掉再打开"的窗口，关闭后重新 `window.Open`，状态保存在你自己的变量里。见[设计决策](decisions.md#选-gio)里列出的代价。

## 组件首帧重叠，交互一次后恢复

先检查尺寸是不是在 Decorate / 绘制阶段才记录，而 Render 已用上一帧的尺寸决定结构。只反复调用 Frame 再断言，容易掩盖这个问题。

本轮有两类修复：Toolbar 在命令区宽度未知时先用“更多”承载操作，测量后再展开，命令区裁剪防止覆盖左右插槽；Dock 用当前根视口初始化停靠尺寸，实际嵌入尺寸不同则请求重绘校正。不要靠先画两帧截图来隐藏错误。

回归应包含未 settle 的第一帧、窗口变窄的第一帧、1× / 2×，并检查被挤占的相邻区域及操作可达性。对应 `TestToolbarFirstFrameKeepsActionsAccessibleBesideSlots`、`TestDockFirstFrameAndResizeKeepCenter`。`examples/components -matrix` 故意只输出首帧，动态浮层和滚动仍通过真实输入路由测试。
