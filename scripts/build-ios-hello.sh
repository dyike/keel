#!/bin/sh
# Build the hello example for the host Mac's iOS simulator architecture.
set -eu

cd "$(dirname "$0")/.."
output=${1:-/tmp/keel-ios-hello.app}
case "$output" in
  *.app) ;;
  *) echo "output must end in .app" >&2; exit 1 ;;
esac
case "$output" in
  /*) ;;
  *) output="$(pwd)/$output" ;;
esac

if [ "$(uname -s)" != Darwin ]; then
  echo "build on a Mac with Xcode installed" >&2
  exit 1
fi

case "$(uname -m)" in
  arm64) goarch=arm64; clangarch=arm64 ;;
  x86_64) goarch=amd64; clangarch=x86_64 ;;
  *) echo "build on a Mac with Xcode installed" >&2; exit 1 ;;
esac

sdk=$(xcrun --sdk iphonesimulator --show-sdk-path)
compiler=$(xcrun --sdk iphonesimulator --find clang)
flags="-target $clangarch-apple-ios18.0-simulator -isysroot $sdk -fobjc-arc"
workdir=$(mktemp -d "${TMPDIR:-/tmp}/keel-ios-simulator.XXXXXX")
trap 'rm -rf "$workdir"' 0
trap 'exit 1' HUP INT TERM
shaderdir=$(go list -m -f '{{.Dir}}' gioui.org/shader)
# shader v1.0.9 identifies the simulator by amd64 alone. Select its existing
# simulator Metal libraries for arm64 too, only in this build's module copy.
# Leave the downloaded module and device builds unchanged.
cp -R "$shaderdir" "$workdir/shader"
chmod -R u+w "$workdir/shader"
cp go.mod "$workdir/go.mod"
cp go.sum "$workdir/go.sum"
go mod edit -modfile "$workdir/go.mod" -replace "gioui.org/shader=$workdir/shader"
python3 - "$workdir/shader" <<'PY'
import pathlib
import sys

source = pathlib.Path(sys.argv[1])
for package in ("gio", "piet"):
    path = source / package / "shaders.go"
    text = path.read_text()
    condition = 'runtime.GOARCH == "amd64"'
    if condition not in text:
        raise SystemExit("shader simulator selection changed; review the build script")
    path.write_text(text.replace(condition, '(' + condition + ' || runtime.GOARCH == "arm64")'))
PY
mkdir -p "$output"
GOOS=ios GOARCH="$goarch" CGO_ENABLED=1 CC="$compiler" \
  CGO_CFLAGS="$flags" CGO_LDFLAGS="$flags" \
  go build -modfile "$workdir/go.mod" -ldflags '-X gioui.org/app.ID=dev.keel.hello' \
  -o "$output/KeelHello" ./examples/hello

cat > "$output/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key><string>KeelHello</string>
  <key>CFBundleIdentifier</key><string>dev.keel.hello</string>
  <key>CFBundleName</key><string>Keel Hello</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>0.1.0</string>
  <key>CFBundleVersion</key><string>1</string>
  <key>CFBundleSupportedPlatforms</key><array><string>iPhoneSimulator</string></array>
  <key>MinimumOSVersion</key><string>18.0</string>
  <key>UIDeviceFamily</key><array><integer>1</integer><integer>2</integer></array>
  <key>UIApplicationSceneManifest</key>
  <dict>
    <key>UIApplicationSupportsMultipleScenes</key><false/>
    <key>UISceneConfigurations</key>
    <dict>
      <key>UIWindowSceneSessionRoleApplication</key>
      <array><dict>
        <key>UISceneConfigurationName</key><string>Keel</string>
        <key>UISceneClassName</key><string>UIWindowScene</string>
        <key>UISceneDelegateClassName</key><string>KeelSceneDelegate</string>
      </dict></array>
    </dict>
  </dict>
  <key>UILaunchScreen</key><dict/>
  <key>UISupportedInterfaceOrientations</key>
  <array><string>UIInterfaceOrientationPortrait</string></array>
</dict>
</plist>
PLIST
plutil -lint "$output/Info.plist"
codesign --force --sign - "$output"
printf 'Built %s\n' "$output"
