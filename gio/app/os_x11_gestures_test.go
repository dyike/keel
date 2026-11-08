// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd || openbsd) && !nox11

package app

import (
	"github.com/dyike/keel/gio/io/system"
	"image"
	"testing"
)

func TestX11FramelessGestures(t *testing.T) {
	if got := x11DecorationHints(false); got != [5]uintptr{2, 0, 0, 0, 0} {
		t.Fatalf("frameless Motif hint: %v", got)
	}
	if got := x11DecorationHints(true); got[0] != 2 || got[2] != 1 {
		t.Fatalf("decorated Motif hint: %v", got)
	}
	for _, tc := range []struct {
		at        image.Point
		direction int
	}{
		{image.Pt(1, 1), 0}, {image.Pt(50, 1), 1}, {image.Pt(99, 1), 2}, {image.Pt(99, 50), 3},
		{image.Pt(99, 99), 4}, {image.Pt(50, 99), 5}, {image.Pt(1, 99), 6}, {image.Pt(1, 50), 7},
	} {
		got, ok := x11MoveResizeDirection(image.Pt(100, 100), 4, tc.at, 0, true)
		if !ok || got != tc.direction {
			t.Fatalf("resize at %v: %d %v", tc.at, got, ok)
		}
	}
	for _, at := range []image.Point{{50, 20}, {50, 80}} {
		if _, ok := x11MoveResizeDirection(image.Pt(100, 100), 4, at, 0, true); ok {
			t.Fatal("button/content became a move handle")
		}
	}
	if got, ok := x11MoveResizeDirection(image.Pt(100, 100), 4, image.Pt(50, 20), system.ActionMove, true); !ok || got != 8 {
		t.Fatal("titlebar is not draggable")
	}
	if _, ok := x11MoveResizeDirection(image.Pt(100, 100), 4, image.Pt(1, 1), 0, false); ok {
		t.Fatal("maximized window starts edge resize")
	}
}
