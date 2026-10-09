# Android

[English](android.md) | 简体中文

`keel build -target android` 把 Go 界面打包成开发签名 APK；`keel run -target android` 按设备 ABI 构建、覆盖安装并启动应用。运行支持已启动的模拟器和已连接、开启 USB 调试的真机。

## 环境和命令

需要 Go 1.26+、Android SDK 的 platforms、build-tools 35+、NDK 和 JDK。运行还需要 platform-tools 中的 `adb`。通过 Android Studio SDK Manager 安装，或使用 `sdkmanager`。设置 `ANDROID_HOME` 指向 SDK 根目录；也接受 `ANDROID_SDK_ROOT`。设置 `JAVA_HOME` 或把 Java 工具放进 `PATH`。

```sh
export ANDROID_HOME=/path/to/android-sdk
export JAVA_HOME=/path/to/jdk
keel new my-app -appid com.example.myapp
cd my-app
keel doctor -target android
keel build -target android
adb devices
keel run -target android
```

构建通过固定版本 `gioui.org/cmd/gogio@v0.10.0` 完成，NDK 优先使用 SDK 下最新的 side-by-side 版本，再检查 `ndk-bundle` 和 `ANDROID_NDK_ROOT`。工具链安装说明见 [Gio Android 文档](https://gioui.org/doc/install/android)；具体兼容性以固定版本打包器和实际构建为准。

产物为 `dist/android/<binary>.apk`。默认包含 arm64 和 amd64，兼顾常见真机及模拟器；可用 `-arch arm64` 缩小包体，也接受 `arm`、`386` 和逗号分隔的架构。`-o` 改变输出根目录，`-n` 打印命令。

`run` 只在恰好一个已就绪设备时自动选择；多个设备连接时指定序列号，离线或未授权设备会提示错误。CLI 不创建 Android 模拟器，请先从 Android Studio Device Manager 启动。

```sh
keel run -target android -serial emulator-5554
adb -s emulator-5554 logcat
```

启动成功后命令返回，日志通过 `adb logcat` 查看。Android 不把 `run -- ...` 参数传入 Go，CLI 会拒绝这些参数。Gio 的 Android 打包器固定剥离符号，因此 `-debug` 不能恢复 Android 原生库的完整调试符号。

## 配置、图标和签名

```json
"android": {
  "minimum_sdk": 23,
  "target_sdk": 35
}
```

旧项目省略 `android` 也使用这些默认值。`minimum_sdk` 至少为 23；`target_sdk` 至少为 31，且不低于最低版本，安装的 platform SDK 必须覆盖目标版本。`appid` 必须为有效的 Java 包名，不能包含连字符；新项目生成的默认 ID 会移除目录名中的连字符和下划线。

应用名、ID、版本、构建号来自 `keel.json`，其中 `build` 是 Android 的 versionCode。图标使用满版原图，由打包器生成各密度图标和 adaptive icon，`icon_mask` 不影响 Android；可用 `icons.android` 指定单独原图。`keel icon` 输出 `dist/icons/android.png` 预览。

CLI 在 `dist/android/debug.keystore` 创建并复用开发密钥，密码为公开的 `android`，只用于开发安装。保留密钥才能覆盖同一签名的旧包；清理 `dist` 或换 `-o` 会生成新密钥，此时需保留原密钥或先卸载旧包（卸载会清除应用数据）。CLI 当前不提供 Android 发布签名、AAB 或商店上传流程。

## 验证范围

```sh
go test ./cmd/keel
KEEL_ANDROID=1 go test ./cmd/keel -run TestNewAndroidProjectBuilds -count=1
```

第二项实际打包生成项目的 arm64 和 amd64 APK，并检查 Manifest、DEX、原生库；包含带空格的项目路径。移动端系统能力与交互边界见 [Mobile 支持](mobile.zh-CN.md)。

本次验证环境为 Apple Silicon、platform SDK 35、build-tools 35.0.0、NDK 27.0.12077973、JDK 21。已完成 arm64/amd64 APK 打包，以及 API 35 arm64 模拟器的覆盖安装、首屏、中文显示、按钮点击、软键盘拉丁字符输入、返回收起键盘和前后台恢复。build-tools 34 的 D8 在 JDK 21 下失败，请升级至 35 或更新版本。真机、拼音组合输入和完整窄屏组件交互仍需验收。

Android 会直接加载系统 `NotoSansCJK-Regular.ttc` 或 `NotoSansSC-Regular.otf`，避免默认字体扫描缺少可用缓存目录时中文显示方框。OEM 没有这些字体时，可在启动前用 `theme.LoadFonts` 加载应用自带的 CJK 字体。
