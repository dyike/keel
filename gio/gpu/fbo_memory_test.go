package gpu

import (
	"github.com/dyike/keel/gio/gpu/internal/driver"
	"image"
	"testing"
)

type memoryTexture struct {
	driver.Texture
	released bool
}

func (t *memoryTexture) Release() { t.released = true }

type memoryDevice struct {
	driver.Device
	textures []*memoryTexture
}

func (d *memoryDevice) Caps() driver.Caps { return driver.Caps{MaxTextureSize: 8192} }
func (d *memoryDevice) NewTexture(_ driver.TextureFormat, _, _ int, _, _ driver.TextureFilter, _ driver.BufferBinding) (driver.Texture, error) {
	t := new(memoryTexture)
	d.textures = append(d.textures, t)
	return t, nil
}

func TestFramebuffersReleaseOversizedStorage(t *testing.T) {
	d := new(memoryDevice)
	var s fboSet
	s.resize(d, driver.TextureFormatFloat, []image.Point{{4000, 3000}})
	large := d.textures[0]
	s.resize(d, driver.TextureFormatFloat, []image.Point{{1000, 800}})
	if !large.released || s.fbos[0].size.X > 1100 || s.fbos[0].size.Y > 880 {
		t.Fatalf("kept oversized framebuffer: %v", s.fbos[0].size)
	}
	current := s.fbos[0].tex
	s.resize(d, driver.TextureFormatFloat, []image.Point{{1000, 800}})
	if s.fbos[0].tex != current {
		t.Fatal("stable frame recreated its framebuffer")
	}
	s.resize(d, driver.TextureFormatFloat, nil)
	if !d.textures[len(d.textures)-1].released {
		t.Fatal("unused framebuffer retained")
	}
}
