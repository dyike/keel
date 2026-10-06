package deps

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// The interface builds for the browser. Go's os/user has no js/wasm
// implementation unless osusergo is set, and Gio's font scanning imports it,
// so web builds pass -tags osusergo (docs/web.md).
func TestUIBuildsForWebAssembly(t *testing.T) {
	if testing.Short() {
		t.Skip("cross-compiles the interface")
	}
	cmd := exec.Command("go", "build", "-tags", "osusergo", "-o", os.DevNull, "../../ui/...", "../../examples/hello", "../../examples/components")
	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("GOOS=js GOARCH=wasm go build: %v\n%s", err, strings.TrimSpace(string(out)))
	}
}

// Compile both layers with cgo enabled: iOS also matches the darwin build
// tag, but must never select the AppKit bindings used by macOS.
func TestBuildsForIOSSimulator(t *testing.T) {
	if testing.Short() || runtime.GOOS != "darwin" {
		t.Skip("requires a Mac with the Xcode iOS simulator SDK")
	}
	sdk, err := exec.Command("xcrun", "--sdk", "iphonesimulator", "--show-sdk-path").Output()
	if err != nil {
		t.Skipf("iOS simulator SDK unavailable: %v", err)
	}
	compiler, err := exec.Command("xcrun", "--sdk", "iphonesimulator", "--find", "clang").Output()
	if err != nil {
		t.Fatal(err)
	}
	arch := runtime.GOARCH
	clangArch := arch
	if arch == "amd64" {
		clangArch = "x86_64"
	}
	flags := "-target " + clangArch + "-apple-ios18.0-simulator -isysroot " + strings.TrimSpace(string(sdk)) + " -fobjc-arc"
	cmd := exec.Command("go", "build", "-o", os.DevNull, "../../ui/...", "../../native/...", "../../examples/hello")
	cmd.Env = append(os.Environ(), "GOOS=ios", "GOARCH="+arch, "CGO_ENABLED=1",
		"CC="+strings.TrimSpace(string(compiler)), "CGO_CFLAGS="+flags, "CGO_LDFLAGS="+flags)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("iOS simulator go build: %v\n%s", err, strings.TrimSpace(string(out)))
	}
}
