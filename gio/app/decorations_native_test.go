// SPDX-License-Identifier: Unlicense OR MIT

//go:build (darwin && !ios) || windows

package app

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/op"
)

func TestNativeDecorationsNeverInsetClient(t *testing.T) {
	var w Window
	if h := w.initDecorations(); h != 0 {
		t.Fatalf("native title bar requested %v dp of client area", h)
	}
	for _, decorated := range []bool{false, true} {
		for _, mode := range []WindowMode{Windowed, Maximized, Minimized, Fullscreen} {
			w.decorations.enabled = decorated
			w.decorations.Config = Config{Decorated: decorated, Mode: mode, Size: image.Pt(640, 480)}
			w.updateDecorationMode(mode)
			if w.fallbackDecorate() {
				t.Fatal("native platform requested a fallback title bar")
			}
			var ops op.Ops
			if p := w.decorate(FrameEvent{}, &ops); p != (image.Point{}) {
				t.Fatal("client inset", p)
			}
			if cfg := w.effectiveConfig(); cfg.Size != image.Pt(640, 480) || cfg.Decorated != decorated || cfg.Mode != mode {
				t.Fatalf("config changed: %+v", cfg)
			}
		}
	}
}
