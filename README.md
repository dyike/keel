# Keel

Keel 用 Go 组合桌面 UI，用 [Go-Gui](https://github.com/go-gui-org/go-gui) 创建和绘制窗口，并提供权限、截图、输入和全局快捷键等原生扩展。应用代码只写 Go；桌面运行时不使用 Wails、WebView、HTML、CSS 或 JavaScript。

当前固定依赖 Go-Gui `v0.82.0`，要求 Go 1.26.1+。macOS 使用 Metal 和 Objective-C/cgo，需要 Xcode Command Line Tools；Keel 的原生扩展要求 macOS 14+。Windows 使用 Win32/OpenGL，Linux 使用 X11/EGL；两者可关闭 cgo 构建，运行时需要对应的桌面和图形环境。

## 先运行多窗口示例

在项目根目录执行：

```sh
go run ./examples/multiwindow
```

启动后显示主窗口。按 `⌘ + ,` 或点击“打开设置窗口”，会在运行时创建设置窗口。主窗口输入名字后点击“打招呼”，文本由 Go 回调更新。设置窗口有独立的输入和复选框状态；关闭后再次按快捷键，会重新创建。已有窗口会被显示并置于前台，连续按键不会创建重复窗口。所有窗口都关闭后退出应用。

单窗口组件示例：`go run ./examples/ui`；加 `-screenshot work/ui.png` 可导出软件渲染预览。最小示例：`go run ./examples/minimal`。完整代码见 [多窗口示例](examples/multiwindow/main.go) 和 [UI 文档](ui/README.md)。

## 最小应用

```go
package main

import (
    "log"

    "github.com/dyike/keel"
    "github.com/dyike/keel/ui"
)

func main() {
    kit := keel.New()
    name := ui.Input(ui.InputOptions{Label: "你的名字"})
    result := ui.Text("等待输入")
    page := ui.NewPage("Hello", ui.Column(
        name,
        ui.Button("打招呼", func() {
            result.SetText("你好，" + name.Value() + "！")
        }),
        result,
    ))
    _, err := kit.Window.New(keel.WindowOptions{
        Title: "Hello Keel", Width: 640, Height: 420, UI: page,
    })
    if err != nil { log.Fatal(err) }
    if err := kit.Run(); err != nil { log.Fatal(err) }
}
```

页面属于窗口，放在 `WindowOptions.UI`。需要第二个窗口，就再调用一次 `kit.Window.New(...)`，传入新的 Page 和组件。按钮回调里也可以调用同一个方法动态开窗。重复使用同一个 Page 创建窗口返回 `ErrConflict`；关闭的窗口和 Page 不能重新打开，应创建新的对象。只想暂时收起窗口，用 `Hide()`，之后 `Show()`。

Keel 尚未发布版本。在消费项目中添加本地替换，保存应用代码后运行 `go mod tidy`：

```sh
go mod edit -require=github.com/dyike/keel@v0.0.0
go mod edit -replace=github.com/dyike/keel=/Users/bytedance/Code/go/src/github.com/dyike/keel
go mod tidy
```

## 窗口和线程

`kit.Run()` 从 `main` 调用一次，启动前至少创建一个窗口。默认所有窗口关闭后退出；隐藏的工具窗口仍然存活。使用 `Options{ExitOnMainClose: true}` 可改成第一个窗口关闭时退出。已有 Go-Gui 应用可通过 `keel.Attach(app)` 接入，再用 Kit 的窗口管理和 `Run()` 管理生命周期。

macOS 的 Dock“退出”会进入窗口关闭及资源清理流程；“隐藏”或 `⌘H` 隐藏当前可见窗口，点击 Dock 图标恢复这些窗口。原本隐藏的工具窗口保持隐藏。终端直接启动时，如果系统应用隐藏接口拒绝操作，Keel 会在原生窗口层完成隐藏与恢复。Keel 必须先于其他原生 GUI 库初始化 `NSApplication`；如果其他库已经创建应用实例，`Run()` 会返回 `ErrConflict`。

运行时创建窗口是异步的：`Window.New` 返回的是窗口对象，原生创建完成后 `OnReady` 才会执行。动态窗口尚未就绪时 `Host()` 为 nil。Go-Gui 原生创建失败目前由上游记录日志，Keel 尚未提供每个异步请求的失败回调；同一管理器最多允许 16 个待创建请求，超过时返回 `ErrNotReady`。应用应统一通过 `kit.Window.New` 创建窗口，避免与 `kit.App.OpenWindow` 的队列混用。

`Show`、`Hide`、`Toggle` 和毛玻璃设置在内部排队到 UI 线程；返回 nil 表示请求已入队。按钮和输入回调运行在 UI 线程，耗时工作放到 goroutine，更新组件后调用 `page.Refresh()`。`window.Host()` 返回原始 `*gui.Window`；后台任务使用它的 `QueueCommand`，不要直接调用原生设置方法。

`WindowOptions.GUI` 接受高级 `gui.WindowCfg` 配置。`WindowOptions.View` 可直接使用 Go-Gui 的控件，不能与 `UI` 同时设置：

```go
_, err := kit.Window.New(keel.WindowOptions{
    Title: "Go-Gui view",
    View: func(w *gui.Window) gui.View {
        return gui.Column(gui.ContainerCfg{
            Sizing: gui.FillFill,
            Content: []gui.View{gui.Label("自定义控件", gui.TextStyle{})},
        })
    },
})
```

## Cmd+, 打开设置

`window.RegisterShortcut` 注册应用内快捷键，回调在 UI 线程执行。它在输入框的按键处理之前触发，所以编辑文本时也能打开设置。绑定在窗口上：主窗口和每个动态创建的设置窗口都绑定同一个操作，代码见 [多窗口示例](examples/multiwindow/main.go)。

```go
// openSettings 创建新的设置窗口，或显示已经存在的窗口。
cancel, err := mainWindow.RegisterShortcut("super+comma", openSettings)
if err != nil { return err }
// 需要提前取消时调用 cancel()；窗口关闭后绑定不再接收事件。
_ = cancel
```

`super+comma` 和 `cmd+,` 都表示 macOS 的 `⌘ + ,`。启动窗口可以在 `Run` 前绑定；动态窗口在 `OnReady` 内绑定。此 API 也能在 Windows/Linux 编译使用，Super 对应平台的系统键。

应用内快捷键与 `kit.Shortcut.Register` 的系统全局热键是两种作用域。全局热键回调在后台运行，操作 UI 时使用 `kit.Dispatch(func() { ... })` 切回 UI 线程。`⌘ + ,` 示例使用应用内绑定，不占用其他应用的设置快捷键。

## 原生能力与限制

| 能力 | macOS | Windows / Linux |
| --- | --- | --- |
| Go UI、多个窗口、动态开窗、显示/隐藏 | Go-Gui / Metal | Go-Gui / Win32、X11、OpenGL |
| 透明窗口、无边框窗口 | Go-Gui | Go-Gui，Linux 透明需合成器 |
| 毛玻璃 | `SetVibrancy`，配合透明窗口 | `ErrUnsupported` |
| Go-Gui 窗口置顶、鼠标穿透 | 当前适配器返回 `ErrUnsupported` | `ErrUnsupported` |
| 全局快捷键 | Carbon | `ErrUnsupported` |
| 权限、显示器、截图、输入扩展 | ApplicationServices、CoreGraphics、ScreenCaptureKit | `ErrUnsupported` |

Go-Gui v0.82.0 没有公开原生窗口句柄、置顶和鼠标穿透接口，因此 Keel 不会把这些调用伪装成成功。`window.NativeDriver` 仍保留给能提供原生句柄和同步 UI 调度的其他宿主。权限和输入扩展使用 `internal/driver.Backend`；平台实现通过构建标签选择，业务层无需 `runtime.GOOS` 分支。

| 需求 | 调用 |
| --- | --- |
| 创建窗口 | `kit.Window.New(keel.WindowOptions{UI: page})` |
| 显示、隐藏、切换、销毁 | `win.Show()` / `win.Hide()` / `win.Toggle()` / `win.Close()` |
| 隐藏的工具窗口 | `kit.NewOverlay(keel.OverlayOptions{UI: page})` |
| 应用内快捷键 | `win.RegisterShortcut("super+comma", handler)` |
| 全局快捷键 | `kit.Shortcut.Register("super+shift+k", handler)` |
| 显示器列表、PNG 截图 | `kit.Screen.Displays()` / `kit.Screen.Capture(id)` |
| 检查、申请权限 | `kit.Permissions.Check(kind)` / `kit.Permissions.Request(kind)` |
| 鼠标位置、移动、点击 | `kit.Input.Position()` / `Move(point)` / `Click(button)` |
| 物理按键 | `kit.Input.KeyDown(key)` / `KeyUp(key)` |
| 启动、请求退出 | `kit.Run()` / `kit.Quit()` |

系统全局快捷键使用 Keel 的低层原生实现，在窗口的 `OnReady` 内注册。回调在后台串行运行，忙时重复事件合并；`Super` 在 macOS 表示 Command。应用退出时释放注册，也可以调用注册对象的 `Close()`。`kit.Quit()` 将清理和关闭请求排队到 UI 线程；`kit.Close()` 释放组件及快捷键资源，不代替正在运行时的退出请求。

`Check` 只检查；只有显式 `Request` 才可能弹系统提示。截图先检查 `permissions.ScreenRecording`，从后台调用 `Capture`；它不申请权限，等待最多 10 秒，返回整块显示器的 PNG。坐标是全局逻辑点，多屏坐标可以为负数。鼠标移动、点击和键盘发送需要辅助功能权限；`KeyDown` 必须配对 `KeyUp`。本版不实现输入监听、Unicode 文本发送和剪贴板。

## 结构与验证

```text
keel.go                    Go-Gui 应用、事件循环和清理
window/                    多窗口管理、Driver、Go-Gui 适配
ui/                        Go 组件、每窗口 Page、Go-Gui 控件渲染
shortcut/                  全局快捷键生命周期
permissions/ screen/ input/ 原生扩展 public API
internal/driver/            统一原生平台接口
internal/platform/          darwin / windows / linux 实现
examples/                  minimal、ui、multiwindow、overlay、smoke、inspect
```

```sh
go test -race ./...
go vet ./...
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ./...
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./...
go run ./examples/smoke
# macOS：打开临时窗口，验证隐藏、Dock Apple Event 恢复/退出与清理
KEEL_TEST_DESKTOP=1 go test -count=1 ./internal/desktop -run TestNativeApplicationLifecycle -v
```

`smoke` 在 macOS 上创建两个原生窗口，通过 Go-Gui 控件事件验证输入→Go 回调，确认状态隔离，再通过 `Cmd+,` 键盘事件分发动态创建设置窗口后退出。它不发送系统鼠标/键盘事件，不申请权限。UI 测试使用软件渲染与真实控件事件分发，不启动桌面窗口。Windows/Linux 的交叉构建通过只说明能编译，桌面运行尚未实测；macOS 关闭 cgo 时能做无界面测试，启动桌面后端会返回 `ErrUnsupported`。

本机工具链可能给 cgo 对象使用较新的 macOS 部署目标，可统一设置：

```sh
export CGO_CFLAGS='-O2 -g -mmacosx-version-min=14.0'
export CGO_LDFLAGS='-O2 -g -mmacosx-version-min=14.0'
go run -ldflags='-extldflags=-mmacosx-version-min=14.0' ./examples/multiwindow
```

当前本机验证不等于已在 macOS 14 或 Windows/Linux 桌面验证。

链接库不再强制最终可执行文件的部署目标，因此普通 `go run` 使用工具链默认值。Go-Gui 的原生依赖可能出现 `ignoring duplicate libraries: -lobjc` 警告；它不影响链接和运行。
