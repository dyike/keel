// SPDX-License-Identifier: Unlicense OR MIT

//go:build (linux || freebsd || openbsd) && !noopengl

package headless

import (
	"image"
	"image/color"
	"testing"

	"github.com/dyike/keel/gio/internal/egl"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
)

func TestReleasingPeerPreservesEGLDisplay(t *testing.T) {
	first, err := NewWindow(32, 32)
	if err != nil {
		t.Skipf("GPU unavailable: %v", err)
	}
	defer func() {
		if first != nil {
			first.Release()
		}
	}()
	if _, ok := first.ctx.(*egl.Context); !ok {
		t.Skip("requires EGL backend")
	}
	second, err := NewWindow(32, 32)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Release()
	first.Release()
	first = nil
	var ops op.Ops
	want := color.NRGBA{R: 255, A: 255}
	paint.Fill(&ops, want)
	if err := second.Frame(&ops); err != nil {
		t.Fatal("surviving context cannot draw:", err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	if err := second.Screenshot(img); err != nil {
		t.Fatal(err)
	}
	if got := img.RGBAAt(16, 16); got != (color.RGBA{R: 255, A: 255}) {
		t.Fatal("surviving context lost output", got)
	}
}
