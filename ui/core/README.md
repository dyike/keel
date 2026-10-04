# ui/core

所有界面模块的地基：

| 名字 | 作用 |
| --- | --- |
| `Widget` | 接口：能 `Layout` 的东西。所有组件、容器都实现它 |
| `Func` | 把一段 Gio 布局函数当组件用 |
| `DecodeImage(ctx, source)` | 读取本地/HTTP/data 图片，限制编码大小与像素数；须在后台调用并管理超时。`ReadImageSource` / `DecodeImageBytes` 是它拆开的两步，kit 的 Image 用来另外解码 SVG、GIF 动图 |
| `Update(fn)` | 从任意 goroutine 修改界面：`fn` 在下一帧执行 |
| `Call(gtx, fn)` | 给写组件的人用：执行用户回调并让所有窗口重绘 |
| `Semantic(gtx, w, ops...)`、`Role(...)` | 给写组件的人用：声明组件的角色、名字、状态，让 Agent 看得见 |
| `Bind`、`BindIn`、`Bindings`、`BindingsIn` | 键位表：动作名到快捷键，`BindIn` 的上下文可以写条件表达式（`Editor && !ReadOnly`、`Pane > Editor`），见 [菜单 · 上下文条件表达式](../../docs/kit/menu.md#上下文条件表达式) |
| `CurrentScrollGesture`、`ReportScrollGesture` | 滚动的设备（滚轮/触控板）和手势阶段；`ui/window` 从 macOS、Wayland 读来报告，Carousel 等组件读取 |
| `WindowControls`、`CurrentWindow()` | 组件拿到所在窗口的激活状态、标题区域登记、最小化、最大化、关闭能力，自定义标题栏用。`ui/window` 在布局期间登记当前窗口 |

- **依赖**：只依赖 Gio，以及内部的 `ui/internal/loop`。
- **被谁依赖**：`el`、`kit`、`window`、`markdown`。

线程规则：回调里直接改组件；其他 goroutine 改组件包进 `core.Update`。原因见 [架构 · 线程规则](../../docs/architecture.md#线程规则)。

`ClipboardData`、`ClipboardImage`、`ClipboardReader` 定义富粘贴交换数据与异步读取接口；core 不读取系统剪贴板，不引用 native。应用适配平台读取结果，由输入组件调度回 UI 线程。
