// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"testing"

	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/dyike/keel/gio/internal/ops"
	gioop "github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
)

type surfaceTestDevice struct {
	mutableTestDevice
	wrapped []uintptr
	filters [][2]driver.TextureFilter
}

func (d *surfaceTestDevice) NewSurfaceTexture(ref uintptr, w, h int, minF, magF driver.TextureFilter) (driver.Texture, error) {
	d.wrapped = append(d.wrapped, ref)
	d.filters = append(d.filters, [2]driver.TextureFilter{minF, magF})
	t := new(mutableTestTexture)
	d.textures = append(d.textures, t)
	return t, nil
}

// surfaceData reads an op the way the renderer does, through its encoding.
func surfaceData(t *testing.T, op paint.ImageOp) imageOpData {
	t.Helper()
	var o gioop.Ops
	op.Add(&o)
	var r ops.Reader
	r.Reset(&o.Internal)
	for e, ok := r.Decode(); ok; e, ok = r.Decode() {
		if ops.OpType(e.Data[0]) == ops.TypeImage {
			d := decodeImageOp(e.Data, e.Refs)
			d.tint = whiteTint
			return d
		}
	}
	t.Fatal("no image op")
	return imageOpData{}
}

func TestSurfaceImageWrapsEachPicture(t *testing.T) {
	var s paint.Surface
	op := paint.NewSurfaceOp(&s)
	if op.Size() != (image.Point{}) {
		t.Fatalf("empty surface size %v", op.Size())
	}
	s.Set(0x1000, image.Pt(64, 32))
	if op.Size() != image.Pt(64, 32) {
		t.Fatalf("size %v", op.Size())
	}
	dev := &surfaceTestDevice{}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()

	data := surfaceData(t, op)
	if data.size() != image.Pt(64, 32) {
		t.Fatalf("decoded size %v", data.size())
	}
	tex := r.texHandle(cache, data)
	if len(dev.wrapped) != 1 || dev.wrapped[0] != 0x1000 || len(dev.textures[0].uploads) != 0 {
		t.Fatalf("first picture: wrapped %v, uploads %d", dev.wrapped, len(dev.textures[0].uploads))
	}
	if dev.filters[0] != [2]driver.TextureFilter{driver.FilterLinear, driver.FilterLinear} {
		t.Errorf("filters %v, want linear without mipmaps", dev.filters[0])
	}

	// The same picture is not wrapped again.
	cache.frame()
	if r.texHandle(cache, surfaceData(t, op)) != tex || len(dev.wrapped) != 1 {
		t.Fatal("unchanged surface wrapped again")
	}

	// A new picture gets a new texture; the old one is released.
	s.Set(0x2000, image.Pt(64, 32))
	cache.frame()
	if r.texHandle(cache, surfaceData(t, op)) == tex || len(dev.wrapped) != 2 || dev.wrapped[1] != 0x2000 {
		t.Fatalf("new picture: wrapped %v", dev.wrapped)
	}
	if !dev.textures[0].released {
		t.Error("previous picture's texture not released")
	}
}

func TestSurfaceImageWithoutDeviceSupport(t *testing.T) {
	var s paint.Surface
	s.Set(0x1000, image.Pt(8, 8))
	dev := &mutableTestDevice{}
	r := renderer{ctx: dev}
	cache := newTextureCache()
	defer cache.release()
	a := r.texHandle(cache, surfaceData(t, paint.NewSurfaceOp(&s)))
	b := r.texHandle(cache, surfaceData(t, paint.NewSurfaceOp(&s)))
	if a == nil || a != b || len(dev.textures) != 1 || len(dev.textures[0].uploads) != 1 {
		t.Fatalf("fallback: %d textures", len(dev.textures))
	}
	if u := dev.textures[0].uploads[0]; u.size != image.Pt(1, 1) || u.pixels[3] != 0 {
		t.Fatalf("fallback is not a transparent pixel: %+v", u)
	}
}
