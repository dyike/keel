package core

import "testing"

func TestOpenURLRejectsInvalidDestinations(t *testing.T) {
	for _, raw := range []string{"", "relative/path", "https:", "mailto:", "file:///tmp/a", "javascript:alert(1)", "https://bad\nvalue"} {
		if err := OpenURL(raw); err == nil {
			t.Fatalf("accepted invalid URL %q", raw)
		}
	}
}
