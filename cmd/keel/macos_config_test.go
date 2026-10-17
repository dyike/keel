package main

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMacOSAccessoryBundleMetadata(t *testing.T) {
	for _, tc := range []struct {
		name, setting string
		accessory     bool
	}{
		{"default", "", false},
		{"regular", `,"macos":{"accessory":false}`, false},
		{"accessory", `,"macos":{"accessory":true}`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var cfg Config
			if err := json.Unmarshal([]byte(`{"name":"Example","appid":"com.example.app","version":"1.0.0","binary":"example"`+tc.setting+`}`), &cfg); err != nil {
				t.Fatal(err)
			}
			bundle := t.TempDir()
			dir := filepath.Join(bundle, "Contents", "MacOS")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "example"), nil, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := (&cli{}).writeInfoPlist(bundle, &cfg); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(bundle, "Contents", "Info.plist"))
			if err != nil {
				t.Fatal(err)
			}
			var document struct{ XMLName xml.Name }
			if err := xml.Unmarshal(data, &document); err != nil {
				t.Fatal(err)
			}
			if got := strings.Contains(string(data), "<key>LSUIElement</key><true/>"); got != tc.accessory {
				t.Fatalf("LSUIElement=%v, want %v\n%s", got, tc.accessory, data)
			}
		})
	}
}
