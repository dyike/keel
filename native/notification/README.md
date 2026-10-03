# notification

独立的系统通知接口，当前实现 macOS 14+、cgo、带 bundle identifier 的 `.app`，以及 Linux 桌面 D-Bus 通知服务。Windows、WebAssembly、iOS/Android 和 macOS 无 cgo 构建返回 `native.ErrUnsupported`；macOS 普通 `go run` 同样不支持。Linux 不要求 X11，Wayland 下也通过会话总线通信。

```go
notification.RequestPermission(func(err error) {
    if err != nil { return }
    notification.Post(notification.Message{
        ID: "download/report", Title: "下载完成", Body: "report.pdf 已保存。",
    }, func(err error) {
        // err == nil 表示系统接受请求，不代表横幅已显示。
    })
})
// 在 Post 完成后按同一 ID 撤回：
notification.Remove("download/report", func(err error) {})
```

- `Available()` 只检查平台实现和应用包身份，不检查授权、签名信任或系统展示策略。
- `RequestPermission` 申请 alert 权限；`Post` 不自动弹权限提示，未授权返回 `ErrPermissionDenied`。
- ID 在整个应用内共享，非空且不能包含 NUL；正文或标题至少一个非空，所有字符串必须是合法 UTF-8。相同 ID 重新投递由系统替换已有请求。
- 完成回调在独立 goroutine 执行，可传 nil。UI 修改应放进 `core.Update`；不要等待异步完成时阻塞主线程。
- 同一 ID 的操作应等待前一次完成后再执行，避免异步投递和撤回交错。Remove 同时撤回待投递与已送达项；系统没有撤回完成确认，回调成功仅表示已发出撤回调用。
- macOS 必须运行应用事件循环；权限对话框、专注模式、系统设置和签名信任影响实际显示。macOS 已注册通知中心 delegate，为本模块通知请求前台 Banner/List；系统设置仍可禁止展示。未实现指定应用窗口激活和跨启动回调恢复。
- kit.Notifier 通过 NoticeSystemBackend 接入；`examples/notification` 提供适配器，处理仅系统/应用内加系统、同 ID 更新和撤回。macOS 示例通过 NoticeSystemInteractiveBackend 接入 kit 点击、关闭和 Window.Raise；Linux 示例也使用交互后端，要求通知服务声明 actions。

可手动运行 `examples/notification`。先构建应用包，再打开它：

```sh
mkdir -p /tmp/KeelNotification.app/Contents/MacOS
go build -o /tmp/KeelNotification.app/Contents/MacOS/KeelNotification ./examples/notification
cat > /tmp/KeelNotification.app/Contents/Info.plist <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>com.keel.notification-demo</string>
<key>CFBundleName</key><string>KeelNotification</string>
<key>CFBundleExecutable</key><string>KeelNotification</string>
<key>CFBundlePackageType</key><string>APPL</string>
</dict></plist>
PLIST
codesign --force --deep --sign - /tmp/KeelNotification.app
open /tmp/KeelNotification.app
```

此临时应用包用于开发验证；正式分发需要正常签名和可信安装位置。点击“申请通知权限”，允许后投递，切到其他应用观察通知中心；再次投递验证替换，再点“撤回任务通知”验证消失。Notifier 串行化系统请求，并显示真实错误；“应用内和系统通知”模式可验证应用内超时不撤回系统通知。`-check` 只打印平台/包身份检查结果，不申请权限、不发送通知。

自动测试覆盖参数校验和未打包进程的拒绝路径，不弹出系统权限窗口。系统授权、真实通知展示/替换/撤回仍需上述手动验收。

接口依据：[Apple UNUserNotificationCenter](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter)。


Linux 使用运行中的 `org.freedesktop.Notifications` 服务，`Available()` 会检查会话总线和服务能力，可能等待总线回应；`RequestPermission` 异步执行同样检查，不弹权限对话框。无服务返回 ErrUnsupported，协议调用失败返回 ErrFailed。Notify 使用系统默认超时、无图标/动作、应用名 Keel；正文根据 body-markup 能力转义，保留纯文本含义。

业务 ID 映射到服务返回的数值 ID，重复 Post 使用 replaces_id，Remove 调用 CloseNotification。收到 NotificationClosed 后删除映射；服务 owner 变化时清空旧映射，避免将旧 ID 发送给重启后的服务。映射仅保留在当前进程，不能跨进程启动撤回旧通知。Linux 已实现 ActionInvoked 默认动作回调；kit 示例连接窗口 Raise，尚未处理 Wayland ActivationToken，实际窗口置前取决于桌面策略。真实桌面展示仍需运行示例验收。协议测试使用可控总线替身，不代表 Linux 桌面验收完成。

协议依据：[Freedesktop Desktop Notifications](https://specifications.freedesktop.org/notification/latest/protocol.html)。


macOS 可设置 `Message.OnClick`。收到系统默认打开动作后，回调在独立 goroutine 执行一次；UI 修改需 `core.Update`。示例“原生点击回调”可手动验证。Linux 服务声明 actions 时也支持 OnClick；不支持 actions 的 Linux 服务及其他未实现平台返回 ErrUnsupported，不会静默丢弃回调。普通通知仍按各平台能力投递。

首次 Post 在主队列安装本模块 delegate；已有其他 delegate 时返回 ErrConflict，不覆盖它。之后应用也不应另行替换 delegate。只有带本模块标记的通知会请求前台展示或触发回调，其他通知不处理。这里只接收当前进程投递后的响应，未实现冷启动/跨启动恢复。

同 ID 替换使用新回调；替换失败恢复旧回调，迟到的旧请求结果不会覆盖更新的注册。回调在点击、成功 Remove 或无回调替换后释放。仅在通知中心手动忽略/关闭横幅可能不会产生默认打开动作，应用应适时 Remove，释放仍保留的回调。回调已经开始执行后，Remove 不能取消其执行。

验证范围：编译、未打包进程拒绝路径、注册表替换/回退/并发消费和 race 测试；尚未完成真实通知横幅和系统点击验收。


Linux 默认动作使用 `default` 标识，监听 ActionInvoked 并消费对应业务 ID 的回调一次；仅接受当前服务 owner 的信号。NotificationClosed、成功 Remove、无回调替换和服务重启都会释放注册。信号使用顺序处理器，保留打开后关闭的处理顺序；回调在连接锁外的独立 goroutine 执行，可以继续投递或撤回。失败替换不丢失原有回调。当前没有接入 ActivationToken，无法保证 Wayland 上点击后窗口置前。
