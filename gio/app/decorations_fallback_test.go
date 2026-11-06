// SPDX-License-Identifier: Unlicense OR MIT

//go:build (!darwin || ios) && !windows

package app

import "testing"

func TestFallbackDecorationState(t *testing.T) {
	var w Window
	if h := w.initDecorations(); h <= 0 {
		t.Fatal("missing fallback title bar height", h)
	}
	w.decorations.enabled = true
	w.decorations.Config = Config{Decorated: false, Mode: Windowed}
	if !w.fallbackDecorate() {
		t.Fatal("server without decorations needs client-side fallback")
	}
	w.updateDecorationMode(Maximized)
	if !w.decorations.Decorations.Maximized {
		t.Fatal("maximize state lost")
	}
	w.updateDecorationMode(Windowed)
	if w.decorations.Decorations.Maximized {
		t.Fatal("restore state lost")
	}
	w.decorations.Config.Decorated = true
	if w.fallbackDecorate() {
		t.Fatal("server decorations must not be doubled")
	}
	w.decorations.Config.Decorated = false
	w.decorations.Config.Mode = Fullscreen
	if w.fallbackDecorate() {
		t.Fatal("fullscreen must not show a title bar")
	}
	w.decorations.Config.Mode = Windowed
	w.nocontext = true
	if w.fallbackDecorate() {
		t.Fatal("custom renderer must not receive fallback drawing")
	}
}
