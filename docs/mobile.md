# Mobile

English | [简体中文](mobile.zh-CN.md)

The same Keel Go UI can be packaged for iOS and Android. Mobile apps use Gio's native window, touch input and IME, without a WebView.

| Platform | Build | Run | Build environment |
| --- | --- | --- | --- |
| [iOS](ios.md) | `keel build -target ios` produces a simulator `.app`; `-device` produces a signed `.ipa` | `keel run -target ios` selects, boots and installs on an iOS simulator | macOS, full Xcode and iOS SDKs |
| [Android](android.md) | `keel build -target android` produces a development-signed `.apk` | `keel run -target android` installs on a running emulator or connected device | Go, Android SDK, NDK and JDK; adb is needed to run |

Check the toolchain with `keel doctor -target ios` or `keel doctor -target android`. Existing projects can omit mobile configuration: iOS defaults to minimum version 18.0; Android defaults to minimum API 23 and target API 35.

## UI and system capabilities

Design for the actual screen width. Desktop window dimensions, decorations, multiple windows and global hotkeys are not mobile behavior guarantees. Touch interactions should not depend solely on hover or context menus; check that form fields remain visible when the soft keyboard opens.

Mobile backends for `native/*` are not implemented. Permissions, screen capture, synthetic input, global hotkeys, system notifications and rich clipboard operations return `native.ErrUnsupported` on iOS and Android. Gio's text editing follows a separate path from `native/clipboard`.

iOS simulator packaging, installation, launch and the first Metal frame have been verified on Apple Silicon with iOS 27. See [Android](android.md) for its commands and verification scope. Before release, validate the soft keyboard, composition input, narrow-screen scrolling, system back navigation and background/foreground recovery on your target devices. Mobile support does not mean every desktop example has a mobile layout.
