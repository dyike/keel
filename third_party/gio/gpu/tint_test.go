// SPDX-License-Identifier: Unlicense OR MIT

package gpu

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/gpu/internal/driver"
	"gioui.org/internal/f32color"
)

type tintTestTexture struct {
	driver.Texture
	pixels   *image.RGBA
	released bool
}

func (t *tintTestTexture) Upload(_ image.Point, size image.Point, pixels []byte, stride int) {
	t.pixels = image.NewRGBA(image.Rectangle{Max: size})
	for y := 0; y < size.Y; y++ {
		copy(t.pixels.Pix[y*t.pixels.Stride:], pixels[y*stride:y*stride+size.X*4])
	}
}
func (t *tintTestTexture) Release() { t.released = true }

type tintTestDevice struct {
	driver.Device
	prepared bool
	count    int
	textures []*tintTestTexture
}

func (d *tintTestDevice) NewTexture(_ driver.TextureFormat, _, _ int, _, _ driver.TextureFilter, _ driver.BufferBinding) (driver.Texture, error) {
	t := new(tintTestTexture)
	d.textures = append(d.textures, t)
	return t, nil
}
func (d *tintTestDevice) PrepareQuads(count int) bool                  { d.count = count; return d.prepared }
func (*tintTestDevice) DrawQuads([]driver.Texture, []driver.Quad) bool { panic("not used") }

func TestTintedImageTextureSharingAndFallback(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	src.SetRGBA(1, 0, f32color.NRGBAToRGBA(color.NRGBA{R: 255, G: 255, B: 255, A: 128}))
	before := append([]byte(nil), src.Pix...)
	red, blue := color.NRGBA{R: 255, A: 255}, color.NRGBA{B: 255, A: 128}
	for _, tc := range []struct {
		name                                          string
		prepared, layers, clipped, disabled, portable bool
	}{
		{name: "shared", prepared: true},
		{name: "allocation failure"},
		{name: "opacity layer", prepared: true, layers: true},
		{name: "path clip", prepared: true, clipped: true},
		{name: "disabled batching", prepared: true, disabled: true},
		{name: "portable device", portable: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dev := &tintTestDevice{prepared: tc.prepared}
			var device driver.Device = dev
			if tc.portable {
				device = struct{ driver.Device }{dev}
			}
			r := renderer{ctx: device, noQuads: tc.disabled}
			cache := newTextureCache()
			defer cache.release()
			handle := new(int)
			ops := []imageOp{{material: material{material: materialTexture, data: imageOpData{src: src, handle: handle, tint: red}}}, {material: material{material: materialTexture, data: imageOpData{src: src, handle: handle, tint: blue}}}}
			if tc.clipped {
				for i := range ops {
					ops[i].clipType = clipTypePath
				}
			}
			r.uploadImages(cache, ops, !tc.layers)
			shared := tc.prepared && !tc.layers && !tc.clipped && !tc.disabled && !tc.portable
			if shared {
				if len(dev.textures) != 1 || ops[0].material.tex != ops[1].material.tex {
					t.Fatal("colors duplicated the texture")
				}
				if dev.count != len(ops) {
					t.Fatal("did not reserve the entire frame")
				}
				if ops[1].material.color != f32color.LinearFromSRGB(blue) {
					t.Fatal("lost per-quad tint")
				}
			} else {
				if len(dev.textures) != 2 || ops[0].material.tex == ops[1].material.tex {
					t.Fatal("fallback did not separate colors")
				}
				for i, c := range []color.NRGBA{red, blue} {
					want := f32color.NRGBAToRGBA(c)
					if got := dev.textures[i].pixels.RGBAAt(0, 0); got != want {
						t.Fatalf("tint: got %v want %v", got, want)
					}
					if ops[i].material.color != (f32color.RGBA{R: 1, G: 1, B: 1, A: 1}) {
						t.Fatal("fallback applies tint twice")
					}
				}
			}
			n := len(dev.textures)
			cache.frame()
			r.uploadImages(cache, ops, !tc.layers)
			if len(dev.textures) != n {
				t.Fatal("cache miss on unchanged image and colors")
			}
			for i, v := range before {
				if src.Pix[i] != v {
					t.Fatal("modified immutable source")
				}
			}
			cache.frame()
			cache.frame()
			for _, tex := range dev.textures {
				if !tex.released {
					t.Fatal("unused texture leaked")
				}
			}
		})
	}
}
