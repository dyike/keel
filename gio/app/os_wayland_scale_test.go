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
