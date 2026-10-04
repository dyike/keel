//go:build darwin && !ios && cgo

package window

import (
	"os/exec"
	"strings"
	"testing"
)

// The bridge reads the same scroller style AppKit reports to other apps.
func TestOverlayScrollersMatchesAppKit(t *testing.T) {
	out, err := exec.Command("osascript", "-l", "JavaScript", "-e", `ObjC.import("AppKit"); $.NSScroller.preferredScrollerStyle`).Output()
	if err != nil {
		t.Skip("osascript unavailable:", err)
	}
	want := strings.TrimSpace(string(out)) == "1"
	if got := overlayScrollers(); got != want {
		t.Fatalf("overlay scrollers %v, AppKit says %q", got, out)
	}
}
