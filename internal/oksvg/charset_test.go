//go:build !keel_legacy_svg

package oksvg

import (
	"strings"
	"testing"
)

func TestDefaultCharset(t *testing.T) {
	for _, label := range []string{"UTF-8", "utf-8", "utf8"} {
		icon, err := ReadIconStream(strings.NewReader(`<?xml version="1.0" encoding="` + label + `"?><svg><title>中文</title></svg>`))
		if err != nil {
			t.Fatal(err)
		}
		if icon.Titles[0] != "中文" {
			t.Fatal(icon.Titles)
		}
	}
	for _, label := range []string{"Big5", "windows-1252", "unknown"} {
		_, err := ReadIconStream(strings.NewReader(`<?xml version="1.0" encoding="` + label + `"?><svg/>`))
		if err == nil || !strings.Contains(err.Error(), "keel_legacy_svg") {
			t.Fatalf("%s: %v", label, err)
		}
	}
}
