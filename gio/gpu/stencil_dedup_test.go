// SPDX-License-Identifier: Unlicense OR MIT
package gpu

import (
	"github.com/dyike/keel/gio/internal/f32"
	"image"
	"testing"
)

func TestStencilSharingUsesExactRasterInputs(t *testing.T) {
	r := &renderer{}
	r.packer.maxDims = image.Pt(8192, 8192)
	base := pathOp{path: true, pathKey: opKey{content: 42}, clip: image.Rect(10, 20, 110, 80), off: f32.Pt(10, 20)}
	variants := []pathOp{base, base, base, base, base, base}
	variants[1].parent = &pathOp{pathKey: opKey{content: 99}}
	variants[2].off.X += .5
	variants[3].clip.Max.X--
	variants[4].pathKey.content++
	variants[5].pathKey.strokeWidth = 2
	paths := make([]*pathOp, len(variants))
	for i := range variants {
		paths[i] = &variants[i]
	}
	r.packStencils(&paths)
	if len(paths) != 5 {
		t.Fatalf("unique masks=%d, want 5", len(paths))
	}
	if variants[0].place != variants[1].place {
		t.Fatal("identical masks under different parents did not share")
	}
	for i := 2; i < len(variants); i++ {
		if variants[i].place == variants[0].place {
			t.Fatalf("different mask %d shared", i)
		}
	}
	// Repacking the next frame must discard prior placement and empty masks.
	empty := base
	empty.clip = image.Rectangle{}
	next := []*pathOp{&variants[2], &empty}
	r.packStencils(&next)
	if len(next) != 1 || next[0].place.Pos != (image.Point{}) {
		t.Fatal("stale atlas placement reused")
	}
}
