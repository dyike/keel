//go:build darwin && cgo

package desktop_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// A child process owns AppKit's main thread and single-use launch lifecycle.
func TestNativeApplicationLifecycle(t *testing.T) {
	if os.Getenv("KEEL_TEST_DESKTOP") != "1" {
		t.Skip("set KEEL_TEST_DESKTOP=1 to open temporary native windows")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "run", "./testdata/lifecycle")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("native lifecycle: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "PASS:") {
		t.Fatalf("missing completion marker:\n%s", out)
	}
}
