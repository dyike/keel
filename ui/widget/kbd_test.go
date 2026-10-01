package widget

import (
	"testing"
)

func TestShortcutLabels(t *testing.T) {
	for _, tt := range []struct{ input, platform, want string }{
		{"mod+shift+p", "darwin", "⇧⌘P"}, {"mod+shift+p", "windows", "Ctrl+Shift+P"},
		{"mod+enter", "linux", "Ctrl+Enter"}, {"cmd+k", "windows", "Win+K"},
		{"alt+backspace", "darwin", "⌥⌫"}, {"esc", "darwin", "Esc"},
		{"ctrl+up", "linux", "Ctrl+↑"}, {"f12", "windows", "F12"},
		{"ctrl+", "darwin", "ctrl+"}, {"ctrl+a+b", "linux", "ctrl+a+b"},
	} {
		if got := shortcutLabel(tt.input, tt.platform); got != tt.want {
			t.Errorf("%s/%s: %q != %q", tt.input, tt.platform, got, tt.want)
		}
	}
}
