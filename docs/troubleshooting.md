# 常见问题

## 中文显示成方框

Gio 自带的 Go 字体没有中文，中文靠系统字体回退。自动回退有时会选到缺字的字体，比如粗体标题里的"组"显示成方框。

`theme.Face` 固定了字体优先级：苹方 → 冬青黑体 → 微软雅黑 → Noto Sans CJK。在 Linux 上显示方框，说明这几种都没装，装上 Noto CJK 字体即可（Debian/Ubuntu：`fonts-noto-cjk`）。要用别的字体，改 `theme.Face`，把字体族名放在最前面。

## 按钮里的中文偏上

中文系统字体的上沿（ascent）很高，同样的内边距下文字看起来偏上 3–4dp。`theme.CJKNudge`（默认 2dp）把按钮、输入框里的文字往下挪。这个值是按 macOS 苹方调的，换平台或字体后如果不居中，调它。

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

- 程序必须在 `window.Main()` 里运行，快捷键事件由主线程的事件循环派发。
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
