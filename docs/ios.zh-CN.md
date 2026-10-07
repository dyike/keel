# iOS

[English](ios.md) | 简体中文

脚手架支持从同一份 Go 界面代码构建 iOS 应用：`keel build -target ios` 打包模拟器 `.app`，`keel run -target ios` 构建、安装并运行；真机通过签名身份和描述文件生成 `.ipa`。当前已验证 Apple Silicon 上的 iOS 27 模拟器，真机签名和移动端交互仍需验收。

## 创建、构建和运行

需要 macOS、完整 Xcode（仅 Command Line Tools 不够）、Go 1.26+ 和已安装的 iOS 模拟器运行时。默认最低系统版本为 iOS 18.0。

```sh
keel new my-app
cd my-app
keel doctor -target ios
keel build -target ios
keel run -target ios
```

开发本地 Keel 时，先在仓库构建 CLI，再通过 `-replace` 让生成项目使用这份源码：

```sh
go build -o /tmp/keel ./cmd/keel
/tmp/keel new /tmp/my-app -replace "$PWD"
cd /tmp/my-app
/tmp/keel run -target ios
```

模拟器包输出到 `dist/ios/<binary>.app`，包含 `keel.json` 配置的应用名、Bundle ID、版本、构建号和图标，使用本地临时签名。`-o` 可改变输出根目录；`-debug` 保留调试信息；`-n` 只打印构建流程。

`run` 默认使用唯一已启动的 iOS 模拟器；没有已启动设备时优先选择最新运行时的 iPhone。多个设备已启动时，必须指定 UDID。它会启动模拟器、等待就绪、打开 Simulator 或 DeviceHub、安装应用，并将应用日志连接到终端。

```sh
xcrun simctl list devices available
keel run -target ios -simulator <UDID>
keel run -target ios -simulator <UDID> -- --my-app-flag
```

模拟器默认按 Mac 架构构建，可显式指定 `-arch arm64` 或 `-arch amd64`，每次只支持一种架构。真机固定使用 arm64。

## 配置和图标

新项目会写入以下配置。旧项目可以不加 `ios`，同样默认使用 18.0。最低版本配置接受 13.0 及以上版本；降低版本后需要自行验证目标系统。

```json
"ios": {
  "minimum_version": "18.0"
}
```

iOS 图标使用满版正方形原图，圆角交给系统处理，透明区域合成到白色背景；`icon_mask` 不影响 iOS。也可以用 `"icons": {"ios": "icon-ios.png"}` 指定单独原图。`keel icon` 会生成 `dist/icons/ios.png` 预览，打包时通过 Xcode `actool` 生成 iPhone/iPad 图标资源。

## 真机打包

需要钥匙串中带私钥的 Apple 签名身份，以及与 `appid` 匹配的 provisioning profile：

```sh
keel build -target ios -device \
  -sign "Apple Development: Your Name (TEAMID)" \
  -provision ./app.mobileprovision
```

产物为 `dist/ios/<binary>.ipa`。脚手架检查描述文件的应用标识，提取签名权限，处理通配符应用标识和钥匙串组，嵌入描述文件并签名。描述文件的设备范围、有效期、证书以及分发方式由开发者配置。当前没有完成真机安装或 App Store 上传验证，也没有自动上传流程。

## 构建和生命周期适配

当前 Gio 打包器将 arm64 当作真机架构，因此 iOS 目标直接调用 Go 和 Xcode 的对应 SDK。`gioui.org/shader v1.0.9` 仅通过 amd64 判断模拟器；CLI 在临时目录复制该模块，让 arm64 使用包内已有的模拟器 Metal 库，再通过临时 `go.mod` 构建。项目依赖文件和下载缓存保持原样，构建结束后清理副本。升级 Gio 或 shader 后需要重新验证适配。模拟器构建以项目的 `go.mod` 为准，忽略 `go.work`；工作区依赖应写入模块的 `replace`。

`ui/window/scene_ios.*` 提供单场景 UIScene 适配，以支持新 SDK 和 iOS 27。CLI 在 `Info.plist` 中配置 `KeelSceneDelegate`，将 Gio 的窗口创建推迟到场景连接、主事件循环开始之后。适配依赖 Gio v0.10 的 `_gioAppDelegate` 和 `GioViewController` 类名；找不到对应类或方法时直接报告不兼容。目前只支持单场景。

## 验证范围

已验证：生成项目实际构建、资源图标、Bundle 元数据、本地签名、自动安装启动和 Metal 首帧。运行以下检查：

```sh
go test ./cmd/keel
KEEL_IOS=1 go test ./cmd/keel -run TestNewIOSProjectBuilds -count=1
go test ./internal/deps -run TestBuildsForIOSSimulator -count=1
```

第二项在 Mac 上实际打包生成项目，包含带空格路径；第三项有 SDK 时交叉编译 `ui/*`、`native/*` 和 hello。实际模拟器 UI 验收仍包括按钮与复选框点击、软键盘、拼音输入、窄屏滚动、前后台恢复。Intel Mac、iPad 和真机也尚未验证。

`native/*` 的 iOS 后端尚未实现，权限、屏幕捕获、合成输入、全局快捷键、通知和富剪贴板返回 `native.ErrUnsupported`。Gio 输入框内部的文本复制粘贴需要单独验证。
