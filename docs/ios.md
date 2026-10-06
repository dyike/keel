# iOS (experimental)

English | [简体中文](ios.zh-CN.md)

The scaffold builds iOS apps from the same Go UI code: `keel build -target ios` packages a simulator `.app`, and `keel run -target ios` builds, installs and launches it. Device builds produce a signed `.ipa` when supplied with an identity and provisioning profile. Apple Silicon with an iOS 27 simulator has been verified; device signing and mobile interactions still need validation.

## Create, build and run

Requires macOS, full Xcode (Command Line Tools alone are insufficient), Go 1.26+, and an installed iOS simulator runtime. The default deployment target is iOS 18.0.

```sh
keel new my-app
cd my-app
keel doctor -target ios
keel build -target ios
keel run -target ios
```

To develop against a local Keel checkout, build the CLI there and point the generated project at that checkout:

```sh
go build -o /tmp/keel ./cmd/keel
/tmp/keel new /tmp/my-app -replace "$PWD"
cd /tmp/my-app
/tmp/keel run -target ios
```

Output is `dist/ios/<binary>.app`, with the name, bundle ID, version, build number and icon from `keel.json`, signed ad hoc. Use `-o` to change the output root, `-debug` to retain debug information, or `-n` to print the build plan.

`run` uses the sole booted iOS simulator, or prefers an iPhone on the newest runtime when none is booted. Multiple booted devices require an explicit UDID. It boots the simulator, waits for readiness, opens Simulator or DeviceHub, installs the app and attaches application logs to the terminal.

```sh
xcrun simctl list devices available
keel run -target ios -simulator <UDID>
keel run -target ios -simulator <UDID> -- --my-app-flag
```

Simulator builds default to the Mac architecture. Explicit `-arch arm64` or `-arch amd64` selects one architecture per build. Devices always use arm64.

## Configuration and icons

New projects include the configuration below. Existing projects may omit `ios` and still default to 18.0. Minimum versions of 13.0 or newer are accepted; lower deployment targets require your own validation.

```json
"ios": {
  "minimum_version": "18.0"
}
```

iOS uses square, full-bleed artwork. The system applies rounded corners, and transparent pixels are composited over white; `icon_mask` does not affect iOS. Override the artwork with `"icons": {"ios": "icon-ios.png"}`. `keel icon` writes a `dist/icons/ios.png` preview; packaging compiles iPhone/iPad assets using Xcode's `actool`.

## Device packaging

Provide an Apple signing identity with its private key in Keychain and a provisioning profile matching `appid`:

```sh
keel build -target ios -device \
  -sign "Apple Development: Your Name (TEAMID)" \
  -provision ./app.mobileprovision
```

Output is `dist/ios/<binary>.ipa`. The CLI validates the profile's application identifier, extracts entitlements, resolves wildcard identifiers and keychain groups, embeds the profile and signs the app. Configure the profile's allowed devices, expiry, certificates and distribution method yourself. Device installation and App Store upload have not been verified; there is no automatic upload flow.

## Build and lifecycle adapters

The pinned Gio packager treats arm64 as a device architecture, so this target invokes Go and the appropriate Xcode SDK directly. `gioui.org/shader v1.0.9` identifies simulators by amd64 alone. The CLI copies that module into a temporary directory, selects its existing simulator Metal libraries on arm64, and builds using a temporary `go.mod`. Project dependency files and the module cache stay unchanged; temporary copies are cleaned afterward. Revalidate these adapters when upgrading Gio or shader. Simulator builds use the project's `go.mod` and ignore `go.work`; put workspace dependencies in module-level `replace` directives.

`ui/window/scene_ios.*` adapts Gio to a single UIScene for newer SDKs and iOS 27. The CLI configures `KeelSceneDelegate` in `Info.plist` and defers Gio window creation until the scene is connected and the main event loop starts. This depends on Gio v0.10's `_gioAppDelegate` and `GioViewController` names; missing classes or methods produce an incompatibility error. Only one scene is supported.

## Verification scope

Verified: generated project packaging, icon resources, bundle metadata, ad hoc signature, automatic installation and launch, and the first Metal frame. Run:

```sh
go test ./cmd/keel
KEEL_IOS=1 go test ./cmd/keel -run TestNewIOSProjectBuilds -count=1
go test ./internal/deps -run TestBuildsForIOSSimulator -count=1
```

The second check builds a generated project on a Mac, including paths with spaces. The third cross-compiles `ui/*`, `native/*` and hello when SDKs are available. UI validation still needs button and checkbox taps, software keyboard, Chinese input, narrow-screen scrolling and background/foreground recovery. Intel Macs, iPads and physical devices remain unverified.

The iOS `native/*` backend is not implemented: permissions, screen capture, synthetic input, global shortcuts, notifications and rich clipboard operations return `native.ErrUnsupported`. Gio text-editor copy and paste requires separate validation.
