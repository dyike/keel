package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAndroidPlans(t *testing.T) {
	dir, cfg := iosProject(t)
	cfg.Name = "Mobile Notes"
	cfg.Android = &AndroidConfig{MinimumSDK: 26, TargetSDK: 35}
	cfg.Icons = map[string]string{"android": cfg.Icon}
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	c, out := newCLI(dir)
	if c.main([]string{"build", "-target", "android", "-arch", "arm64, amd64", "-n"}) != 0 {
		t.Fatal(out.String())
	}
	for _, want := range []string{"-target android -arch arm64,amd64", "-minsdk 26 -targetsdk 35", "-appid dev.keel.notes -name Mobile Notes -version 0.1.0.1", "-ldflags -s -w -X github.com/dyike/keel/third_party/gio/app.ID=dev.keel.notes", "android/notes.apk", "-signkey", "-genkeypair"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q:\n%s", want, out.String())
		}
	}
	// Rebuilding must reuse the key rather than changing the app's identity.
	if err := os.MkdirAll(filepath.Join(dir, "dist", "android"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dist", "android", "debug.keystore"), []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if c.main([]string{"run", "-target", "android", "-serial", "emulator-5554", "-n"}) != 0 {
		t.Fatal(out.String())
	}
	for _, want := range []string{"-s emulator-5554 shell getprop ro.product.cpu.abi", "-s emulator-5554 install -r", "shell am start -W -S -n dev.keel.notes/org.gioui.GioActivity"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "-genkeypair") {
		t.Fatal("existing keystore replaced")
	}
}

func TestAndroidValidation(t *testing.T) {
	dir, cfg := iosProject(t)
	for _, args := range [][]string{
		{"build", "-target", "android", "-arch", "wasm", "-n"},
		{"build", "-target", "android", "-arch", "arm64,arm64", "-n"},
		{"build", "-target", "android", "-sign", "X", "-n"},
		{"run", "-target", "android", "-simulator", "ID", "-n"},
		{"run", "-serial", "ID", "-n"},
		{"run", "-target", "ios", "-serial", "ID", "-n"},
		{"run", "-target", "android", "-n", "--", "--flag"},
	} {
		c, out := newCLI(dir)
		if c.main(args) == 0 {
			t.Errorf("accepted %v: %s", args, out.String())
		}
	}
	cfg.AppID = "com.example.my-app"
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	c, out := newCLI(dir)
	if c.main([]string{"build", "-target", "android", "-n"}) == 0 {
		t.Fatal("accepted hyphenated package ID", out.String())
	}
	for _, android := range []*AndroidConfig{{MinimumSDK: 22}, {TargetSDK: 30}, {MinimumSDK: 36, TargetSDK: 35}, {MinimumSDK: -1}} {
		cfg.Android = android
		if cfg.validate() == nil {
			t.Errorf("accepted %+v", android)
		}
	}
}

func TestAndroidDeviceSelection(t *testing.T) {
	inventory := "List of devices attached\nphone\tdevice product:phone\noffline\toffline\nlocked\tunauthorized\n"
	if got, err := selectAndroidDevice(inventory, ""); err != nil || got != "phone" {
		t.Fatal(got, err)
	}
	for _, serial := range []string{"offline", "locked", "missing"} {
		if _, err := selectAndroidDevice(inventory, serial); err == nil {
			t.Fatal("accepted", serial)
		}
	}
	inventory += "emulator-5554\tdevice\n"
	if _, err := selectAndroidDevice(inventory, ""); err == nil {
		t.Fatal("ambiguous devices accepted")
	}
	if got, err := selectAndroidDevice(inventory, "emulator-5554"); err != nil || got != "emulator-5554" {
		t.Fatal(got, err)
	}
	if _, err := selectAndroidDevice("List of devices attached\n", ""); err == nil {
		t.Fatal("empty inventory accepted")
	}
	for abi, arch := range map[string]string{"arm64-v8a": "arm64", "armeabi-v7a": "arm", "x86": "386", "x86_64": "amd64"} {
		if got, err := androidABIArchitecture(abi + "\n"); err != nil || got != arch {
			t.Fatal(got, err)
		}
	}
	if _, err := androidABIArchitecture("unknown"); err == nil {
		t.Fatal("unknown ABI accepted")
	}
}

// Opt in to the complete Android SDK/NDK/Java packaging check.
func TestNewAndroidProjectBuilds(t *testing.T) {
	if testing.Short() || os.Getenv("KEEL_ANDROID") != "1" {
		t.Skip("set KEEL_ANDROID=1 with Android SDK, NDK and Java installed")
	}
	repo, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	c, out := newCLI(t.TempDir())
	if c.main([]string{"new", "mobile notes", "-appid", "dev.keel.notes", "-replace", repo}) != 0 {
		t.Fatal(out.String())
	}
	c.dir = filepath.Join(c.dir, "mobile notes")
	for _, arch := range []string{"arm64", "amd64"} {
		out.Reset()
		if c.main([]string{"build", "-target", "android", "-arch", arch}) != 0 {
			t.Fatal(out.String())
		}
		apk, err := zip.OpenReader(filepath.Join(c.dir, "dist", "android", "mobile-notes.apk"))
		if err != nil {
			t.Fatal(err)
		}
		files := map[string]bool{}
		for _, file := range apk.File {
			files[file.Name] = true
		}
		apk.Close()
		abi := map[string]string{"arm64": "arm64-v8a", "amd64": "x86_64"}[arch]
		for _, name := range []string{"AndroidManifest.xml", "classes.dex", "lib/" + abi + "/libgio.so"} {
			if !files[name] {
				t.Errorf("APK missing %s", name)
			}
		}
	}
}

func TestAndroidScaffoldDefaults(t *testing.T) {
	c, out := newCLI(t.TempDir())
	if c.main([]string{"new", "my-notes_app", "-offline"}) != 0 {
		t.Fatal(out.String())
	}
	cfg, err := loadConfig(filepath.Join(c.dir, "my-notes_app"))
	if err != nil {
		t.Fatal(err)
	}
	if !androidAppIDPattern.MatchString(cfg.AppID) || cfg.androidMinimumSDK() != 23 || cfg.androidTargetSDK() != 35 {
		t.Fatalf("invalid Android scaffold defaults: %+v", cfg)
	}
	cfg.Android = nil
	if cfg.androidMinimumSDK() != 23 || cfg.androidTargetSDK() != 35 {
		t.Fatal("old projects lost mobile defaults")
	}
}

func TestAndroidSDKDiscovery(t *testing.T) {
	t.Setenv("ANDROID_HOME", "")
	t.Setenv("ANDROID_SDK_ROOT", "sdk-fallback")
	if androidSDK() != "sdk-fallback" {
		t.Fatal("SDK_ROOT fallback ignored")
	}
	t.Setenv("ANDROID_HOME", "sdk-primary")
	if androidSDK() != "sdk-primary" {
		t.Fatal("ANDROID_HOME should take priority")
	}
	root := t.TempDir()
	for _, version := range []string{"9.0.0", "27.0.12077973", "26.3.11579264"} {
		if err := os.MkdirAll(filepath.Join(root, "ndk", version), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("ANDROID_NDK_ROOT", "ndk-override")
	if got := androidNDK(root); got != filepath.Join(root, "ndk", "27.0.12077973") {
		t.Fatal("NDK selection differs from gogio:", got)
	}
	if got := androidNDK(t.TempDir()); got != "ndk-override" {
		t.Fatal("NDK_ROOT fallback ignored:", got)
	}
}
