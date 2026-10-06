# native

English | [简体中文](README.zh-CN.md)

A collection of system capability modules. Each subdirectory is an independent module, you can use only one of them, or you can not use the `ui` interface at all.

| Modules | Capabilities | Dependencies |
| --- | --- | --- |
| [permission](permission/README.md) | Check and apply for system permissions | `native`, `native/internal/sys` |
| [screen](screen/README.md) | Monitor list, screenshots | Same as above |
| [input](input/README.md) | Synthesize mouse and keyboard events | Same as above |
| [hotkey](hotkey/README.md) | Global shortcut keys | Same as above |
| [clipboard](clipboard/README.md) | Asynchronously read clipboard text, encoded images, and file references (macOS / Windows / Linux Wayland, X11) | Same as above |
| [notification](notification/README.md) | System notification permissions, delivery and withdrawal (macOS .app / Linux D-Bus) | Same as above |

Modules do not reference each other, nor do they reference `ui` and Gio. This is guaranteed by the tests in `internal/deps`.

The two shared parts are not independent modules:

- `native` (`native.go` for this directory): Error value `native.Err*` common to all modules.
- `native/internal/sys`: Platform bindings, macOS's cgo and Objective-C, Windows' Win32 calls, Linux's X11 protocol and D-Bus are all here. Cannot be referenced externally.
- `native/internal/wlclip`: The cgo binding of the Linux Wayland clipboard is only referenced by clipboard, so other modules will not link to libwayland.

Supports macOS 14+, Windows, and Linux, other platforms return `native.ErrUnsupported`. On Linux, screenshots and composite input go through X11 (via XWayland in Wayland sessions), notifications go through D-Bus, and the clipboard is supported by both Wayland and X11. For details on each ability on each platform, see [Native ability](../docs/native.md).

Document: [Native ability](../docs/native.md) · [New native ability](../docs/extending.md#add-native-capabilities)
