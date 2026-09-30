//go:build darwin && cgo

package platform

import (
	"bytes"
	"image/png"
	"os"
	"runtime"
	"testing"
)

// Keep the process main thread assigned to the test runner, as a GUI host does.
func init() { runtime.LockOSThread() }

// Opt-in: reads current desktop pixels in memory; never saves or logs the screenshot.
func TestNativeCapture(t *testing.T) {
	if os.Getenv("KEEL_TEST_CAPTURE") != "1" {
		t.Skip("set KEEL_TEST_CAPTURE=1 with existing Screen Recording permission")
	}
	b := New()
	displays, err := b.Displays()
	if err != nil {
		t.Fatal(err)
	}
	if len(displays) == 0 {
		t.Fatal("no displays")
	}
	data, err := b.Capture(displays[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != displays[0].PixelWidth || cfg.Height != displays[0].PixelHeight {
		t.Fatalf("unexpected image dimensions: %dx%d", cfg.Width, cfg.Height)
	}
}
