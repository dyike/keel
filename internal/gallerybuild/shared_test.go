package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedBuildMissingToolsDoesNotPublish(t *testing.T) {
	for _, binaryen := range []string{"", filepath.Join(t.TempDir(), "missing")} {
		out := filepath.Join(t.TempDir(), "output")
		err := buildShared(out, binaryen)
		if err == nil || !strings.Contains(err.Error(), "version_133") {
			t.Fatalf("missing tools must give actionable version error: %v", err)
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatalf("failed tool preflight created publication output: %v", err)
		}
	}
}
