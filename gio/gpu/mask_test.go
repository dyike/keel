// SPDX-License-Identifier: Unlicense OR MIT
package gpu

import (
	"errors"
	"github.com/dyike/keel/gio/gpu/internal/driver"
	"image"
	"image/color"
	"testing"
)

type maskTestTexture struct {
	driver.Texture
	pixels   []byte
	released bool
}

func (m *maskTestTexture) Upload(_ image.Point, size image.Point, pixels []byte, stride int) {
	for y := 0; y < size.Y; y++ {
		m.pixels = append(m.pixels, pixels[y*stride:y*stride+size.X]...)
	}
}
func (m *maskTestTexture) Release() { m.released = true }

type maskTestDevice struct {
	*tintTestDevice
	mask *maskTestTexture
	fail bool
}

func (m *maskTestDevice) NewMaskTexture(_, _ int, _, _ driver.TextureFilter) (driver.Texture, error) {
	if m.fail {
		return nil, errors.New("allocation failed")
	}
	m.mask = new(maskTestTexture)
	return m.mask, nil
}
func TestMaskTextureAndPortableFallback(t *testing.T) {
	backing := image.NewAlpha(image.Rect(0, 0, 8, 5))
	src := backing.SubImage(image.Rect(2, 1, 5, 3)).(*image.Alpha)
	copy(src.Pix, []byte{0, 128, 255})
	copy(src.Pix[src.Stride:], []byte{32, 64, 192})
	for _, tc := range []struct {
		name                                             string
		unsupported, fail, opacity, clipped, reserveFail bool
	}{
		{name: "shared R8"}, {name: "allocation failure", fail: true}, {name: "unsupported", unsupported: true}, {name: "opacity", opacity: true}, {name: "path fallback", clipped: true}, {name: "quad reservation failure", reserveFail: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dev := &maskTestDevice{tintTestDevice: &tintTestDevice{prepared: !tc.reserveFail}, fail: tc.fail}
			var device driver.Device = dev
			if tc.unsupported {
				device = struct{ driver.Device }{dev}
			}
			r := renderer{ctx: device}
			cache := newTextureCache()
			defer cache.release()
			red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 128}
			images := []imageOp{{material: material{material: materialTexture, data: imageOpData{mask: src, handle: src, tint: red}}}, {material: material{material: materialTexture, data: imageOpData{mask: src, handle: src, tint: blue}}}}
			for i := range images {
				images[i].inLayer = tc.opacity
				if tc.clipped {
					images[i].clipType = clipTypePath
				}
			}
			r.uploadImages(cache, images)
			shared := !tc.opacity && !tc.clipped && !tc.reserveFail && !tc.unsupported
			if shared && !tc.fail {
				if dev.mask == nil || len(dev.textures) != 0 || images[0].material.tex != images[1].material.tex {
					t.Fatal("mask/tint did not share R8 texture")
				}
				want := []byte{0, 128, 255, 32, 64, 192}
				for i, v := range want {
					if dev.mask.pixels[i] != v {
						t.Fatal("subimage stride or alpha changed")
					}
				}
			} else {
				if dev.mask != nil {
					t.Fatal("fallback unexpectedly allocated R8")
				}
				wantTextures := 2
				if shared {
					wantTextures = 1
				}
				if len(dev.textures) != wantTextures {
					t.Fatalf("fallback textures %d want %d", len(dev.textures), wantTextures)
				}
				if shared {
					if got := dev.textures[0].pixels.RGBAAt(1, 0); got.A != 128 || got.R != maskWhite[128] {
						t.Fatal("lost linear coverage on allocation failure")
					}
				}
			}
			cache.frame()
			n := len(dev.textures)
			r.uploadImages(cache, images)
			if len(dev.textures) != n {
				t.Fatal("unchanged mask was uploaded again")
			}
			cache.frame()
			cache.frame()
			if dev.mask != nil && !dev.mask.released {
				t.Fatal("unused mask texture leaked")
			}
		})
	}
}
