//go:build darwin && !ios

package window

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dyike/keel/ui/internal/loop"
	"github.com/dyike/keel/ui/theme"
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

func TestScrollerCallbackDoesNotWaitForRedraw(t *testing.T) {
	old := theme.SystemScrollbarsAutoHide()
	defer theme.SetSystemScrollbarsAutoHide(old)
	theme.SetSystemScrollbarsAutoHide(false)

	// A redraw cannot finish until AppKit's preference callback returns.
	// Calling Invalidate on that thread would deadlock in Gio's invMu.
	resume := make(chan struct{})
	redraw := make(chan struct{}, 1)
	key := new(int)
	loop.Register(key, func() {
		select {
		case redraw <- struct{}{}:
		default:
		}
		<-resume
	})
	defer loop.Unregister(key)
	done := make(chan struct{})
	go func() {
		scrollersChanged(true)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		close(resume)
		<-done
		t.Fatal("AppKit callback waited for a redraw")
	}
	close(resume)
	select {
	case <-redraw:
	case <-time.After(time.Second):
		t.Fatal("preference update did not request a redraw")
	}
	loop.Lock()
	loop.Drain()
	loop.Unlock()
	if !theme.SystemScrollbarsAutoHide() {
		t.Fatal("scroller preference was not applied on the UI queue")
	}
}
