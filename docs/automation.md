# Agent 端到端测试

`cmd/keel-mcp` 是一个 MCP server。接上之后，Agent 可以启动 Keel 应用，读出窗口里有什么，点击、输入、按键、滚动、截图，像人一样把应用走一遍。

两种模式：

- **可见模式**：窗口正常显示在屏幕上。你能看着 Agent 操作，也可以自己上手操作，再交给 Agent 接着做。
- **无界面模式**：窗口不上屏，只在内存里渲染。适合批量跑测试，不打扰你。

两种模式都不动真实的鼠标键盘，也不需要辅助功能等系统权限。一次完整的"输入 → 点击 → 开新窗口 → 保存 → 关闭"流程约 1.5 秒（不含编译）。

## 接入

在 Keel 仓库里：

```sh
go install ./cmd/keel-mcp
claude mcp add keel -- keel-mcp
```

`claude mcp add` 会写入 Claude Code 的配置。其他支持 MCP 的客户端，把 `keel-mcp` 配成 stdio 类型的 server 即可。`keel-mcp` 不需要参数；要测哪个应用，在 `launch` 工具里指定。

## 一次测试长什么样

Agent 调用 `launch`，参数 `{"command": "go run ./examples/multiwindow"}`，得到：

```
Window w1 "Keel 主窗口" 640×420
e1 text "Keel · Gio" @24,24 582×34
e2 text "你的名字" @42,88 546×21
e3 textbox "你的名字" value="" @42,115 546×40
e4 button "打招呼" @42,167 74×38
e5 button "打开设置窗口" @124,167 116×38
e6 text "主窗口和设置窗口拥有各自的状态。" @42,217 546×24
e7 text "⌘ + , 打开设置。" @42,253 546×21
```

每行是一个元素：引用（`e3`）、角色、名字、状态、位置（`@x,y 宽×高`，单位 dp）。然后：

```
type       {"ref": "e3", "text": "小明"}
click      {"text": "打招呼"}          → e6 text "你好，小明！"
press_key  {"key": "mod+,"}           → Open windows: w1 "Keel 主窗口" (shown above); w2 "设置" (new)
click      {"window": "w2", "text": "保存"}
screenshot {"window": "w2"}           → PNG
```

每个操作都返回操作后的元素列表，Agent 不用再单独调 `snapshot` 确认结果。操作打开了新窗口时，列表末尾会标出 `(new)`。

## 操作你自己启动的应用

应用正常显示在屏幕上，你先手动操作到某个状态，再让 Agent 接着做，你看着它操作：

```sh
KEEL_AUTOMATION=1 go run ./examples/multiwindow
```

应用启动后打印连接地址，比如 `keel automation: listening on /var/folders/…/T/keel/multiwindow-4821.sock`。然后让 Agent 调用 `attach`（不带参数）：

- 只有一个应用在跑时，直接连上，返回它当前的窗口和元素；
- 有多个时，列出所有地址，让 Agent 选一个传 `socket` 参数；
- 找不到时，确认应用是用 `KEEL_AUTOMATION=1` 启动的，而且没有退出。

和 `launch` 的区别：

| | `launch` | `attach` |
| --- | --- | --- |
| 谁启动应用 | `keel-mcp` | 你 |
| `stop` | 结束应用 | 只断开连接，应用继续运行 |
| 再次连接 | 重新启动，状态清空 | 再 `attach` 一次，状态保留 |
| `logs` | 能看到输出 | 看不到，输出在你启动它的终端里 |

一个应用同一时间只接受一个连接。另一个 Agent 会话已经连着时，`attach` 会在 3 秒后报"另一个客户端已连接"，不会一直卡住。

Agent 的每次操作都会立刻反映在屏幕上的窗口里：输入的文字、点击后的结果、新打开的窗口。你在窗口里的操作，Agent 下一次读取时也能看到。

不想显示窗口时，加 `KEEL_HEADLESS=1`。

`KEEL_AUTOMATION` 也可以直接写 socket 路径（`KEEL_AUTOMATION=/tmp/my.sock`），这时 `attach` 要传同样的 `socket`。

## 工具

| 工具 | 参数 | 作用 |
| --- | --- | --- |
| `launch` | `command`，`dir?`，`timeout_seconds?`，`visible?` | 用 shell 启动应用，等它就绪，返回第一个窗口。默认不显示窗口；`visible: true` 显示在屏幕上。会先停掉之前启动的应用。编译时间算在超时里，默认 120 秒 |
| `attach` | `socket?` | 连接你用 `KEEL_AUTOMATION=1` 启动的应用，保留它当前的状态 |
| `snapshot` | `window?` | 窗口里的元素 |
| `click` | `ref` 或 `text` 或 `x`+`y` | 左键点击元素中心或指定坐标 |
| `type` | `text`，`ref?`，`clear?` | 在光标处输入。给 `ref` 先点击聚焦；`clear` 先全选，用新文字替换 |
| `press_key` | `key` | 按一个组合键：`enter` `esc` `tab` `shift+tab` `space` `backspace` `up` `mod+a` `ctrl+shift+s`。窗口快捷键会触发，Tab 会移动焦点 |
| `scroll` | `dy`，`ref?` 或 `x`+`y` | 滚轮滚动，正数向下，单位 dp。默认在窗口中心 |
| `wait_for` | `text`，`timeout_ms?` | 等到某个元素的名字或值包含 `text`。用于后台 goroutine 更新界面的场景，默认 5 秒 |
| `screenshot` | `window?` | 窗口截图，1 像素 = 1 dp |
| `windows` | | 列出打开的窗口 |
| `close_window` | `window?` | 像点关闭按钮一样关窗口。关掉最后一个，应用退出 |
| `logs` | `lines?` | 应用最近的标准输出和错误输出：panic、日志、编译错误。只对 `launch` 启动的应用有效 |
| `stop` | | `launch` 的应用：结束它；`attach` 的应用：只断开 |

除 `launch`、`attach`、`logs`、`stop` 外都可以带 `window`（如 `w2`），省略时作用于当前窗口：最近一次被操作的窗口；还没操作过时，是最近打开的窗口。

### 元素

| 角色 | 来自 | 额外信息 |
| --- | --- | --- |
| `text` | `widget.Text`、`Heading`、`Muted` | |
| `button` | `widget.Button` | `disabled` |
| `link` | `widget.Link` | |
| `textbox` | `widget.Input`、`TextArea` | `value`；密码框的值是等长的 `•` |
| `checkbox` | `widget.Checkbox` | `checked` / `unchecked`；半选时 `value` 为 mixed |
| `radio` | `widget.RadioGroup` 的每个选项 | `checked` / `unchecked` |
| `switch` | `widget.Switch` | `checked` / `unchecked` |
| `select` | `widget.Select` | `value` 是当前选中项；点击后出现 `option` |
| `option` | 展开的下拉选项 | `selected` |
| `tab` | `widget.Tabs` 的标签 | `selected` |
| `table` | `widget.Table` | `value` 是总行数，如 `36 行` |
| `columnheader` | 表头，点击排序 | |
| `row` | 表格中可见的行，名字是各列用竖线连起来 | `selected` |
| `slider` | `widget.Slider` | `value` 是当前数值；点击轨道或聚焦后按方向键调整 |
| `accordion` | `widget.Accordion` | 标题和展开内容单独列出 |
| `tag` | kit 标签（不可点击） | 名字是文字，value 为语义级别 |
| `group` | kit 组件分组 | 名字为分组标题，保留子组件语义 |
| `status` | kit 状态栏 | 名字为主状态，value 为详情 |
| `alert` | kit 行内提示（不可点击） | 名字是标题，value 为 neutral/info/success/warning/danger；正文单独列出 |
| `badge` | 数字、圆点、图标角标（不可点击） | 名字为原始计数，`value` 为显示值、`dot` 或 `icon` |
| `toggle` | 状态按钮 | `selected` 表示选中，支持 `disabled` |
| `disclosure` | 折叠面板标题 | `value` 是 expanded / collapsed，支持 `disabled` |
| `image` | `widget.Image`、Markdown 图片、kit Avatar | `value` 是 loading / loaded / error；Avatar 回退为 initials。名字是替代文字 |
| `footnotes` | Markdown 脚注 | 引用和返回链接单独列出 |
| `progressbar` | `widget.Progress` | `value` 是百分比或 indeterminate |
| `dialog` | 打开的 `widget.Dialog` | 它里面的文字和按钮单独列出 |
| `link` | Markdown 段落里的链接、`widget.Link` | `value` 是网址 |
| `code` | Markdown 代码块 | 名字是语言；里面的代码文字和"复制"按钮单独列出 |

元素列表只包含看得见的部分：滚出视野的表格行、页面内容不会列出，部分可见的元素按可见部分报告位置。要看更多行，先 `scroll`。

`ref` 只在下一次操作之前有效，因为每次操作都会重新编号。按文字定位（`text`）更稳：完全匹配优先于部分匹配，控件优先于普通文字；同分时取后画的那个，所以对话框、下拉框里的按钮优先于被它们盖住的同名元素。

坐标只在没有更好的办法时使用。截图和坐标用同一套单位：截图上的一个像素就是一个 dp。

## 原理

```
Agent ──MCP(stdio)──► keel-mcp ──JSON 行(unix socket)──► 应用进程
                         │                                 │
               launch：启动应用并设置             ui/window 自动化模式：
               KEEL_AUTOMATION=<socket>          窗口不上屏，按请求渲染
               attach：连接你启动的应用           一个连接断开后等下一个
```

应用进程里：

- `ui/window` 启动时读到 `KEEL_AUTOMATION` 环境变量，就给每个窗口配一个**影子窗口**：同一套组件、同样的尺寸，但有自己的 Gio 输入路由。Agent 的操作都送到影子窗口。
- 可见模式下，真实窗口照常显示，你的鼠标键盘走真实窗口原来的路径，完全不受影响。两边共享同一批组件对象，所以 Agent 通过影子窗口点了按钮，回调执行、状态改变，真实窗口立刻重绘；你在真实窗口里输入的内容，影子窗口下一次渲染时也能看到。影子窗口的尺寸跟随真实窗口，你拖动窗口改变大小，Agent 读到的坐标也跟着变。
- 无界面模式（`KEEL_HEADLESS=1`）下没有真实窗口，`window.Main` 不进入系统事件循环，只在 socket 上处理请求。
- 组件、回调、窗口快捷键、`core.Update`、帧锁，走的都是和真实窗口完全相同的代码。
- "页面有什么"来自 Gio 每帧生成的语义树：每个组件声明自己的角色、名字、状态，路由器算出它在窗口里的绝对位置。Keel 的组件在 `ui/widget/semantics.go` 的 `area` 里声明这些信息。
- 点击、输入、滚动被转换成 Gio 的指针和键盘事件，送进这个窗口的路由器，然后渲染到画面稳定为止：回调改了状态要再画一帧才能看到，最多画 10 帧。

应用侧协议写在 `ui/window/automation_server.go` 的文件注释里。`keel-mcp` 不引用任何 Keel 包，只说这个协议；想用别的语言写测试客户端，照着协议发 JSON 即可。

## 局限

- **焦点各管各的。** Agent 点进一个输入框后，焦点在影子窗口里，屏幕上的输入框不显示光标和蓝色边框，但输入的文字会显示出来。反过来，你在屏幕上点进输入框，Agent 调 `type` 时不带 `ref` 会报"没有焦点"，要带上 `ref`。
- **你和 Agent 同时操作时，按先后顺序执行**，不会冲突。但你按住鼠标拖动的过程中 Agent 插进来点击，拖动状态可能会乱。
- **关闭窗口会等它真正关掉再返回。** 可见模式下 `close_window` 关的是屏幕上的真实窗口，窗口销毁后才回复，所以返回的窗口列表就是屏幕上的样子。
- **Agent 操作的不是真实窗口本身。** 系统窗口层面的问题测不到：窗口位置和尺寸、系统菜单、输入法候选框，以及真实窗口之间与主线程相关的死锁。后者用 `KEEL_DESKTOP=1` 的真实窗口测试覆盖，见[测试](testing.md#真实窗口测试)。
- **`native/*` 不在范围内。** 权限、截屏、全局快捷键调的是系统 API，自动化模式不会模拟它们。
- **时间只在请求时前进。** 应用只在收到请求时渲染。后台 goroutine 的 `core.Update` 在下一次请求时生效；要等它，用 `wait_for`。
- **不支持悬停、拖拽、右键、双击。** 现有组件用不到，需要时在 `ui/window/automation.go` 里加。
- **不报告焦点位置。** `type` 不带 `ref` 时输入到当前有焦点的输入框；没有焦点会报错。
- **自己用 `core.Func` 写的布局对 Agent 不可见**，除非在里面加 Gio 的 `semantic` 操作。新组件的做法见[扩展指南](extending.md#新增组件)。

## 在 Go 测试里用

`cmd/keel-mcp/main_test.go` 用 MCP 官方 SDK 的客户端启动 `keel-mcp`，把 multiwindow 示例完整走一遍（`TestEndToEnd`），并测试 `attach` 的连接、断开、重连和占用提示（`TestAttach`），也是写这类测试的样板。`ui/window/automation_test.go` 在进程内直接测试自动化模式（滚动、Tab 焦点、回调里关窗口）。两者都随 `go test ./...` 运行，不弹窗口。

| `avatar` | kit 头像 | 名字为人名，value 为 online/busy/offline；无状态为空 |
