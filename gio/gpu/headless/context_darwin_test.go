// SPDX-License-Identifier: Unlicense OR MIT

//go:build !ios

package headless

import (
	"testing"

	"github.com/dyike/keel/gio/gpu"
	"github.com/ebitengine/purego/objc"
)

// TestMetalContext checks the context either binding creates (go test and
// go test -tags purego): the queue must come from the device it reports.
func TestMetalContext(t *testing.T) {
	ctx, err := newContextPrimary()
	if err != nil {
		t.Skip(err)
	}
	defer ctx.Release()
	api, ok := ctx.API().(gpu.Metal)
	if !ok {
		t.Fatalf("API() = %T, want gpu.Metal", ctx.API())
	}
	if api.Device == 0 || api.Queue == 0 {
		t.Fatalf("device %#x, queue %#x", api.Device, api.Queue)
	}
	if got := uintptr(objc.ID(api.Queue).Send(objc.RegisterName("device"))); got != api.Device {
		t.Errorf("queue's device = %#x, want %#x", got, api.Device)
	}
	if api.PixelFormat != 71 { // MTLPixelFormatRGBA8Unorm_sRGB
		t.Errorf("pixel format = %d, want 71", api.PixelFormat)
	}
}
