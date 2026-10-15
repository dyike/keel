// SPDX-License-Identifier: Unlicense OR MIT

package widget

import (
	"image/color"
	"testing"

	"golang.org/x/exp/shiny/materialdesign/icons"
)

func TestIconCachesAlternatingSizesAndColors(t *testing.T) {
	ic, err := NewIcon(icons.ActionHome)
	if err != nil {
		t.Fatal(err)
	}
	for i := range iconCacheEntries {
		ic.image(16+i, color.NRGBA{R: uint8(i * 20), A: 255})
	}
	first := ic.image(16, color.NRGBA{A: 255})
	if got := testing.AllocsPerRun(20, func() {
		for i := range iconCacheEntries {
			ic.image(16+i, color.NRGBA{R: uint8(i * 20), A: 255})
		}
	}); got != 0 {
		t.Fatalf("warm alternating icon allocated %.0f times", got)
	}
	if got := ic.image(16, color.NRGBA{A: 255}); got != first {
		t.Fatal("a cache hit changed image/texture identity")
	}
	// Touch first, then add a ninth variant. The oldest, not first, is evicted.
	oldest := ic.image(17, color.NRGBA{R: 20, A: 255})
	ic.image(16, color.NRGBA{A: 255})
	for i := 2; i < iconCacheEntries; i++ {
		ic.image(16+i, color.NRGBA{R: uint8(i * 20), A: 255})
	}
	ic.image(100, color.NRGBA{B: 255, A: 255})
	if ic.image(16, color.NRGBA{A: 255}) != first {
		t.Fatal("hot variant was evicted")
	}
	if ic.image(17, color.NRGBA{R: 20, A: 255}) == oldest {
		t.Fatal("cold variant was not evicted")
	}
}

func TestIconCachePixelBudgetAndOversizedImages(t *testing.T) {
	ic, err := NewIcon(icons.ActionHome)
	if err != nil {
		t.Fatal(err)
	}
	for i := range 12 {
		ic.image(240, color.NRGBA{R: uint8(i), A: 255})
		if ic.imageCount > iconCacheEntries || ic.imageBytes > iconCacheBytes {
			t.Fatal("pixel budget exceeded")
		}
	}
	big := ic.image(600, color.NRGBA{A: 255})
	if ic.imageCount != 1 || ic.imageBytes <= iconCacheBytes {
		t.Fatal("oversized icon must retain only itself")
	}
	if ic.image(600, color.NRGBA{A: 255}) != big {
		t.Fatal("oversized current image was rerasterized")
	}
	ic.image(16, color.NRGBA{A: 255})
	if ic.imageBytes > iconCacheBytes {
		t.Fatal("small icon retained oversized storage")
	}
	for _, entry := range ic.images[ic.imageCount:] {
		if entry != (iconImage{}) {
			t.Fatal("evicted images kept texture references")
		}
	}
}

func BenchmarkIconAlternatingVariants(b *testing.B) {
	ic, err := NewIcon(icons.ActionHome)
	if err != nil {
		b.Fatal(err)
	}
	for i := range iconCacheEntries {
		ic.image(16+i, color.NRGBA{R: uint8(i * 20), A: 255})
	}
	b.ReportAllocs()
	for b.Loop() {
		for i := range iconCacheEntries {
			ic.image(16+i, color.NRGBA{R: uint8(i * 20), A: 255})
		}
	}
}
