# 快速开始

[English](getting-started.md) | 简体中文

用 `keel` 脚手架创建应用，在生成的 `app.go` 中开发界面，再用同一套命令运行和打包。

## 环境

- Go 1.26.1 或更新。
- macOS：安装 Xcode Command Line Tools（`xcode-select --install`）。Gio 走 cgo（`native/` 不用）；原生能力要求 macOS 14+。
- Windows：安装 Go 即可开始；Linux 需要 Wayland/X11、xkbcommon、EGL 的开发包。各平台的原生能力见 [原生能力](native.zh-CN.md)。

## 检查环境

先安装脚手架，再检查当前系统需要的工具：

```sh
go install github.com/dyike/keel/cmd/keel@latest
keel doctor
```

macOS 会检查 Xcode 命令行工具和 `iconutil`；Linux 会检查窗口系统与字体相关的开发包。首次运行或打包需要下载依赖并编译 cgo。macOS 的 `-lobjc` 链接警告见 [常见问题](troubleshooting.zh-CN.md#链接时出现--lobjc-警告)。

## 新建项目

```sh
keel new my-notes
cd my-notes
keel run
```

生成的目录：

| 文件 | 内容 |
| --- | --- |
| `main.go` | 设置应用图标，打开窗口，挂上界面 |
| `app.go` | 界面：一个视图结构体和它的 `Render` |
| `keel.json` | 应用名、ID、版本、图标，打包时读取 |
| `appicon.png` | 1024×1024 的占位原图（满版，不带圆角），替换成自己的 |
| `go.mod` | 依赖当前版本的 Keel，生成后自动 `go get` 和 `go mod tidy` |
| `.gitignore`、`README.md` | 忽略 `dist/` 和打包中间文件 |

参数：`-name "我的笔记"` 设置显示名称（默认由目录名得到，`my-notes` → `My Notes`），`-appid com.yourname.notes` 设置应用 ID（默认 `com.example.<目录名>`），`-module` 设置 Go 模块路径，`-offline` 只写文件不拉依赖。目录必须为空。

## 写第一个窗口

在脚手架生成的项目中修改 `app.go`。`main.go` 继续负责应用图标、窗口和事件循环；不用重新手写入口。把 `app.go` 替换为：

```go
package main

import (
    "strings"

    "github.com/dyike/keel/ui/el"
    "github.com/dyike/keel/ui/kit"
    "github.com/dyike/keel/ui/theme"
)

type app struct {
    name   *kit.InputView
    result string
}

func newApp() *app {
    a := &app{
        name: kit.Input("你的名字").Placeholder("例如：小明"),
        result: "等待输入",
    }
    a.name.OnSubmit(func(string) { a.greet() })
    return a
}

func (a *app) greet() {
    if name := strings.TrimSpace(a.name.Value()); name != "" {
        a.result = "你好，" + name + "！"
    }
}

func (a *app) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Items(el.Start).Child(
        el.Text("第一个窗口").TextSize(22).Bold(),
        a.name.Render(cx),
        kit.Button("打招呼", a.greet).Render(cx),
        el.Text(a.result).TextColor(theme.Muted),
    )
}
```

保存后执行 `keel run`。输入名字，点击按钮或按回车，下面会显示问候语。

有状态的组件在 `newApp` 中创建一次，保存在视图字段里；`Render` 根据当前状态搭建元素树。回调只改字段，执行完会自动重绘。更多布局和事件写法见 [元素与视图](el.zh-CN.md)，后台更新遵守 [线程规则](architecture.zh-CN.md#线程规则)。

## keel.json

```json
{
  "name": "My Notes",
  "appid": "com.example.mynotes",
  "version": "0.1.0",
  "build": 1,
  "binary": "my-notes",
  "icon": "appicon.png",
  "main": ".",
  "ios": {"minimum_version": "18.0"},
  "android": {"minimum_sdk": 23, "target_sdk": 35}
}
```

- `name`：给人看的名字，用作 `.app` 名、菜单项和 Linux 启动器里的名字。
- `appid`：反向域名形式的唯一 ID。macOS 的权限授权、系统通知按它记录，Linux 桌面按它匹配图标，**发布后不要再改**。
- `version` 是 `主.次.修订`，`build` 是同一版本的构建号；macOS 写进 Info.plist，Windows 写进 .exe 的版本信息。
- `binary` 是可执行文件名，`main` 是 main 包相对 `keel.json` 的路径。

## 运行

桌面端 `keel run` 默认开启热更新：监听项目源码、资源和本地 `replace`/`go.work` 模块，保存后重新编译，编译成功再重启应用；编译失败时保留当前窗口，等待下一次修改。连续保存会合并触发，构建产物放在临时目录。macOS/Linux 下，重启前会执行窗口的 `OnClose` 回调，供应用保存持久化状态；普通内存状态仍会重置。用户关闭最后一个窗口或按 Cmd+Q 正常退出应用时，`keel run` 也会结束监听；即使正在编译，也会取消构建，不重新打开窗口。应用崩溃后仍会等待源码修改。按 Ctrl+C 停止，或使用 `keel run -watch=false` 关闭监听。

`keel run` 把 `appid` 告诉 Gio，并使用 `keel.json` 配置的图标，遵循 `icon_mask` 和 `icons` 中的平台覆盖配置。macOS 启动签名后的临时应用包。Windows 与 `keel build` 共用 PerMonitorV2 DPI 清单、Common Controls 6、长路径支持、图标和版本资源，然后启动 GUI EXE；临时 `.syso` 在编译结束后清理。热更新和 `-watch=false` 都走这套打包路径。应用主动调用的 `window.SetIcon` 优先。Wayland 的图标要求见[窗口与应用 · 应用图标](app.zh-CN.md#应用图标)。`keel run -- --flag` 把 `--` 之后的参数交给应用。

iOS 模拟器使用 `keel run -target ios`：自动构建、安装和启动，`-simulator <UDID>` 选择设备。环境检查用 `keel doctor -target ios`，详见 [iOS](ios.zh-CN.md)。

Android 使用 `keel run -target android` 安装并启动，多个设备时用 `-serial <serial>` 选择。先执行 `keel doctor -target android` 检查工具链，详见 [Mobile 支持](mobile.zh-CN.md)。

## 打包

```sh
keel build                    # 当前平台
keel build -target windows    # 在任何系统上都能打 Windows 包
keel build -target js         # WebAssembly
keel build -target ios        # macOS + 完整 Xcode，模拟器 .app
keel build -target android    # Android SDK + NDK + JDK，开发 APK
keel build -n                 # 只打印要执行的命令
```

输出在 `dist/`（`-o` 可改）。打包默认去掉符号表、调试信息和本机路径（`-s -w -trimpath`），体积约小四分之一，崩溃时仍会打印函数名；要用调试器时加 `-debug` 保留。Android 原生库会被 gogio 固定剥离符号，见 [Android](android.zh-CN.md)。代码高亮和网络图片各约 4 MB，默认不链接，生成的 `main.go` 里有两行注释，用到时取消注释，见 [kit · 可选功能与体积](kit.zh-CN.md#可选功能与体积)。macOS、Android 和浏览器第一次打包会下载 Gio 的打包工具 gogio。

| 目标 | 产物 | 图标 | 能在哪里打包 |
| --- | --- | --- | --- |
| `darwin` | `dist/My Notes.app` | 生成 `icon.icns` 写进包里 | macOS（需要 cgo 和 `iconutil`） |
| `windows` | `dist/my-notes.exe` 和 `my-notes.ico` | 14 个尺寸嵌进 .exe | 任何系统（不需要 cgo） |
| `linux` | `dist/linux/`：程序、`<appid>.desktop`、各尺寸图标、`install.sh` | 装进 hicolor 图标主题 | Linux（需要 Wayland/X11 开发头文件） |
| `ios` | `dist/ios/my-notes.app`；`-device` 输出签名 `.ipa` | iPhone/iPad 图标资源 | macOS（完整 Xcode） |
| `android` | `dist/android/my-notes.apk` | 各密度图标与 adaptive icon | Android SDK、NDK、JDK |
| `js` | `dist/web/` | — | 任何系统 |

**macOS**：`-arch arm64,amd64` 打通用包（默认本机架构）。脚手架会重写 Info.plist（应用类型、名称、版本、最低 macOS 14），再给整个 .app 签名：不给身份时是只在本机有效的临时签名；要分发给别人，用 `-sign "Developer ID Application: 你的名字 (TEAMID)"` 签名，再用 `xcrun notarytool` 公证。

**Windows**：默认 `-arch amd64`，`-arch amd64,arm64` 时每个架构一个 .exe。打包过程中临时写入的 `.syso` 资源文件，编译完就删掉。

**Linux**：程序编译时带上 `appid`，Wayland 的 app_id 和 X11 的 WM_CLASS 都等于它，桌面环境据此把窗口和 `<appid>.desktop`、图标对上。`dist/linux/install.sh` 默认安装到 `~/.local`，`PREFIX=/usr/local sudo ./install.sh` 装到系统目录。

## 图标

`appicon.png` 是**原图**：正方形 PNG，1024×1024，满版（不要自己加圆角、留边或阴影）。`keel build` 按各平台的规范把它裁成各自的形状：

| 平台 | 画布 | 图形主体 | 圆角 | 其他 | 依据 |
| --- | --- | --- | --- | --- | --- |
| macOS | 1024 | 824×824 居中，四周留 100 | 185.4，连续曲率（和系统图标一样平滑过渡，不是普通圆弧） | 阴影：向下 12、模糊 28、黑色 50% | Apple 的 macOS 应用图标模板 |
| Windows | 48 网格 | 42×42 居中 | 2（外圆角） | 透明背景，无阴影；.exe 里嵌 16、20、24、30、32、36、40、48、60、64、72、80、96、256 共 14 个尺寸 | 微软的应用图标设计与构建指南 |
| Linux | 128 | 104×104 居中，四周留 12 | 8 | 无阴影；导出 hicolor 的 16–512 共 8 个尺寸 | GNOME 应用图标模板 |

每个尺寸都从原图直接画到目标大小，不从大图缩小，所以 16、24 这些小图标也清晰。

先看效果再打包：`keel icon` 把各平台的图标都写到 `dist/icons/`（`android.png`、`ios.png`、`macos.png`、`windows.ico` 和 `windows/<尺寸>.png`、`linux/<尺寸>.png`）。

iOS 使用满版原图，系统裁圆角，透明区域合成到白色背景，`icon_mask` 不影响 iOS。可通过 `icons.ios` 指定单独原图。

两种例外情况，在 `keel.json` 里配置：

- 原图本身就是透明背景的图形（比如一个圆形 logo，不想放在底板上）：`"icon_mask": "none"`，只按各平台的主体区域缩放，保留原图的轮廓和透明度。
- 设计师已经为某个平台做好了成品图标：`"icons": {"darwin": "icon-mac.png", "windows": "icon-win.png", "linux": "icon-linux.png"}`，指定的平台直接使用成品（正方形 PNG，只缩放不再加工），其余平台照常生成。

Windows 的 .exe 由 `keel build` 自己写资源：图标、清单（按显示器 DPI 缩放 PerMonitorV2、Common Controls 6、长路径）和版本信息（产品名、文件说明、版本号），再用 `go build -H=windowsgui` 编译；同一个图标另存为 `dist/<程序名>.ico`，给安装程序和快捷方式用。main 包目录里已经有其他 `.syso` 时会报错，避免资源冲突。

## 开发本地 Keel

修改框架后在应用里试效果，仍通过脚手架创建项目，用 `-replace` 指向本地仓库：

```sh
keel new my-app -replace ../keel
cd my-app
keel run
```

脚手架会在应用的 `go.mod` 中写入本地替换。已有脚手架项目可以执行 `go mod edit -replace=github.com/dyike/keel=../keel` 和 `go mod tidy`。

## Gio 的导入路径

Keel 用自带的 Gio 副本 `github.com/dyike/keel/third_party/gio` 绘制界面，这样能直接修复和优化 Gio，不必等上游发版。代码里和 Keel 一起使用的 Gio 包（`layout`、`op`、`unit` 等）从这里导入。之前导入 `gioui.org/...` 的项目切换一次：

```sh
keel migrate
```

它改写项目里 Go 文件的导入路径，包括 go-text 的（`github.com/go-text/typesetting` → `github.com/dyike/keel/third_party/typesetting`），再运行 `go mod tidy`。`gioui.org/shader` 保持不变。两套混用时，凡是 Keel API 接收 Gio 类型的地方都编译不过：它们是不同的包。

## 下一步

- [组件参考](kit.zh-CN.md)：按用途找组件，操作在线示例，查看展示代码。
- [元素与视图](el.zh-CN.md)：布局、状态、事件、焦点和浮层。
- [窗口与应用](app.zh-CN.md)：多窗口、快捷键和截图。
- [在浏览器里运行](web.zh-CN.md)：用 `keel build -target js` 构建 WebAssembly。
