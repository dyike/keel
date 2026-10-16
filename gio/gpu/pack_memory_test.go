// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"testing"
)

func TestPackingCornerBandsDoesNotAllocateTheirEmptyInterior(t *testing.T) {
	p := packer{maxDims: image.Pt(8192, 8192)}
	// A 4000-pixel-wide window's top/bottom corners and a tall pane's
	// side strokes. One bounding atlas used to approach a full window.
	sizes := []image.Point{{4000, 32}, {4000, 32}, {3968, 24}, {3968, 24}, {4, 2044}, {4, 2044}}
	var placed []struct {
		page int
		rect image.Rectangle
	}
	for _, size := range sizes {
		place, ok := p.add(size)
		if !ok {
			t.Fatal("could not place", size)
		}
		r := image.Rectangle{Min: place.Pos, Max: place.Pos.Add(size)}
		for _, old := range placed {
			if old.page == place.Idx && old.rect.Overlaps(r) {
				t.Fatal("overlapping coverage masks")
			}
		}
		placed = append(placed, struct {
			page int
			rect image.Rectangle
		}{place.Idx, r})
	}
	var pixels int
	for _, size := range p.sizes {
		pixels += size.X * size.Y
	}
	if pixels > 1<<20 {
		t.Fatalf("sparse corner/border atlas reserves %d pixels", pixels)
	}
	for _, v := range placed {
		if !v.rect.In(image.Rectangle{Max: p.sizes[v.page]}) {
			t.Fatal("mask outside its page")
		}
	}
	p.clear()
	place, ok := p.add(image.Pt(10, 20))
	if !ok || place.Idx != 0 || place.Pos != (image.Point{}) || p.area != 200 {
		t.Fatal("packing state survived clear")
	}
}

func TestPackingDenseMasksKeepsOnePage(t *testing.T) {
	p := packer{maxDims: image.Pt(2048, 2048)}
	for range 100 {
		if _, ok := p.add(image.Pt(100, 100)); !ok {
			t.Fatal("could not place dense mask")
		}
	}
	if len(p.sizes) != 1 || p.area != 1000000 {
		t.Fatalf("dense masks split into %d pages, area %d", len(p.sizes), p.area)
	}
}
