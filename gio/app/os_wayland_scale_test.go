// SPDX-License-Identifier: Unlicense OR MIT

//go:build ((linux && !android) || freebsd) && !nowayland

package app

import (
	"github.com/dyike/keel/gio/unit"
	"image"
	"testing"
)

func TestWaylandTextScaleDoesNotResizeLayout(t *testing.T) {
	for _, outputScale := range []int{1, 2, 3} {
		for _, textScale := range []float32{1, 1.3636, 2} {
			w := window{scale: outputScale, textScale: textScale, size: image.Pt(800, 600)}
			size, metric := w.getConfig()
			if size != image.Pt(800*outputScale, 600*outputScale) {
				t.Fatal(size)
			}
			if metric.Dp(unit.Dp(39)) != 39*outputScale {
				t.Fatalf("titlebar scaled with text: %v", metric)
			}
			if metric.PxPerSp != float32(outputScale)*textScale {
				t.Fatalf("text preference lost: %v", metric)
			}
		}
	}
}
func TestParseWaylandTextScale(t *testing.T) {
	for _, s := range []string{"", "0", "-1", "NaN", "+Inf", "garbage"} {
		if got := parseTextScale([]byte(s)); got != 1 {
			t.Fatalf("%q: %v", s, got)
		}
	}
	if got := parseTextScale([]byte(" 1.3636\n")); got != float32(1.3636) {
		t.Fatal(got)
	}
}

func TestWaylandFractionalMetricsAndCoordinates(t *testing.T) {
	for _, numerator := range []uint32{120, 150, 180, 240} {
		for _, text := range []float32{1, 1.6364, 2} {
			w := window{scale: 2, fractionalScale: numerator, textScale: text, size: image.Pt(801, 603)}
			size, metric := w.getConfig()
			scale := float32(numerator) / 120
			if metric.PxPerDp != scale || metric.PxPerSp != scale*text {
				t.Fatal(metric)
			}
			if w.surfaceSize(size) != w.size {
				t.Fatalf("round trip: %v -> %v", w.size, size)
			}
			if got := w.bufferSize(image.Pt(2, 2)); numerator == 150 && got != image.Pt(3, 3) {
				t.Fatalf("half pixels must round away from zero: %v", got)
			}
		}
	}
	w := window{scale: 2, textScale: 1, size: image.Pt(800, 600)}
	size, _ := w.getConfig()
	if size != image.Pt(1600, 1200) {
		t.Fatal("integer fallback", size)
	}
}

func TestWaylandFramelessHasNoDecorationHeight(t *testing.T) {
	w := window{config: Config{decoHeight: 40}}
	if w.decoHeight() != 0 {
		t.Fatal("frameless window gained titlebar height")
	}
	w.requestedDecorated = true
	if w.decoHeight() != 40 {
		t.Fatal("client-side titlebar height lost")
	}
	w.config.Decorated = true
	if w.decoHeight() != 0 {
		t.Fatal("server titlebar counted as client pixels")
	}
}

func TestTextScaleDesktopSource(t *testing.T) {
	for _, tc := range []struct{ current, session, want string }{
		{"GNOME", "", "org.gnome.desktop.interface"},
		{"ubuntu:GNOME", "", "org.gnome.desktop.interface"},
		{"X-Cinnamon", "", "org.cinnamon.desktop.interface"},
		{"", "gnome", "org.gnome.desktop.interface"},
		{"Hyprland", "gnome", ""}, {"KDE", "", ""}, {"", "", ""},
	} {
		if got := textScaleSchema(tc.current, tc.session); got != tc.want {
			t.Errorf("%q/%q: %q, want %q", tc.current, tc.session, got, tc.want)
		}
	}
}
