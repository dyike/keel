package window

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestRealWindows opens native windows in child processes, which own the main
// thread. Set KEEL_DESKTOP=1 to run it.
func TestRealWindows(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to open real windows")
	}
	cases := [][]string{
		{"./testdata/raise"},                   // Raise/Close from UI code must not deadlock
		{"./testdata/reopen", "update", "now"}, // closing right after opening must not crash
		{"./testdata/reopen", "locked", "now"},
	}
	if runtime.GOOS == "darwin" {
		cases = append(cases, []string{"./testdata/center"})
	}
	for _, args := range cases {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, "go", append([]string{"run"}, args...)...).CombinedOutput()
			if err != nil || !strings.Contains(string(out), "PASS") {
				t.Fatalf("%v\n%s", err, out)
			}
		})
	}
}
