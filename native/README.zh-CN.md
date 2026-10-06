# native

[English](README.md) | 简体中文

系统能力模块集合。每个子目录是一个独立模块，可以只用其中一个，也可以完全不用 `ui` 界面。

| 模块 | 能力 | 依赖 |
| --- | --- | --- |
| [permission](permission/README.zh-CN.md) | 检查、申请系统权限 | `native`、`native/internal/sys` |
| [screen](screen/README.zh-CN.md) | 显示器列表、截图 | 同上 |
| [input](input/README.zh-CN.md) | 合成鼠标、键盘事件 | 同上 |
| [hotkey](hotkey/README.zh-CN.md) | 全局快捷键 | 同上 |
| [clipboard](clipboard/README.zh-CN.md) | 异步读取剪贴板文本、编码图片和文件引用（macOS / Windows / Linux Wayland、X11） | 同上 |
| [notification](notification/README.zh-CN.md) | 系统通知权限、投递与撤回（macOS .app / Linux D-Bus） | 同上 |

模块之间互不引用，也不引用 `ui` 和 Gio，这一点由 `internal/deps` 里的测试保证。

两个共享部分不是独立模块：

- `native`（本目录的 `native.go`）：所有模块共用的错误值 `native.Err*`。
- `native/internal/sys`：平台绑定，macOS 的 cgo 与 Objective-C、Windows 的 Win32 调用、Linux 的 X11 协议和 D-Bus 都在这里。外部不能引用。
- `native/internal/wlclip`：Linux Wayland 剪贴板的 cgo 绑定，只有 clipboard 引用，所以别的模块不会链接 libwayland。

支持 macOS 14+、Windows 和 Linux，其他平台返回 `native.ErrUnsupported`。Linux 上截图和合成输入走 X11（Wayland 会话里通过 XWayland），通知走 D-Bus，剪贴板 Wayland 和 X11 都支持。各能力在各平台的细节见[原生能力](../docs/native.zh-CN.md)。

文档：[原生能力](../docs/native.zh-CN.md) · [新增原生能力](../docs/extending.zh-CN.md#新增原生能力)
