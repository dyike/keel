package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestPlistValue(t *testing.T) {
	for _, tc := range []struct {
		in   string // JSON
		want string
	}{
		{`"a<b&c"`, "<string>a&lt;b&amp;c</string>"},
		{`true`, "<true/>"},
		{`false`, "<false/>"},
		{`2`, "<integer>2</integer>"},
		{`-7`, "<integer>-7</integer>"},
		{`1.5`, "<real>1.5</real>"},
		{`[]`, "<array></array>"},
		{`["_vnc._tcp", 1]`, "<array><string>_vnc._tcp</string><integer>1</integer></array>"},
		{`{"b": true, "a": {}}`, "<dict><key>a</key><dict></dict><key>b</key><true/></dict>"},
	} {
		var v any
		if err := json.Unmarshal([]byte(tc.in), &v); err != nil {
			t.Fatal(err)
		}
		got, err := plistValue(v)
		if err != nil || got != tc.want {
			t.Errorf("plistValue(%s) = %q, %v; want %q", tc.in, got, err, tc.want)
		}
	}
	for _, in := range []string{`null`, `[1, null]`, `{"a": null}`} {
		var v any
		json.Unmarshal([]byte(in), &v)
		if _, err := plistValue(v); err == nil {
			t.Errorf("plistValue(%s) succeeded, want error", in)
		}
	}
}

func TestInfoPlistValidation(t *testing.T) {
	cfg := Config{Name: "App", AppID: "com.example.app", Version: "1.0.0", Binary: "app",
		MacOS: &MacOSConfig{InfoPlist: map[string]any{
			"CFBundleIdentifier":             "com.other",
			"NSLocalNetworkUsageDescription": "ok",
		}},
		IOS: &IOSConfig{InfoPlist: map[string]any{
			"MinimumOSVersion":   "12.0",
			"UIDeviceFamily":     []any{2.0}, // a default, may be overridden
			"NSBonjourServices":  nil,
			"":                   "x",
			"UISupportedOrients": []any{"x"},
		}},
	}
	err := cfg.validate()
	if err == nil {
		t.Fatal("validate succeeded")
	}
	msg := err.Error()
	for _, want := range []string{
		`macos.info_plist["CFBundleIdentifier"] is written by keel from appid`,
		`ios.info_plist["MinimumOSVersion"] is written by keel from ios.minimum_version`,
		`ios.info_plist["NSBonjourServices"]: null`,
		`ios.info_plist: empty key`,
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error lacks %q:\n%s", want, msg)
		}
	}
	for _, unwanted := range []string{"NSLocalNetworkUsageDescription", "UIDeviceFamily", "UISupportedOrients"} {
		if strings.Contains(msg, unwanted) {
			t.Errorf("error mentions allowed key %s:\n%s", unwanted, msg)
		}
	}
}

func TestLoadConfigInfoPlist(t *testing.T) {
	dir := t.TempDir()
	data := `{"name": "Mirage", "appid": "com.dyike.mirage", "version": "0.1.0", "binary": "mirage",
		"macos": {"info_plist": {"NSLocalNetworkUsageDescription": "LAN"}},
		"ios": {"minimum_version": "18.0", "info_plist": {"NSBonjourServices": ["_rfb._tcp"]}}}`
	os.WriteFile(filepath.Join(dir, configFile), []byte(data), 0o644)
	cfg, err := loadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.macosInfoPlist()["NSLocalNetworkUsageDescription"]; got != "LAN" {
		t.Errorf("macos info_plist = %v", cfg.macosInfoPlist())
	}
	if got := cfg.iosInfoPlist()["NSBonjourServices"]; !reflect.DeepEqual(got, []any{"_rfb._tcp"}) {
		t.Errorf("ios info_plist = %v", cfg.iosInfoPlist())
	}
}

// plistJSON converts plist XML to JSON with plutil, which also proves the
// document is a valid property list.
func plistJSON(t *testing.T, doc string) map[string]any {
	t.Helper()
	if runtime.GOOS != "darwin" {
		t.Skip("plutil is macOS only")
	}
	path := filepath.Join(t.TempDir(), "Info.plist")
	os.WriteFile(path, []byte(doc), 0o644)
	out, err := exec.Command("plutil", "-convert", "json", "-o", "-", path).CombinedOutput()
	if err != nil {
		t.Fatalf("plutil: %v: %s\n%s", err, out, doc)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMacOSInfoPlistMerge(t *testing.T) {
	cfg := &Config{Name: "Mirage", AppID: "com.dyike.mirage", Version: "0.1.0", Build: 3, Binary: "mirage",
		MacOS: &MacOSConfig{Accessory: true, InfoPlist: map[string]any{
			"NSLocalNetworkUsageDescription": "LAN <access>",
			"LSApplicationCategoryType":      "public.app-category.utilities",
		}}}
	m := plistJSON(t, plistDocument(macosInfoPlist(cfg, "Mirage")))
	for k, want := range map[string]any{
		"CFBundleIdentifier":             "com.dyike.mirage",
		"CFBundleVersion":                "3",
		"LSUIElement":                    true,
		"NSLocalNetworkUsageDescription": "LAN <access>",
		"LSApplicationCategoryType":      "public.app-category.utilities",
		"NSHighResolutionCapable":        true,
	} {
		if !reflect.DeepEqual(m[k], want) {
			t.Errorf("%s = %#v, want %#v", k, m[k], want)
		}
	}
}

func TestIOSInfoPlistMerge(t *testing.T) {
	cfg := &Config{Name: "Mirage", AppID: "com.dyike.mirage", Version: "0.1.0", Binary: "mirage",
		IOS: &IOSConfig{InfoPlist: map[string]any{
			"UIDeviceFamily":                 []any{2.0},
			"NSLocalNetworkUsageDescription": "LAN",
			"NSBonjourServices":              []any{"_rfb._tcp"},
			"UIRequiresFullScreen":           false,
		}}}
	doc := iosInfoPlist(cfg, "iPhoneOS")
	m := plistJSON(t, doc)
	for k, want := range map[string]any{
		"UIDeviceFamily":                 []any{2.0},
		"NSLocalNetworkUsageDescription": "LAN",
		"NSBonjourServices":              []any{"_rfb._tcp"},
		"UIRequiresFullScreen":           false,
		"MinimumOSVersion":               "18.0",
		"CFBundleSupportedPlatforms":     []any{"iPhoneOS"},
	} {
		if !reflect.DeepEqual(m[k], want) {
			t.Errorf("%s = %#v, want %#v", k, m[k], want)
		}
	}
	if strings.Count(doc, "<key>UIDeviceFamily</key>") != 1 {
		t.Error("UIDeviceFamily appears more than once")
	}
	scene, _ := m["UIApplicationSceneManifest"].(map[string]any)
	if !strings.Contains(doc, "KeelSceneDelegate") || scene["UIApplicationSupportsMultipleScenes"] != false {
		t.Errorf("scene manifest lost: %#v", scene)
	}
}
