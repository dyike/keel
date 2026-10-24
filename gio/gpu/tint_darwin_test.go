// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package gpu

import (
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"image"
	"image/color"
	"testing"
)

func TestUnrelatedOpacityKeepsAtlasTextureShared(t *testing.T) {
	dev := &tintTestDevice{prepared: true}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	var frame op.Ops
	src := image.NewRGBA(image.Rect(0, 0, 1024, 1024))
	img := paint.NewImageOp(src)
	for i := range 8 {
		tr := op.Offset(image.Pt(i*12, 0)).Push(&frame)
		c := clip.Rect(image.Rect(0, 0, 10, 10)).Push(&frame)
		img.AddTinted(&frame, color.NRGBA{R: uint8(20 + i*20), A: 255})
		paint.PaintOp{}.Add(&frame)
		c.Pop()
		tr.Pop()
	}
	opacity := paint.PushOpacity(&frame, .5)
	paint.FillShape(&frame, color.NRGBA{A: 255}, clip.Rect(image.Rect(0, 20, 10, 30)).Op())
	opacity.Pop()
	d := drawOps{pathCache: newOpCache()}
	d.reset(image.Pt(100, 100))
	d.collect(&frame, image.Pt(100, 100))
	r.uploadImages(cache, d.imageOps)
	if len(dev.textures) != 1 {
		t.Fatalf("unrelated opacity duplicated atlas into %d textures; want 1", len(dev.textures))
	}
}
