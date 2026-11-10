package window

import (
	"github.com/dyike/keel/gio/layout"
	"github.com/dyike/keel/gio/unit"
	"github.com/dyike/keel/ui/core"
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestSizeStateRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "windows", "main.json")
	state, err := LoadSizeState(path)
	if err != nil {
		t.Fatal(err)
	}
	state.remember(image.Pt(1234, 789))
	if err := state.Save(); err != nil {
		t.Fatal(err)
	}
	restored, err := LoadSizeState(path)
	if err != nil {
		t.Fatal(err)
	}
	w := newWindow(Options{AutoSize: true, Width: 640, Height: 480, SizeState: restored})
	if w.opts.Width != 1234 || w.opts.Height != 789 || w.opts.AutoSize {
		t.Fatal(w.opts)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
	for _, content := range []string{`not json`, `{"X":-1,"Y":200}`, `{"X":999999,"Y":200}`} {
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		state, err := LoadSizeState(path)
		if err == nil || state.snapshot() != (image.Point{}) {
			t.Fatal(content, err)
		}
	}
}
func TestInitialWindowSize(t *testing.T) {
	for _, tc := range []struct {
		available, requested, want image.Point
		auto                       bool
	}{
		{image.Pt(3000, 1700), image.Pt(640, 480), image.Pt(1600, 1000), true},
		{image.Pt(1200, 800), image.Pt(640, 480), image.Pt(900, 600), true},
		{image.Pt(800, 600), image.Pt(1400, 900), image.Pt(800, 600), false},
		{image.Pt(3000, 1700), image.Pt(1100, 760), image.Pt(1100, 760), false},
		{image.Point{}, image.Pt(960, 640), image.Pt(960, 640), true},
	} {
		if got := initialWindowSize(tc.requested, tc.available, tc.auto); got != tc.want {
			t.Fatal(tc, got)
		}
	}
}

func TestSizeStateIgnoresMaximizedFrames(t *testing.T) {
	state := new(SizeState)
	w := newWindow(Options{SizeState: state})
	frame := func(size image.Point) {
		w.observeSize(core.C{Constraints: layout.Exact(size), Metric: unit.Metric{PxPerDp: 1.25, PxPerSp: 2}})
	}
	frame(image.Pt(1000, 750))
	if got := state.snapshot(); got != image.Pt(800, 600) {
		t.Fatal(got)
	}
	w.normalMode = false
	frame(image.Pt(3000, 2000))
	if got := state.snapshot(); got != image.Pt(800, 600) {
		t.Fatal("maximized frame replaced normal size", got)
	}
	w.normalMode = true
	frame(image.Pt(1250, 875))
	if got := state.snapshot(); got != image.Pt(1000, 700) {
		t.Fatal(got)
	}
}
