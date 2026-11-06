// SPDX-License-Identifier: Unlicense OR MIT

//go:build noopengl

package gpu

import (
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
)

func TestNoOpenGLRegistration(t *testing.T) {
	if driver.NewOpenGLDevice != nil {
		t.Fatal("noopengl still registers the GL backend")
	}
	if g, err := New(OpenGL{}); err == nil || g != nil {
		t.Fatalf("disabled backend: gpu=%v err=%v", g, err)
	}
}
