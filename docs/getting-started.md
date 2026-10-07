# Getting started

English | [简体中文](getting-started.zh-CN.md)

Create an application with the `keel` scaffold, develop its UI in the generated `app.go`, and use the same tool to run and package it.

## Environment

- Go 1.26.1 or newer.
- macOS: install Xcode Command Line Tools (`xcode-select --install`). Gio and `native/` use cgo; native capabilities require macOS 14+.
- Windows: install Go. Linux also requires development packages for Wayland/X11, xkbcommon, and EGL. See [Native APIs](native.md) for platform support.

## Check the environment

Install the CLI, then check the tools required by your system:

```sh
go install github.com/dyike/keel/cmd/keel@latest
keel doctor
```

On macOS, `keel doctor` checks Xcode Command Line Tools and `iconutil`. On Linux, it checks window-system and font development packages. The first run or build downloads dependencies and compiles cgo code. See [the macOS `-lobjc` warning](troubleshooting.md#-lobjc-warning-when-linking) if it appears during linking.

## Create a project

```sh
keel new my-notes
cd my-notes
keel run
```

The scaffold generates:

| File | Purpose |
| --- | --- |
| `main.go` | Sets the application icon, opens the window, and mounts the UI |
| `app.go` | Defines the application view and its `Render` method |
| `keel.json` | Application name, ID, version, and icon used by the packager |
| `appicon.png` | A square 1024×1024 placeholder image without rounded corners; replace it with your own |
| `go.mod` | Depends on the current Keel version; generation runs `go get` and `go mod tidy` |
| `.gitignore`, `README.md`, `README.zh-CN.md` | Project instructions and exclusions for `dist/` and build intermediates |

Options: `-name "My Notes"` sets the display name, which otherwise derives from the directory (`my-notes` → `My Notes`). `-appid com.yourname.notes` sets the application ID, defaulting to `com.example.<directory>`. `-module` sets the Go module path. `-offline` writes files without downloading dependencies. The destination directory must be empty.

## Write your first window

Edit `app.go` in the generated project. Keep the scaffold’s `main.go` to manage the icon, window, and event loop. Replace `app.go` with:

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
        name: kit.Input("Your name").Placeholder("e.g. Alex"),
        result: "Waiting for input",
    }
    a.name.OnSubmit(func(string) { a.greet() })
    return a
}

func (a *app) greet() {
    if name := strings.TrimSpace(a.name.Value()); name != "" {
        a.result = "Hello, " + name + "!"
    }
}

func (a *app) Render(cx *el.Context) el.Element {
    return el.Div().P(24).Gap(12).Items(el.Start).Child(
        el.Text("My first window").TextSize(22).Bold(),
        a.name.Render(cx),
        kit.Button("Say hello", a.greet).Render(cx),
        el.Text(a.result).TextColor(theme.Muted),
    )
}
```

Save and run `keel run`. Enter a name, then click the button or press Enter to display a greeting. Replace the example labels with your application’s text.

Create stateful components once in `newApp` and retain them in view fields. `Render` builds an element tree from the current state. Event callbacks change the fields; the framework redraws after the callback. See [Elements and views](el.md) for layout and events, and follow the [threading rules](architecture.md#threading) for background updates.

## keel.json

```json
{
  "name": "My Notes",
  "appid": "com.example.my-notes",
  "version": "0.1.0",
  "build": 1,
  "binary": "my-notes",
  "icon": "appicon.png",
  "main": ".",
  "ios": {"minimum_version": "18.0"}
}
```

- `name` is the display name used by the `.app` bundle, menus, and Linux launcher.
- `appid` is a unique reverse-domain identifier. macOS uses it for permissions and system notifications; Linux uses it to match desktop icons. **Keep it stable after release.**
- `version` follows `major.minor.patch`; `build` distinguishes builds of the same version. These values populate macOS Info.plist and Windows executable version information.
- `binary` names the executable. `main` is the main package path relative to `keel.json`.

## Run

Desktop `keel run` watches project files, assets and local `replace`/`go.work` modules by default. Saving rebuilds the app and restarts it only after a successful compilation; build errors leave the current window running until the next edit. Consecutive saves are debounced, and development executables stay in a temporary directory without creating an application bundle. On macOS/Linux, Keel runs the windows’ `OnClose` callbacks before restarting so applications can persist state. Ordinary in-memory state resets. Closing the last window or quitting normally (Cmd+Q on macOS) stops the watcher and cancels any pending build without reopening the app. Application crashes keep the watcher alive for the next edit. Press Ctrl+C to stop, or use `keel run -watch=false` to disable watching.

`keel run` passes `appid` to Gio, and supplies the icon configured in `keel.json` when the first window opens. It honors `icon_mask` and platform overrides under `icons`, without creating an application bundle. Explicit `window.SetIcon` calls take precedence. Wayland has separate requirements; see [Application icons](app.md#app-icons). Use `keel run -- --flag` to pass arguments to your application.

Use `keel run -target ios` to build, install and launch on an iOS simulator; `-simulator <UDID>` selects the device. Check the environment with `keel doctor -target ios`. See [iOS (experimental)](ios.md).

## Package

```sh
keel build                    # Current platform
keel build -target windows    # Build a Windows package on any platform
keel build -target js         # WebAssembly
keel build -target ios        # macOS + full Xcode; simulator .app
keel build -n                 # Print commands without executing them
```

Output goes to `dist/`; change it with `-o`. Builds strip symbols, debug information, and local paths by default (`-s -w -trimpath`), reducing size by about a quarter. Crash traces still show function names. Add `-debug` when using a debugger. Syntax highlighting and network images each add about 4 MB and are optional: uncomment their imports in the generated `main.go` when needed. See [Optional features and binary size](kit.md#optional-features-and-binary-size). The first macOS or browser build downloads Gio’s packaging tool, gogio.

| Target | Output | Icon | Build host |
| --- | --- | --- | --- |
| `darwin` | `dist/My Notes.app` | Generates `icon.icns` inside the bundle | macOS with cgo and `iconutil` |
| `windows` | `dist/my-notes.exe` and `my-notes.ico` | Embeds 14 icon sizes in the executable | Any platform; no cgo required |
| `linux` | `dist/linux/`: executable, `<appid>.desktop`, icons, and `install.sh` | Uses the hicolor icon theme | Linux with Wayland/X11 development headers |
| `ios` | `dist/ios/my-notes.app`; `-device` produces a signed `.ipa` | iPhone/iPad assets | macOS with full Xcode |
| `js` | `dist/web/` | — | Any platform |

**macOS:** `-arch arm64,amd64` creates a universal bundle; the default uses the host architecture. The CLI rewrites Info.plist with the application type, name, version, and minimum macOS version (14), then signs the entire bundle. Without a signing identity, it uses an ad hoc signature for local use. For distribution, sign with `-sign "Developer ID Application: Your Name (TEAMID)"`, then notarize with `xcrun notarytool`.

**Windows:** the default is `-arch amd64`. `-arch amd64,arm64` creates one executable per architecture. Temporary `.syso` resource files are removed after compilation.

**Linux:** the build embeds `appid` as Wayland’s app_id and X11’s WM_CLASS. The desktop environment uses it to match the window to `<appid>.desktop` and its icon. `dist/linux/install.sh` installs under `~/.local` by default. Use `PREFIX=/usr/local sudo ./install.sh` for a system installation.

## Icons

`appicon.png` is the **source image**: a square 1024×1024 PNG with artwork filling the canvas. Do not add platform corners, borders, or shadows yourself. `keel build` applies each platform’s shape:

| Platform | Canvas | Artwork | Corners | Other treatment | Reference |
| --- | --- | --- | --- | --- | --- |
| macOS | 1024 | 824×824, centered with 100 padding | 185.4, continuous curvature rather than circular arcs | Shadow: 12 down, blur 28, black at 50% | Apple macOS app icon template |
| Windows | 48 grid | 42×42, centered | 2 | Transparent background, no shadow; embeds 16, 20, 24, 30, 32, 36, 40, 48, 60, 64, 72, 80, 96, and 256 sizes | Microsoft app icon design and construction guidelines |
| Linux | 128 | 104×104, centered with 12 padding | 8 | No shadow; exports eight hicolor sizes from 16 to 512 | GNOME app icon template |

Each size is rendered directly from the source image rather than downscaled from a larger output, keeping small 16 and 24 icons clear.

Preview before packaging: `keel icon` writes all platforms’ icons to `dist/icons/`: `ios.png`, `macos.png`, `windows.ico`, `windows/<size>.png`, and `linux/<size>.png`.

iOS uses full-bleed artwork with system rounding and transparency flattened over white; `icon_mask` does not affect iOS. Use `icons.ios` for separate artwork.

Configure exceptions in `keel.json`:

- For artwork with its own transparent silhouette, such as a circular logo, set `"icon_mask": "none"`. The tool scales it to each platform’s artwork area while retaining its outline and transparency.
- For finished platform icons, set `"icons": {"darwin": "icon-mac.png", "windows": "icon-win.png", "linux": "icon-linux.png"}`. These square PNGs are only resized; other platforms use normal generation.

For Windows, `keel build` embeds the icon, manifest (PerMonitorV2 DPI awareness, Common Controls 6, and long paths), and version metadata (product name, description, and version), then compiles with `go build -H=windowsgui`. It also saves `dist/<binary>.ico` for installers and shortcuts. Other `.syso` files in the main package cause an error to prevent resource conflicts.

## Develop against local Keel

To test framework changes in an application, create a scaffolded project and point `-replace` to your local checkout:

```sh
keel new my-app -replace ../keel
cd my-app
keel run
```

The scaffold adds a local replacement to the application’s `go.mod`. In an existing project, run `go mod edit -replace=github.com/dyike/keel=../keel` followed by `go mod tidy`.

## Next steps

- [Components](kit.md): find components by purpose, try live examples, and inspect their source.
- [Elements and views](el.md): layout, state, events, focus, and overlays.
- [Windows](app.md): multiple windows, shortcuts, and screenshots.
- [WebAssembly](web.md): build WebAssembly with `keel build -target js`.
