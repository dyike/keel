# notification

独立的系统通知接口，当前实现 macOS 14+、cgo、带 bundle identifier 的 `.app`。Windows、Linux、WebAssembly、iOS 和无 cgo 构建返回 `native.ErrUnsupported`，普通 `go run` 同样不支持。

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
- macOS 必须运行应用事件循环；权限对话框、专注模式、系统设置和签名信任影响实际显示。当前尚无通知中心 delegate，未实现前台展示控制、点击响应、应用/窗口激活和跨启动回调恢复。
- 尚未与 kit.Notifier 连接，不能把本模块存在理解为 Notification 组件已经完成系统投递。

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

此临时应用包用于开发验证；正式分发需要正常签名和可信安装位置。点击“申请通知权限”，允许后投递，切到其他应用观察通知中心；再次投递验证替换，再点撤回验证消失。按钮串行化请求，并显示真实错误。`-check` 只打印平台/包身份检查结果，不申请权限、不发送通知。

自动测试覆盖参数校验和未打包进程的拒绝路径，不弹出系统权限窗口。系统授权、真实通知展示/替换/撤回仍需上述手动验收。

接口依据：[Apple UNUserNotificationCenter](https://developer.apple.com/documentation/usernotifications/unusernotificationcenter)。
