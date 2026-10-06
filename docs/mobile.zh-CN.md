# Mobile 支持

[English](mobile.md) | 简体中文

同一份 Keel Go 界面代码可以打包到 iOS 和 Android。移动端使用 Gio 的原生窗口、触摸和输入法，不需要 WebView。

| 平台 | 构建 | 运行 | 构建环境 |
| --- | --- | --- | --- |
| [iOS](ios.zh-CN.md) | `keel build -target ios` 输出模拟器 `.app`；`-device` 输出签名 `.ipa` | `keel run -target ios` 自动选择、启动并安装到 iOS 模拟器 | macOS、完整 Xcode、iOS SDK |
| [Android](android.zh-CN.md) | `keel build -target android` 输出开发签名 `.apk` | `keel run -target android` 安装到已启动的模拟器或已连接真机 | Go、Android SDK、NDK、JDK；运行还需要 adb |

先运行 `keel doctor -target ios` 或 `keel doctor -target android` 检查工具链。旧项目无需补充移动端配置即可构建：iOS 默认最低 18.0，Android 默认最低 API 23、目标 API 35。

## 界面和系统能力

用实际屏幕宽度安排布局；桌面的 `Width`、`Height`、窗口装饰、多窗口和全局快捷键不能作为移动端行为保证。触摸操作不要只依赖悬停提示或右键菜单，表单应检查软键盘弹出后的可见区域。

`native/*` 的移动端后端尚未实现。权限、屏幕捕获、合成输入、全局快捷键、系统通知和富剪贴板在 iOS、Android 返回 `native.ErrUnsupported`。Gio 输入框内的文本编辑与 `native/clipboard` 是不同路径。

iOS 已验证 Apple Silicon 上的 iOS 27 模拟器构建、安装启动和 Metal 首帧。Android 的构建、运行命令和验证范围见 [Android](android.zh-CN.md)。发布前仍需在目标设备验收软键盘、拼音输入、窄屏滚动、系统返回和前后台恢复。移动端支持不代表每个桌面示例都已经完成窄屏布局。
