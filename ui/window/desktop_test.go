package window

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestRealWindows opens native windows in a child process, which owns the main
// thread. Set KEEL_DESKTOP=1 to run it.
func TestRealWindows(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to open real windows")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "go", "run", "./testdata/raise").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "PASS") {
		t.Fatalf("%v\n%s", err, out)
	}
}
