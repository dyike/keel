//go:build darwin && !ios && !nometal && keelnativeqa

package window

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestNativeMaterialComposition(t *testing.T) {
	if os.Getenv("KEEL_DESKTOP") != "1" {
		t.Skip("set KEEL_DESKTOP=1 to open native material windows")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "go", "run", "-tags", "keelnativeqa", "./testdata/materialcapture").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "PASS:") {
		t.Fatalf("%v\n%s", err, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
