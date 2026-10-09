# cmd/keel

English | [简体中文](README.zh-CN.md)

Keel's scaffolding: `keel new` creates a new project, `keel run` runs it, `keel build` packages it by platform (with icon, name, version), and `keel doctor` checks the environment. For usage, see [Quick Start](../../docs/getting-started.md).

On macOS, `keel run` builds a temporary signed app bundle with the configured name and icon. Both watching and `-watch=false` launch its executable directly, preserving the working directory, arguments, stdio and signal handling. The bundle and icon are removed after the application exits.

On Windows, `keel run` and `keel build` share the native resources: PerMonitorV2 DPI awareness, Common Controls 6, long-path support, icons and version metadata. Both watching and `-watch=false` launch a GUI executable. Temporary `.syso` objects are removed after compilation; existing objects cause a conflict error and remain untouched. Legacy projects without a default icon still receive the manifest and version resources.

```sh
go install github.com/dyike/keel/cmd/keel@latest
```

- **Dependencies**: `internal/svgicon` (draw the SVG of the original image into PNG), `internal/appicon` (the icon shape of each platform, `ui/window` also uses it when running), `golang.org/x/image` (scaling, vector rasterization), `github.com/tc-hib/winres` (Windows resources: icons, manifests, version information). Keel's interface package is not referenced; macOS, Android and browser packaging call Gio's gogio (fixed in v0.10.0), and macOS signature calls the system's `codesign`.
- **Test**: `go test ./cmd/keel` generates the project, checks the packaging commands for each platform (`-n`), and compiles the generated project with Keel in this repository (`-short` skips).

| File | Responsibility |
| --- | --- |
| `main.go` | Subcommand distribution, running external commands |
| `config.go` | `keel.json` Reading, writing and verification |
| `new.go` | New project, template in `template/` |
| `run.go`, `run_bundle_*.go` | `keel run` flags, development icons and macOS app bundles |
| `run_watch.go`, `run_process_*.go` | File watching, rebuilds and process lifecycle |
| `build.go` | Packaging for each platform, Windows resources, Info.plist, Linux desktop files |
| `icons.go` | Generate icons, `keel icon`, .ico by platform and size |
| `doctor.go` | Environmental Check |

iOS builds invoke Go and Xcode SDKs, `actool` and signing tools directly; the runner uses `simctl`, without Python. `ios.go` handles packaging, Metal compatibility, icons and signing; `ios_run.go` selects, installs and launches simulators. Full packaging check: `KEEL_IOS=1 go test ./cmd/keel -run TestNewIOSProjectBuilds -count=1`. See [iOS](../../docs/ios.md).

Android builds use the pinned Gio packager. `android.go` handles development signing, APK packaging, adb device selection and launch; `android_doctor.go` checks the SDK, NDK, Java and adb. Full packaging check: `KEEL_ANDROID=1 go test ./cmd/keel -run TestNewAndroidProjectBuilds -count=1`. See [Mobile support](../../docs/mobile.md) and [Android](../../docs/android.md).
