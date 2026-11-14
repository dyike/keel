//go:build linux || freebsd || openbsd

package fontscan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFontDirectoriesNilLoggerWithInvalidConfig(t *testing.T) {
	config := filepath.Join(t.TempDir(), "fonts.conf")
	if err := os.WriteFile(config, []byte("<fontconfig><invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FONTCONFIG_FILE", config)
	// Fontconfig errors must use the default logger, as the other public fontscan
	// entry points do, instead of dereferencing a nil Logger. Directories remain
	// host-dependent, so either a fallback result or an error is valid here.
	_, _ = DefaultFontDirectories(nil)
}
