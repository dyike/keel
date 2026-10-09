# Android

English | [简体中文](android.zh-CN.md)

`keel build -target android` packages your Go UI as a development-signed APK. `keel run -target android` builds for the device ABI, installs the APK and launches the app on a running emulator or a connected device with USB debugging enabled.

## Environment and commands

Install Go 1.26+, Android SDK platforms, build-tools 35+, the NDK and a JDK through Android Studio's SDK Manager or `sdkmanager`. Running also requires `adb` from platform-tools. Set `ANDROID_HOME` to the SDK root (`ANDROID_SDK_ROOT` is also accepted), and set `JAVA_HOME` or put Java tools on `PATH`.

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

Builds use `gioui.org/cmd/gogio@v0.10.0`. NDK lookup prefers the newest side-by-side version under the SDK, then `ndk-bundle`, then `ANDROID_NDK_ROOT`. See [Gio's Android installation guide](https://gioui.org/doc/install/android); compatibility is determined by the pinned packager and actual builds.

The output is `dist/android/<binary>.apk`. Builds include arm64 and amd64 by default; use `-arch arm64` for a smaller APK. `arm`, `386` and comma-separated architectures are also accepted. `-o` changes the output root; `-n` prints commands.

`run` selects the sole ready device. With multiple devices, specify its serial; offline or unauthorized devices produce an error. Start your emulator in Android Studio's Device Manager first; the CLI does not create emulators.

```sh
keel run -target android -serial emulator-5554
adb -s emulator-5554 logcat
```

The command returns after launching. Read logs with `adb logcat`. Android does not pass `run -- ...` arguments into Go; the CLI rejects them. Gio's Android packager always strips native symbols, so `-debug` cannot restore full Android native-library debug information.

## Configuration, icons and signing

```json
"android": {
  "minimum_sdk": 23,
  "target_sdk": 35
}
```

These defaults also apply when `android` is omitted. `minimum_sdk` must be at least 23. `target_sdk` must be at least 31 and no lower than the minimum; install a platform SDK covering it. `appid` must be a Java package name without hyphens. New projects remove hyphens and underscores from the directory name when generating a default ID.

Name, ID, version and build number come from `keel.json`; `build` becomes Android's versionCode. Icons use full-bleed artwork; the packager generates density variants and an adaptive icon. `icon_mask` does not affect Android. Override the artwork with `icons.android`; `keel icon` writes `dist/icons/android.png`.

The CLI creates and reuses `dist/android/debug.keystore` with the public development password `android`. Keep this key to replace installed builds. Deleting `dist` or changing `-o` creates a new key; preserve the old key or uninstall the old app first (uninstalling removes app data). The CLI does not currently provide Android release signing, AAB packaging or store uploads.

## Verification

```sh
go test ./cmd/keel
KEEL_ANDROID=1 go test ./cmd/keel -run TestNewAndroidProjectBuilds -count=1
```

The second check builds arm64 and amd64 APKs from a generated project in a path containing spaces, checking the manifest, DEX and native libraries. See [Mobile support](mobile.md) for system API and interaction limitations.

Verified on Apple Silicon with platform SDK 35, build-tools 35.0.0, NDK 27.0.12077973 and JDK 21: arm64/amd64 APK packaging, plus replacement installation, first frame, Chinese text, button taps, Latin-character soft-keyboard input, back to dismiss the keyboard and background/foreground recovery on an API 35 arm64 emulator. D8 from build-tools 34 failed with JDK 21; upgrade to 35 or newer. Physical devices, composition input and comprehensive narrow-screen component interactions still need validation.

Android loads the system `NotoSansCJK-Regular.ttc` or `NotoSansSC-Regular.otf` directly, avoiding missing Chinese glyphs when default font scanning has no usable cache directory. If an OEM provides neither font, load a bundled CJK font with `theme.LoadFonts` before starting the UI.
