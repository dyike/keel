// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/dyike/keel/gio/internal/f32color"
)

type coveredBatch struct {
	tex   []driver.Texture
	quads []driver.Quad
}
type coveredDevice struct {
	driver.Device
	batches []coveredBatch
}

func (*coveredDevice) PrepareQuads(int) bool       { return true }
func (*coveredDevice) SupportsCoverageQuads() bool { return true }
func (d *coveredDevice) DrawQuads(tex []driver.Texture, quads []driver.Quad) bool {
	d.batches = append(d.batches, coveredBatch{append([]driver.Texture(nil), tex...), append([]driver.Quad(nil), quads...)})
	return true
}

func TestCoveredQuadsTextureSlotsAndOrder(t *testing.T) {
	dev := new(coveredDevice)
	coverage, intersection := new(tintTestTexture), new(tintTestTexture)
	r := renderer{ctx: dev, pather: &pather{stenciler: &stenciler{
		fbos:          fboSet{fbos: []FBO{{tex: coverage, size: image.Pt(100, 100)}}},
		intersections: fboSet{fbos: []FBO{{tex: intersection, size: image.Pt(100, 100)}}},
	}}}
	var images []imageOp
	for i := 0; i < 12; i++ {
		img := imageOp{clip: image.Rect(i, 0, i+10, 20), clipType: clipTypePath, material: material{material: materialTexture, tex: new(tintTestTexture), opacity: 1, color: f32color.RGBA{R: .5, A: 1}}}
		img.place.Pos = image.Pt(20, 30)
		if i%2 != 0 {
			img.clipType = clipTypeIntersection
		}
		images = append(images, img)
	}
	r.drawOps(false, image.Point{}, image.Pt(100, 100), images)
	if len(dev.batches) < 2 {
		t.Fatal("texture limit did not flush")
	}
	index := 0
	for _, b := range dev.batches {
		if len(b.tex) > driver.MaxQuadTextures {
			t.Fatal("texture slot overflow")
		}
		for _, q := range b.quads {
			wantCover := driver.Texture(coverage)
			if index%2 != 0 {
				wantCover = intersection
			}
			if b.tex[int(q.UV0[3])] != images[index].material.tex || b.tex[int(q.UV1[3])-1] != wantCover {
				t.Fatal("lost image order or coverage binding")
			}
			if q.Color != ([4]float32{.5, 0, 0, 1}) {
				t.Fatal("lost tint")
			}
			if q.Cover != ([4]float32{.1, .2, .2, .3}) {
				t.Fatalf("coverage UV=%v", q.Cover)
			}
			index++
		}
	}
	if index != len(images) {
		t.Fatal("lost quads")
	}
}
