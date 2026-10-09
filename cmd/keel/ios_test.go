package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func iosProject(t *testing.T) (string, *Config) {
	t.Helper()
	c, out := newCLI(t.TempDir())
	if c.main([]string{"new", "notes", "-offline", "-appid", "dev.keel.notes"}) != 0 {
		t.Fatal(out.String())
	}
	dir := filepath.Join(c.dir, "notes")
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, cfg
}

func TestIOSPlans(t *testing.T) {
	dir, _ := iosProject(t)
	c, out := newCLI(dir)
	if c.main([]string{"build", "-target", "ios", "-arch", "arm64", "-n"}) != 0 {
		t.Fatal(out.String())
	}
	for _, want := range []string{"--sdk iphonesimulator", "prepare simulator Metal shaders", "go build -trimpath -ldflags -s -w -X github.com/dyike/keel/third_party/gio/app.ID=dev.keel.notes", "KeelSceneDelegate", "--platform iphonesimulator", "codesign --force --sign -", "ios/notes.app"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("simulator plan missing %q:\n%s", want, out.String())
		}
	}
	out.Reset()
	if c.main([]string{"build", "-target", "ios", "-device", "-sign", "Apple Development: Example", "-provision", "profile.mobileprovision", "-n"}) != 0 {
		t.Fatal(out.String())
	}
	for _, want := range []string{"--sdk iphoneos", "--platform iphoneos", "security cms -D -i", "--entitlements", "embedded.mobileprovision", "ios/notes.ipa"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("device plan missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "prepare simulator Metal shaders") {
		t.Fatal("device builds must use device Metal libraries")
	}
	out.Reset()
	if c.main([]string{"run", "-target", "ios", "-simulator", "TEST-UDID", "-n", "--", "--sample", "hello world"}) != 0 {
		t.Fatal(out.String())
	}
	for _, want := range []string{"go build -trimpath=false", "simctl boot TEST-UDID", "simctl bootstatus TEST-UDID -b", "simctl install TEST-UDID", "simctl launch --terminate-running-process --console-pty TEST-UDID dev.keel.notes --sample hello world"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("run plan missing %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "-s -w") {
		t.Fatal("run should retain debug information")
	}
}

func TestIOSRejectsInvalidBuilds(t *testing.T) {
	dir, cfg := iosProject(t)
	for _, args := range [][]string{
		{"build", "-target", "ios", "-device", "-n"},
		{"build", "-target", "ios", "-device", "-sign", "-", "-provision", "x", "-n"},
		{"build", "-target", "ios", "-sign", "Developer ID: X", "-n"},
		{"build", "-target", "ios", "-arch", "arm64,amd64", "-n"},
		{"build", "-target", "ios", "-device", "-arch", "amd64", "-sign", "X", "-provision", "x", "-n"},
		{"build", "-target", "js", "-device", "-n"},
		{"run", "-simulator", "ID", "-n"},
	} {
		c, out := newCLI(dir)
		if c.main(args) == 0 {
			t.Errorf("accepted %v: %s", args, out.String())
		}
	}
	for _, minimum := range []string{"12.9", "18", "18.0-simulator", "oops"} {
		cfg.IOS.MinimumVersion = minimum
		if cfg.validate() == nil {
			t.Errorf("accepted minimum version %q", minimum)
		}
	}
	for _, minimum := range []string{"13.0", "18.0", "18.4.1"} {
		cfg.IOS.MinimumVersion = minimum
		if err := cfg.validate(); err != nil {
			t.Errorf("rejected minimum version %q: %v", minimum, err)
		}
	}
	cfg.IOS = nil
	if cfg.iosMinimumVersion() != "18.0" || cfg.validate() != nil {
		t.Fatal("old project configurations must still work")
	}
}

func TestIOSBundleMetadataAndProfile(t *testing.T) {
	_, cfg := iosProject(t)
	cfg.Name, cfg.Build = "Notes & <More>", 7
	for _, platform := range []string{"iPhoneSimulator", "iPhoneOS"} {
		data := iosInfoPlist(cfg, platform)
		decoder := xml.NewDecoder(strings.NewReader(data))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("invalid XML: %v", err)
			}
		}
		for _, want := range []string{"Notes &amp; &lt;More&gt;", "dev.keel.notes", "<integer>2</integer>", "KeelSceneDelegate", "<string>7</string>", platform} {
			if !strings.Contains(data, want) {
				t.Errorf("missing %q in plist", want)
			}
		}
	}
	for _, profile := range []string{"TEAM.dev.keel.notes", "TEAM.dev.keel.*", "TEAM.*"} {
		if value, err := iosProfileAppID(profile, cfg.AppID); err != nil || value != "TEAM.dev.keel.notes" {
			t.Errorf("profile %q: %q %v", profile, value, err)
		}
	}
	for _, profile := range []string{"TEAM.com.other.*", "broken", ".*", "TEAM.dev.keel.other"} {
		if _, err := iosProfileAppID(profile, cfg.AppID); err == nil {
			t.Errorf("accepted nonmatching profile %q", profile)
		}
	}
}

func TestIOSSimulatorSelection(t *testing.T) {
	devices := map[string][]iosSimulator{
		"com.apple.CoreSimulator.SimRuntime.iOS-18-5":     {{Name: "iPhone 16", UDID: "OLD", State: "Shutdown", IsAvailable: true}},
		"com.apple.CoreSimulator.SimRuntime.iOS-27-0":     {{Name: "iPhone 18", UDID: "NEW", State: "Shutdown", IsAvailable: true}, {Name: "iPhone Missing", UDID: "MISSING", IsAvailable: false}},
		"com.apple.CoreSimulator.SimRuntime.watchOS-27-0": {{Name: "Watch", UDID: "WATCH", State: "Booted", IsAvailable: true}},
	}
	inventory := func() []byte {
		data, _ := json.Marshal(map[string]any{"devices": devices})
		return data
	}
	if device, err := selectIOSSimulator(inventory(), ""); err != nil || device.UDID != "NEW" {
		t.Fatal(device, err)
	}
	if device, err := selectIOSSimulator(inventory(), "old"); err != nil || device.UDID != "OLD" {
		t.Fatal(device, err)
	}
	if _, err := selectIOSSimulator(inventory(), "MISSING"); err == nil {
		t.Fatal("unavailable simulator accepted")
	}
	devices["com.apple.CoreSimulator.SimRuntime.iOS-18-5"][0].State = "Booted"
	if device, err := selectIOSSimulator(inventory(), ""); err != nil || device.UDID != "OLD" {
		t.Fatal(device, err)
	}
	devices["com.apple.CoreSimulator.SimRuntime.iOS-27-0"][0].State = "Booted"
	if _, err := selectIOSSimulator(inventory(), ""); err == nil {
		t.Fatal("ambiguous booted simulators accepted")
	}
	if _, err := selectIOSSimulator([]byte(`{"devices":{}}`), ""); err == nil {
		t.Fatal("empty inventory accepted")
	}
}

func TestIOSSimulatorAppLocations(t *testing.T) {
	for _, relative := range []string{"Developer/Applications/Simulator.app", "Applications/Simulator.app", "Applications/DeviceHub.app"} {
		contents := t.TempDir()
		want := filepath.Join(contents, relative)
		if err := os.MkdirAll(want, 0o755); err != nil {
			t.Fatal(err)
		}
		got, err := iosSimulatorApp(filepath.Join(contents, "Developer"))
		if err != nil || got != want {
			t.Fatalf("%s: %q, %v", relative, got, err)
		}
	}
	if _, err := iosSimulatorApp(t.TempDir()); err == nil {
		t.Fatal("missing Simulator.app accepted")
	}
}

func TestIOSArchivePreservesExecutable(t *testing.T) {
	work := t.TempDir()
	app := filepath.Join(work, "Payload", "notes.app")
	os.MkdirAll(app, 0o755)
	os.WriteFile(filepath.Join(app, "notes"), []byte("binary"), 0o755)
	os.WriteFile(filepath.Join(app, "Info.plist"), []byte("metadata"), 0o644)
	output := filepath.Join(t.TempDir(), "notes.ipa")
	if err := zipIOSPayload(work, output); err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(output)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	found := false
	for _, file := range archive.File {
		if file.Name == "Payload/notes.app/notes" {
			found = true
			if file.Mode().Perm()&0o111 == 0 {
				t.Fatal("archive lost executable permissions")
			}
		}
	}
	if !found {
		t.Fatal("archive has no Payload executable")
	}
}

// Real packaging covers the generated project, shader adapter and actool,
// including paths with spaces. It neither boots nor installs a simulator.
func TestNewIOSProjectBuilds(t *testing.T) {
	if testing.Short() || runtime.GOOS != "darwin" || os.Getenv("KEEL_IOS") != "1" {
		t.Skip("set KEEL_IOS=1 on a Mac with full Xcode to build a simulator app")
	}
	repo, _ := filepath.Abs("../..")
	root := t.TempDir()
	c, out := newCLI(root)
	if c.main([]string{"new", "Mobile Notes", "-replace", repo}) != 0 {
		t.Fatal(out.String())
	}
	c, out = newCLI(filepath.Join(root, "Mobile Notes"))
	if c.main([]string{"build", "-target", "ios"}) != 0 {
		t.Fatal(out.String())
	}
	app := filepath.Join(c.dir, "dist", "ios", "mobile-notes.app")
	if data, err := exec.Command("codesign", "--verify", "--strict", app).CombinedOutput(); err != nil {
		t.Fatalf("bundle signature: %v\n%s", err, data)
	}
	if data, err := os.ReadFile(filepath.Join(app, "Info.plist")); err != nil || !bytes.Contains(data, []byte("KeelSceneDelegate")) {
		t.Fatalf("scene manifest missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(app, "Assets.car")); err != nil {
		t.Fatal("icon assets missing:", err)
	}
}
