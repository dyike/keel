# 原生能力

`native/` 下的各个包提供 Gio 没有的系统能力。它们不依赖界面模块，可以单独使用。

| 包 | 能力 | 需要的权限 |
| --- | --- | --- |
| `native/permission` | 检查、申请权限 | 无 |
| `native/screen` | 列出显示器；截图 | 截图需要屏幕录制 |
| `native/input` | 移动鼠标、点击、按键 | 读取鼠标位置以外的操作都需要辅助功能 |
| `native/hotkey` | 全局快捷键 | 无 |
| `native/clipboard` | 异步读取文本、编码图片、文件路径（macOS / Windows / Linux X11） | 无 |

支持三个平台，其他平台以及关闭 cgo 构建的 macOS 上，所有函数返回 `native.ErrUnsupported`，程序照常编译。

| 平台 | 实现 | 说明 |
| --- | --- | --- |
| macOS 14+ | cgo 调用系统框架 | 需要用户授权，见下文 |
| Windows 10+ | 直接调用 user32、gdi32，不需要 cgo | 不需要授权；坐标按主显示器的 DPI 换算成逻辑点 |
| Linux | 纯 Go 实现 X11 协议，不需要 cgo | 需要 X11 会话；合成输入需要 X 服务器的 XTEST 扩展；逻辑点按 `Xft.dpi` 换算 |

Linux 的 Wayland 会话不允许普通程序截取整个屏幕或向其他程序注入输入。有 XWayland（`DISPLAY` 已设置）时可以调用，但只能看到、操作 X11 程序的窗口；没有 `DISPLAY` 时返回 `ErrUnsupported`。

## 错误

所有错误都包装自 `native` 包里的哨兵值，用 `errors.Is` 判断：

| 错误 | 含义 |
| --- | --- |
| `ErrUnsupported` | 当前平台没有实现 |
| `ErrPermissionDenied` | 缺少所需权限 |
| `ErrInvalidArgument` | 参数不合法，比如未知按键名、显示器 ID 为 0 |
| `ErrTimeout` | 系统调用超时（截图超过 10 秒） |
| `ErrConflict` | 全局快捷键已被本进程或其他程序占用 |
| `ErrFailed` | 其他系统错误，错误信息里带状态码 |

```go
if _, err := screen.Capture(id); errors.Is(err, native.ErrPermissionDenied) {
    permission.Request(permission.ScreenRecording)
}
```

## permission：权限

```go
ok, err := permission.Granted(permission.Accessibility) // 只查，不弹窗
ok, err := permission.Request(permission.Accessibility) // 可能弹出系统授权框
```

| 常量 | 系统设置里的名字 | 谁需要 |
| --- | --- | --- |
| `Accessibility` | 辅助功能 | `input` 包 |
| `ScreenRecording` | 屏幕录制（macOS 15 起叫"屏幕与系统录音"） | `screen.Capture` |
| `InputMonitoring` | 输入监控 | 目前没有功能用到，预留 |

Windows 和 Linux 没有这几种授权，`Granted` 和 `Request` 总是返回 `true`。

在 macOS 上使用时要注意三点：

- **`Request` 不等用户回答。** 它弹框后立刻返回当时的授权状态，通常是 `false`。用户在系统设置里打开开关后，你需要再调 `Granted` 确认。
- **`false` 不区分"拒绝过"和"还没问过"。** macOS 不提供这个信息。
- **授权记在哪个程序名下。** 打包成 `.app` 运行时，授权记在这个 app 名下；在终端里 `go run`，macOS 通常把授权记在终端程序（终端、iTerm、VS Code）名下。屏幕录制权限授予后，一般要重启程序才生效。

## screen：显示器与截图

```go
displays, err := screen.Displays()
for _, d := range displays {
    fmt.Println(d.ID, d.X, d.Y, d.Width, d.Height, d.PixelWidth, d.PixelHeight, d.Primary)
}
```

`X`、`Y`、`Width`、`Height` 是逻辑点坐标，原点在主显示器左上角，副屏可能是负坐标。`input` 包用同一套坐标。

```go
png, err := screen.Capture(d.ID)
```

- 返回 PNG 字节，尺寸为 `PixelWidth × PixelHeight`，不含鼠标指针。
- macOS 上需要屏幕录制权限，自己不会弹框，没有权限直接返回 `ErrPermissionDenied`。
- 最多阻塞 10 秒。**不要在回调里调用**，它会让所有窗口卡住。放进 goroutine，结果用 `core.Update` 送回界面。
- macOS 上不能在主线程调用，否则返回 `ErrFailed`。`main` 函数在 `window.Main()` 之前运行在主线程上，也不能在那里调用。

## input：合成键鼠

```go
x, y, err := input.MousePosition()   // 不需要权限
input.MouseMove(100, 200)            // 以下都需要辅助功能权限
input.Click(input.Left)              // 在当前鼠标位置点击；Left / Right / Middle
input.Tap("enter")                   // 按下并松开
input.KeyDown("cmd"); input.Tap("c"); input.KeyUp("cmd")   // ⌘C
```

按键名按美式键盘的物理位置，与当前输入法和键盘布局无关（Linux 例外：X11 按当前布局查找能打出这个字符的键）：

- 字母、数字、符号：`a`–`z`、`0`–`9`、`-` `=` `[` `]` `\` `;` `'` `,` `.` `/` `` ` ``
- 功能键：`enter` `tab` `space` `backspace` `delete` `escape` `home` `end` `pageup` `pagedown` `up` `down` `left` `right` `f1`–`f12`
- 修饰键：`cmd` `shift` `alt` `ctrl`。Windows 和 Linux 上 `cmd` 是 Windows 键（Super）。

`KeyDown` 和 `KeyUp` 必须成对调用，否则系统会认为这个键一直按着。

## hotkey：全局快捷键

其他应用在前台时也能触发。窗口有焦点时才生效的快捷键用 [`window.Options.Shortcuts`](app.md#窗口快捷键)。

```go
unregister, err := hotkey.Register("cmd+shift+k", func() {
    core.Update(func() { status.SetText("触发了") })
})
defer unregister()
```

- 写法是 `修饰键+按键`，至少要一个修饰键。修饰键：`cmd` `ctrl` `alt`（或 `option`）`shift`；按键名同 `input` 包。
- 回调在独立的 goroutine 里执行，不持有界面锁，**改界面必须包进 `core.Update`**。
- 回调还在执行时又按了几次，只会再触发一次，不会排队。
- 组合键已被占用时返回 `ErrConflict`。
- `unregister` 可以重复调用，只有第一次生效。

`cmd` 在 Windows 和 Linux 上是 Windows 键（Super）。跨平台的快捷键通常写成 macOS 用 `cmd`、其他平台用 `ctrl`。

macOS 上需要 `window.Main()` 在运行：快捷键事件由主线程的事件循环派发。Windows 和 Linux 有自己的消息线程，不受这个限制。macOS 上不开窗口的纯后台程序暂时用不了，见[常见问题](troubleshooting.md#全局快捷键没反应)。

## notification：系统通知

`native/notification` 提供 Available、RequestPermission、Post 和 Remove，完成回调在独立 goroutine 执行。当前实现 macOS .app 的授权、按 ID 投递/替换和撤回，Linux 桌面 D-Bus 后端，以及 Windows 托盘气泡后端（同一时间显示一条，支持点击回调）；其他未支持平台明确返回不支持。kit.Notifier 通过应用适配器接入，macOS 已实现原生前台展示与 Message.OnClick，kit 可通过交互后端接收系统点击并请求 Window.Raise，Linux 已支持声明 actions 的服务的默认点击，Wayland ActivationToken 仍未完成，完整用法与验收步骤见 [模块文档](../native/notification/README.md)。

## clipboard：富剪贴板快照

`clipboard.Read(func(data clipboard.Data, err error))` 异步读取；完成回调在后台 goroutine 执行，界面更新需转回 UI 帧。`Text` 为文本，`Images` 为 MIME 和编码字节，`Files` 为路径引用，不打开文件。

macOS 使用 AppKit，图片优先 PNG、其次 TIFF。Windows 使用 Win32，读取 Unicode 文本、PNG、CF_DIB 和 CF_HDROP；DIB 添加 BMP 文件头后以 `image/bmp` 返回，不解码像素。文件引用优先于资源管理器附带的图片预览。文本、路径 UTF-8 字节与编码图片合计最多 16MiB，文件最多 128 个；错误不返回部分结果。Windows 读取期间保持剪贴板打开，并在关闭前复制所有数据；剪贴板占用时最多尝试 8 次，间隔 15ms。

DIB 支持 40/108/124 字节头的 RGB、调色板和位域布局；压缩位图及带独立颜色配置文件的 V5 布局暂不支持。HTML/RTF 尚未作为独立格式返回。iOS 及无 cgo 的 macOS 返回 `ErrUnsupported`。

组件库的 Input/Textarea 和 CodeEditor 示例已共用这个读取适配；失败时沿原有 Gio 文本通路粘贴。Windows 已通过格式解析、像素解码、错误/上限测试和交叉编译，系统剪贴板及真实窗口粘贴尚未在 Windows 真机验收。

Linux X11 使用独立连接读取 CLIPBOARD，支持 UTF-8/Latin-1 文本、PNG/JPEG/TIFF/BMP/WebP 编码图片、URI 列表及 GNOME 文件引用。只返回本机绝对路径，忽略远程文件主机和非文件 URL，不执行复制或剪切。遵守相同的 16MiB/128 文件限制；处理 INCR 分块传输，整个连接与读取限时 5 秒。读取前后检查所有者及其提供的 TIMESTAMP；不提供时间戳的旧应用若在同一所有者内部更改内容，无法保证跨格式原子快照。

连接使用 DISPLAY 和 XAUTHORITY（默认 ~/.Xauthority），支持 MIT-MAGIC-COOKIE-1。纯 Wayland 尚无读取后端；有 DISPLAY 时通过 X11/XWayland 请求。格式、分块协议替身、错误上限测试与交叉编译通过，Linux 桌面真实剪贴板和 XWayland 桥接仍待验收。
