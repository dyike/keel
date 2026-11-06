//go:build keel_legacy_svg

package oksvg

import (
	"strings"
	"testing"
)

func TestLegacyCharset(t *testing.T) {
	for _, tc := range []struct{ label, encoded, decoded string }{
		{"Big5", "\xa4\xa4\xa4\xe5", "中文"},
		{"GBK", "\xd6\xd0\xce\xc4", "中文"},
		{"Shift_JIS", "\x93\xfa\x96\x7b", "日本"},
		{"windows-1252", "caf\xe9", "café"},
	} {
		icon, err := ReadIconStream(strings.NewReader(`<?xml version="1.0" encoding="` + tc.label + `"?><svg><title>` + tc.encoded + `</title></svg>`))
		if err != nil {
			t.Fatal(tc.label, err)
		}
		if icon.Titles[0] != tc.decoded {
			t.Fatalf("%s: %q", tc.label, icon.Titles)
		}
	}
	if _, err := ReadIconStream(strings.NewReader(`<?xml version="1.0" encoding="unknown"?><svg/>`)); err == nil {
		t.Fatal("accepted unknown charset")
	}
}
