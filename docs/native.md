# Native APIs

English | [简体中文](native.zh-CN.md)

Packages under `native/` provide system capabilities not found in Gio. They do not depend on interface modules and can be used independently.

| Packages | Capabilities | Required Permissions |
| --- | --- | --- |
| `native/permission` | Check, request permissions | None |
| `native/screen` | List monitors; screenshots | Screenshots require screen recording |
| `native/input` | Moving the mouse, clicking, and pressing keys | Operations other than reading the mouse position require accessibility functions |
| `native/hotkey` | Global shortcut keys | None |
| `native/clipboard` | Asynchronously read text, encoded images, file paths (macOS / Windows / Linux Wayland, X11) | None |

Three platforms are supported. On other platforms and macOS built with cgo turned off, all functions return `native.ErrUnsupported` and the program compiles as usual.

| Platform | Implementation | Description |
| --- | --- | --- |
| macOS 14+ | cgo calls the system framework | User authorization is required, see below |
| Windows 10+ | Directly call user32, gdi32, no cgo required | No authorization required; coordinates are converted into logical points according to the DPI of the main display |
| Linux | Pure Go X11 protocol implementation; no cgo | Requires an X11 session and the X server’s XTEST extension for synthetic input; logical points are converted using `Xft.dpi` |

Wayland sessions for Linux do not allow normal programs to capture the entire screen or inject input into other programs. It can be called when there is XWayland (`DISPLAY` is set), but it can only see and operate the window of the X11 program; when there is no `DISPLAY`, it returns `ErrUnsupported`.

## Errors

All errors are packaged from the sentinel value in the `native` package, and are judged using `errors.Is`:

| Error | Meaning |
| --- | --- |
| `ErrUnsupported` | Not implemented on the current platform |
| `ErrPermissionDenied` | Missing required permissions |
| `ErrInvalidArgument` | The parameter is illegal, such as unknown button name, display ID is 0 |
| `ErrTimeout` | System call timeout (screenshot exceeds 10 seconds) |
| `ErrConflict` | The global shortcut key has been occupied by this process or other programs |
| `ErrFailed` | Other system errors, error messages with status codes |

```go
if _, err := screen.Capture(id); errors.Is(err, native.ErrPermissionDenied) {
    permission.Request(permission.ScreenRecording)
}
```

## Permission: permission

```go
ok, err := permission.Granted(permission.Accessibility) // Only check, no pop-ups
ok, err := permission.Request(permission.Accessibility) // A system authorization box may pop up
```

| Constant | Name in system settings | Who needs it |
| --- | --- | --- |
| `Accessibility` | Accessibility | `input` Package |
| `ScreenRecording` | Screen recording (called "Screen and System Recording" from macOS 15) | `screen.Capture` |
| `InputMonitoring` | Input monitoring | Currently no function is used, reserved |

Windows and Linux do not have these authorizations, and `Granted` and `Request` always return `true`.

There are three points to note when using macOS:

- **`Request` is not waiting for the user to answer.** It will return to the current authorization status immediately after popping up the box, usually `false`. After the user turns on the switch in the system settings, you need to adjust `Granted` again to confirm.
- **`false` does not distinguish between "rejected" and "not asked yet".** macOS does not provide this information.
- **Which program name is the authorization recorded under?** When packaged as `.app` and run, the authorization is recorded under the name of the app; in the terminal `go run`, macOS usually records the authorization under the name of the terminal program (terminal, iTerm, VS Code). After screen recording permission is granted, the program generally needs to be restarted to take effect.

## Screen: monitor and screenshots

```go
displays, err := screen.Displays()
for _, d := range displays {
    fmt.Println(d.ID, d.X, d.Y, d.Width, d.Height, d.PixelWidth, d.PixelHeight, d.Primary)
}
```

`X`, `Y`, `Width`, `Height` are logical point coordinates. The origin is in the upper left corner of the main display. The secondary screen may be negative coordinates. `input` package uses the same set of coordinates.

```go
png, err := screen.Capture(d.ID)
```

- Returns PNG bytes with dimensions `PixelWidth × PixelHeight`, excluding the mouse pointer.
- Screen recording permission is required on macOS, and the pop-up box will not pop up. Without permission, it will directly return to `ErrPermissionDenied`.
- Block for up to 10 seconds. **Don't call it in a callback**, it will freeze all windows. Put it into goroutine, and use `core.Update` to send the result back to the interface.
- Cannot be called on the main thread on macOS, otherwise `ErrFailed` will be returned. The `main` function runs on the main thread before `window.Main()` and cannot be called there.

## Input: synthetic keyboard and mouse

```go
x, y, err := input.MousePosition()   // No permission required
input.MouseMove(100, 200)            // The following all require accessibility permissions
input.Click(input.Left)              // Click at the current mouse position; Left / Right / Middle
input.Tap("enter")                   // press and release
input.KeyDown("cmd"); input.Tap("c"); input.KeyUp("cmd")   // ⌘C
```

The key name is based on the physical location of the American keyboard and has nothing to do with the current input method and keyboard layout (Linux exception: X11 searches for the key that can type this character according to the current layout):

- Letters, numbers, symbols: `a`–`z`, `0`–`9`, `-` `=` `[` `]` `\` `;` `'` `,` `.` `/` `` ` ``
- Function keys: `enter` `tab` `space` `backspace` `delete` `escape` `home` `end` `pageup` `pagedown` `up` `down` `left` `right` `f1`–`f12`
- Modifier keys: `cmd` `shift` `alt` `ctrl`. On Windows and Linux `cmd` is the Windows key (Super).

`KeyDown` and `KeyUp` must be called in pairs, otherwise the system will think that this key is pressed all the time.

## Hotkey: global shortcut key

It can also be triggered when other apps are in the foreground. The shortcut key that takes effect only when the window has focus is [`window.Options.Shortcuts`](app.md#shortcuts).

```go
unregister, err := hotkey.Register("cmd+shift+k", func() {
    core.Update(func() { status.SetText("Triggered") })
})
defer unregister()
```

- The writing method is `modifier+key`, which requires at least one modifier key. Modifier keys: `cmd` `ctrl` `alt` (or `option`) `shift`; the key name is the same as `input` package.
- The callback is executed in an independent goroutine and does not hold the interface lock. **Changes to the interface must be included in `core.Update`**.
- If you press it several times while the callback is still executing, it will only be triggered once and will not be queued.
- Returns `ErrConflict` when the key combination is already occupied.
- `unregister` can be called repeatedly, and it will only take effect the first time.

`cmd` is the Windows key (Super) on Windows and Linux. Cross-platform shortcut keys are usually written as `cmd` for macOS and `ctrl` for other platforms.

Requires `window.Main()` to be running on macOS: shortcut key events are dispatched by the main thread's event loop. Windows and Linux have their own message threads and are not subject to this limitation. Pure background programs that do not open windows on macOS cannot be used temporarily. See [FAQ](troubleshooting.md#global-shortcut-keys-are-not-responding).

## Notification: system notification

`native/notification` provides Available, RequestPermission, Post and Remove, and the completion callback is executed in a separate goroutine. Currently, macOS .app authorization, delivery/replacement by ID, and withdrawal are implemented, Linux desktop D-Bus backend, and Windows tray bubble backend (display one at the same time, support click callback); other unsupported platforms explicitly return unsupported. kit.Notifier is accessed through the application adapter. macOS has implemented native foreground display and Message.OnClick. Kit can receive system clicks and request Window.Raise through the interactive backend. Linux has supported the default click of services declaring actions. The activation token brought by the system click can be handed over to `window.Window.Activate` (Wayland xdg-activation / X11 startup ID). For complete usage and acceptance steps, see [Module Document](../native/notification/README.md).

## Clipboard: rich clipboard snapshot

`clipboard.Read(func(data clipboard.Data, err error))` reads asynchronously; the completion callback is executed in the background goroutine, and the interface update needs to be transferred back to the UI frame. `Text` is text, `Images` is MIME and encoded bytes, `Files` is a path reference and does not open the file.

macOS uses AppKit, and images are PNG first and TIFF second. Windows uses Win32, reads Unicode text, PNG, CF_DIB, and CF_HDROP; DIB adds a BMP file header and returns as `image/bmp`, without decoding pixels. File references take precedence over image previews included with Explorer. The total size of text, path UTF-8 bytes and encoded images can be up to 16MiB, and the number of files can be up to 128; errors do not return partial results. Windows keeps the clipboard open during reading and copies all data before closing; it tries up to 8 times with 15ms intervals while the clipboard is occupied.

DIB supports RGB, palette and bitfield layout of 40/108/124 byte header; compressed bitmap and V5 layout with independent color profile are not supported yet. HTML/RTF has not yet been returned as a standalone format. iOS and macOS without cgo return `ErrUnsupported`.

The component library's Input/Textarea and CodeEditor examples already share this read adaptation; on failure, it pastes along the original Gio text path. Windows has passed format parsing, pixel decoding, error/upper limit testing and cross-compilation. System clipboard and real window pasting have not yet been accepted on Windows real devices.

Linux X11 uses an independent connection to read CLIPBOARD, supporting UTF-8/Latin-1 text, PNG/JPEG/TIFF/BMP/WebP encoded images, URI lists, and GNOME file references. Only local absolute paths are returned, remote file hosts and non-file URLs are ignored, and no copying or cutting is performed. Respects the same 16MiB/128 file limit; handles INCR chunked transfers, with a 5 second limit on entire connections and reads. Check the owner and its provided TIMESTAMP before and after reading; legacy applications that do not provide timestamps cannot guarantee cross-format atomic snapshots if content is changed within the same owner.

Connection uses DISPLAY and XAUTHORITY (default ~/.Xauthority), supports MIT-MAGIC-COOKIE-1.

Wayland only gives the clipboard to the client with keyboard focus, so it cannot open another connection: the application gives it the connection of the focused window, `clipboard.UseWaylandDisplay(w.WaylandDisplay())`, and then `Read` opens a private event queue on this connection, binds the data device of the seat, reads the type provided by the current selection (the same text, URI list and image format and restrictions as X11, 5-second period), without interfering with Gio's own event processing. Fallback to X11/XWayland when there is no incoming connection, or the compositor has no data device. This path links to libwayland-client and can be removed when building with `-tags nowayland`. The C code only uses the core protocol, is grouped by opcodes, has been syntax and cgo type checked on macOS with real header files, and has not yet been run on the Wayland desktop. Format, chunked protocol surrogate, error cap testing and cross-compilation passed, Linux desktop real clipboard and XWayland bridging are still pending acceptance.
