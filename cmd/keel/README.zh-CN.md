# cmd/keel

[English](README.md) | 简体中文

Keel 的脚手架：`keel new` 新建项目，`keel run` 运行，`keel build` 按平台打包（带图标、名称、版本），`keel doctor` 检查环境。用法见 [快速开始](../../docs/getting-started.zh-CN.md)。

macOS 的 `keel run` 构建带配置名称和图标的临时应用包，并做本地签名。热重载和 `-watch=false` 都直接启动包内的可执行文件，保留工作目录、参数、标准输入输出和信号处理；应用退出后清理包和图标。

Windows 的 `keel run` 和 `keel build` 使用同一套原生资源：PerMonitorV2 DPI 清单、Common Controls 6、长路径支持、图标和版本。热重载和 `-watch=false` 都启动带资源的 GUI EXE；临时 `.syso` 在编译结束后清理，已有 `.syso` 时报告冲突并保留用户文件。没有默认图标的旧项目仍会嵌入清单和版本。

```sh
go install github.com/dyike/keel/cmd/keel@latest
```

- **依赖**：`internal/svgicon`（把占位原图的 SVG 画成 PNG）、`internal/appicon`（各平台图标形状，`ui/window` 运行时也用它）、`golang.org/x/image`（缩放、矢量光栅化）、`github.com/tc-hib/winres`（Windows 资源：图标、清单、版本信息）。不引用 Keel 的界面包；macOS、Android 和浏览器打包调用 Gio 的 gogio（固定在 v0.10.0），macOS 签名调用系统的 `codesign`。
- **测试**：`go test ./cmd/keel` 生成项目、检查各平台的打包命令（`-n`），并用本仓库的 Keel 编译生成的项目（`-short` 跳过）。

| 文件 | 内容 |
| --- | --- |
| `main.go` | 子命令分发，运行外部命令 |
| `config.go` | `keel.json` 读写与校验 |
| `new.go` | 新建项目，模板在 `template/` |
| `run.go`、`run_bundle_*.go` | `keel run` 参数、开发图标和 macOS 应用包 |
| `run_watch.go`、`run_process_*.go` | 文件监听、重新编译和进程管理 |
| `build.go` | 各平台打包、Windows 资源、Info.plist、Linux 桌面文件 |
| `icons.go` | 按平台和尺寸生成图标、`keel icon`、.ico |
| `doctor.go` | 环境检查 |
| `migrate.go` | `keel migrate`：把应用里的 `gioui.org` 和 go-text 导入改到 Keel 在 `third_party` 下的副本 |

iOS 构建直接调用 Go 和 Xcode 的 SDK、`actool`、签名工具；运行调用 `simctl`，不依赖 Python。`ios.go` 负责打包、Metal 兼容、图标和签名，`ios_run.go` 负责模拟器选择、安装和启动。完整打包验证：`KEEL_IOS=1 go test ./cmd/keel -run TestNewIOSProjectBuilds -count=1`。用法见 [iOS](../../docs/ios.zh-CN.md)。

Android、macOS 和 Web 构建使用 `third_party/gio/cmd/gogio` 里的 Gio 打包器，按应用所依赖的 Keel 模块编译；`android.go` 负责开发签名、APK 打包、adb 设备选择和启动，`android_doctor.go` 检查 SDK、NDK、Java 与 adb。完整打包验证：`KEEL_ANDROID=1 go test ./cmd/keel -run TestNewAndroidProjectBuilds -count=1`。用法见 [Mobile 支持](../../docs/mobile.zh-CN.md) 和 [Android](../../docs/android.zh-CN.md)。
