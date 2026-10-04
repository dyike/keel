# 脚手架与打包

`keel` 命令新建项目、运行、按平台打包，并把应用图标、名称、版本一起打进去。

```sh
go install github.com/dyike/keel/cmd/keel@latest
```

## 新建项目

```sh
keel new my-notes
cd my-notes
keel run
```

生成的目录：

| 文件 | 内容 |
| --- | --- |
| `main.go` | 打开窗口，挂上界面 |
| `app.go` | 界面：一个视图结构体和它的 `Render` |
| `keel.json` | 应用名、ID、版本、图标，打包时读取 |
| `appicon.png` | 1024×1024 的占位图标，替换成自己的 |
| `go.mod` | 依赖当前版本的 Keel，生成后自动 `go get` 和 `go mod tidy` |
| `.gitignore`、`README.md` | 忽略 `dist/` 和打包中间文件 |

参数：`-name "我的笔记"` 设置显示名称（默认由目录名得到，`my-notes` → `My Notes`），`-appid com.yourname.notes` 设置应用 ID（默认 `com.example.<目录名>`），`-module` 设置 Go 模块路径，`-offline` 只写文件不拉依赖。目录必须为空。

## keel.json

```json
{
  "name": "My Notes",
  "appid": "com.example.my-notes",
  "version": "0.1.0",
  "build": 1,
  "binary": "my-notes",
  "icon": "appicon.png",
  "main": "."
}
```

- `name`：给人看的名字，用作 `.app` 名、菜单项和 Linux 启动器里的名字。
- `appid`：反向域名形式的唯一 ID。macOS 的权限授权、系统通知按它记录，Linux 桌面按它匹配图标，**发布后不要再改**。
- `version` 是 `主.次.修订`，`build` 是同一版本的构建号；macOS 写进 Info.plist，Windows 写进 .exe 的版本信息。
- `binary` 是可执行文件名，`main` 是 main 包相对 `keel.json` 的路径。

## 运行

`keel run` 等同于 `go run`，并把 `appid` 告诉 Gio（Linux 上窗口的 app_id）。`keel run -- --flag` 把 `--` 之后的参数交给应用。

## 打包

```sh
keel build                    # 当前平台
keel build -target windows    # 在任何系统上都能打 Windows 包
keel build -target js         # WebAssembly
keel build -n                 # 只打印要执行的命令
```

输出在 `dist/`（`-o` 可改）。第一次打包会下载 Gio 的打包工具 gogio。

| 目标 | 产物 | 图标 | 能在哪里打包 |
| --- | --- | --- | --- |
| `darwin` | `dist/My Notes.app` | 生成 `icon.icns` 写进包里 | macOS（需要 cgo 和 `iconutil`） |
| `windows` | `dist/my-notes.exe` | 嵌进 .exe 的资源段 | 任何系统（不需要 cgo） |
| `linux` | `dist/linux/`：程序、`<appid>.desktop`、各尺寸图标、`install.sh` | 装进 hicolor 图标主题 | Linux（需要 Wayland/X11 开发头文件） |
| `js` | `dist/web/` | — | 任何系统 |

**macOS**：`-arch arm64,amd64` 打通用包（默认本机架构）。脚手架会重写 Info.plist（应用类型、名称、版本、最低 macOS 14），再给整个 .app 签名：不给身份时是只在本机有效的临时签名；要分发给别人，用 `-sign "Developer ID Application: 你的名字 (TEAMID)"` 签名，再用 `xcrun notarytool` 公证。

**Windows**：默认 `-arch amd64`。gogio 打包时会在源码目录生成 `*_windows_*.syso` 资源文件，`keel build` 结束后会删掉它（打包前就存在的不动）。

**Linux**：程序编译时带上 `appid`，Wayland 的 app_id 和 X11 的 WM_CLASS 都等于它，桌面环境据此把窗口和 `<appid>.desktop`、图标对上。`dist/linux/install.sh` 默认安装到 `~/.local`，`PREFIX=/usr/local sudo ./install.sh` 装到系统目录。

## 图标

换图标只要替换 `appicon.png`：正方形 PNG，建议 1024×1024，背景透明。macOS 的图标规范在四周留约 10% 空白（主体 824×824 居中），占位图标就是这样画的；小于 512px 时打包会提醒。

## 检查环境

`keel doctor` 检查 Go 版本和当前系统打包需要的工具：macOS 上的 Xcode 命令行工具和 `iconutil`，Linux 上 Wayland、X11、xkbcommon、EGL 的开发包。
