// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"fmt"
	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
	"image"
	"runtime"
	"testing"
	"unsafe"
)

func testBackend(t *testing.T) *Backend {
	t.Helper()
	runtime.LockOSThread()
	t.Cleanup(runtime.UnlockOSThread)
	lib, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatal(err)
	}
	var create func() uintptr
	purego.RegisterLibFunc(&create, lib, "MTLCreateSystemDefaultDevice")
	uploadInit()
	pool := uploadSend(uintptr(objc.GetClass("NSAutoreleasePool")), "new")
	t.Cleanup(func() { objc.ID(pool).Send(objc.RegisterName("drain")) })
	d := create()
	if d == 0 {
		t.Skip("no Metal device")
	}
	q := uintptr(objc.ID(d).Send(objc.RegisterName("newCommandQueue")))
	v, err := newMetalDevice(driver.Metal{Device: d, Queue: q, PixelFormat: 71})
	if err != nil {
		t.Fatal(err)
	}
	b := v.(*Backend)
	t.Cleanup(b.Release)
	uploadSend(q, "release")
	uploadSend(d, "release")
	return b
}
func TestUploadMipmapsMatchMetal(t *testing.T) {
	b := testBackend(t)
	var read func(uintptr, uintptr, unsafe.Pointer, uintptr, uploadRegion, uintptr)
	purego.RegisterFunc(&read, uploadMsg)
	readSel := uintptr(objc.RegisterName("getBytes:bytesPerRow:fromRegion:mipmapLevel:"))
	for _, sz := range []image.Point{{16, 16}, {7, 5}, {1, 16}, {96, 96}, {255, 127}} {
		for _, format := range []driver.TextureFormat{driver.TextureFormatSRGBA, driver.TextureFormatRGBA8} {
			stride := sz.X*4 + 12
			p := make([]byte, stride*sz.Y)
			for i := range p {
				p[i] = byte(i*37 + i/7)
			}
			makeTex := func() *Texture {
				v, e := b.NewTexture(format, sz.X, sz.Y, driver.FilterLinearMipmapLinear, driver.FilterLinear, driver.BufferBindingTexture)
				if e != nil {
					t.Fatal(e)
				}
				return v.(*Texture)
			}
			got, want := makeTex(), makeTex()
			want.direct = false
			b.BeginFrame(nil, false, sz)
			got.Upload(image.Point{}, sz, p, stride)
			want.Upload(image.Point{}, sz, p, stride)
			b.endCmdBuffer(true)
			dim := sz
			for level := 0; ; level++ {
				a, c := make([]byte, dim.X*dim.Y*4), make([]byte, dim.X*dim.Y*4)
				r := uploadRegion{Width: uintptr(dim.X), Height: uintptr(dim.Y), Depth: 1}
				read(uintptr(got.texture), readSel, unsafe.Pointer(&a[0]), uintptr(dim.X*4), r, uintptr(level))
				read(uintptr(want.texture), readSel, unsafe.Pointer(&c[0]), uintptr(dim.X*4), r, uintptr(level))
				maxDiff := 0
				for i := range a {
					d := int(a[i]) - int(c[i])
					if d < 0 {
						d = -d
					}
					maxDiff = max(maxDiff, d)
				}
				if maxDiff > 2 {
					t.Errorf("size=%v format=%v level=%d max channel difference=%d", sz, format, level, maxDiff)
				}
				if dim.X == 1 && dim.Y == 1 {
					break
				}
				dim = image.Pt(max(1, dim.X/2), max(1, dim.Y/2))
			}
			got.Release()
			want.Release()
		}
	}
}

func TestInitialUploadAndOrderedFallback(t *testing.T) {
	b := testBackend(t)
	if !b.unified {
		t.Skip("unified memory required")
	}
	size := image.Pt(8, 8)
	v, e := b.NewTexture(driver.TextureFormatRGBA8, 8, 8, driver.FilterLinearMipmapLinear, driver.FilterLinear, driver.BufferBindingTexture)
	if e != nil {
		t.Fatal(e)
	}
	tex := v.(*Texture)
	defer tex.Release()
	b.BeginFrame(nil, false, size)
	src := make([]byte, 8*8*4)
	for i := range src {
		src[i] = 80
	}
	tex.Upload(image.Point{}, size, src, 32)
	if b.blitEnc != 0 || b.stagingBuf != 0 {
		t.Fatal("initial image upload started a GPU copy")
	}
	if tex.direct {
		t.Fatal("later uploads must use the ordered path")
	}
	geometry, e := b.NewImmutableBuffer(driver.BufferBindingIndices, make([]byte, 8192))
	if e != nil {
		t.Fatal(e)
	}
	defer geometry.Release()
	if b.blitEnc != 0 || b.stagingBuf != 0 {
		t.Fatal("immutable upload started a GPU copy")
	}
	// A partial update must not regenerate mipmaps from only the updated subrectangle.
	tex.Upload(image.Pt(2, 3), image.Pt(1, 1), []byte{240, 241, 242, 243}, 4)
	if b.blitEnc == 0 {
		t.Fatal("partial update bypassed GPU ordering")
	}
	b.endCmdBuffer(true)
	out := make([]byte, len(src))
	if e := tex.ReadPixels(image.Rect(0, 0, 8, 8), out, 32); e != nil {
		t.Fatal(e)
	}
	copy(src[3*32+2*4:], []byte{240, 241, 242, 243})
	for i := range src {
		if src[i] != out[i] {
			t.Fatalf("partial update byte %d: got %d want %d", i, out[i], src[i])
		}
	}
	// Later writes to a shared immutable buffer still use GPU copy ordering.
	geometry.Upload([]byte{1, 2, 3, 4})
	dst := make([]byte, 4)
	if e := geometry.Download(dst); e != nil {
		t.Fatal(e)
	}
	for i, v := range dst {
		if v != byte(i+1) {
			t.Fatal("buffer update failed", dst)
		}
	}
}
func TestNonUnifiedUploadFallback(t *testing.T) {
	b := testBackend(t)
	b.unified = false
	v, e := b.NewTexture(driver.TextureFormatRGBA8, 2, 2, driver.FilterNearest, driver.FilterNearest, driver.BufferBindingTexture)
	if e != nil {
		t.Fatal(e)
	}
	defer v.Release()
	tex := v.(*Texture)
	if tex.direct {
		t.Fatal("discrete GPU selected direct upload")
	}
	b.BeginFrame(nil, false, image.Pt(2, 2))
	tex.Upload(image.Point{}, image.Pt(2, 2), make([]byte, 16), 8)
	if b.blitEnc == 0 {
		t.Fatal("missing upload fallback")
	}
	b.endCmdBuffer(true)
}
func BenchmarkCPUMipmaps(b *testing.B) {
	initMipGamma()
	for _, n := range []int{96, 256, 1024} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			p := make([]byte, n*n*4)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dim := image.Pt(n, n)
				src := p
				for dim.X > 1 || dim.Y > 1 {
					next := image.Pt(max(1, dim.X/2), max(1, dim.Y/2))
					src = downsampleMip(dim, next, src, dim.X*4, true)
					dim = next
				}
			}
		})
	}
}

func TestMaskTextureStorageAndOrderedUpload(t *testing.T) {
	b := testBackend(t)
	if !b.unified {
		t.Skip("unified-memory R8 masks unavailable")
	}
	sz := image.Pt(256, 256)
	tex, err := b.NewMaskTexture(sz.X, sz.Y, driver.FilterLinearMipmapLinear, driver.FilterLinear)
	if err != nil {
		t.Fatal(err)
	}
	defer tex.Release()
	mask := tex.(*Texture)
	rgba, err := b.NewTexture(driver.TextureFormatSRGBA, sz.X, sz.Y, driver.FilterLinearMipmapLinear, driver.FilterLinear, driver.BufferBindingTexture)
	if err != nil {
		t.Fatal(err)
	}
	defer rgba.Release()
	format := uintptr(objc.ID(mask.texture).Send(objc.RegisterName("pixelFormat")))
	if format != 10 {
		t.Fatalf("mask pixel format %d, want R8Unorm", format)
	}
	maskBytes := uintptr(objc.ID(mask.texture).Send(objc.RegisterName("allocatedSize")))
	rgbaBytes := uintptr(objc.ID(rgba.(*Texture).texture).Send(objc.RegisterName("allocatedSize")))
	t.Logf("256px mask+all mip levels: R8=%d bytes, RGBA=%d bytes", maskBytes, rgbaBytes)
	if maskBytes*3 > rgbaBytes {
		t.Fatal("R8 did not reduce allocated texture storage")
	}
	b.BeginFrame(nil, false, sz)
	stride := sz.X + 3
	src := make([]byte, stride*sz.Y)
	for i := range src {
		src[i] = byte(i*17 + i/13)
	}
	mask.Upload(image.Point{}, sz, src, stride)
	if b.blitEnc != 0 {
		t.Fatal("initial immutable mask used a staging blit")
	}
	mask.Upload(image.Pt(7, 9), image.Pt(1, 1), []byte{241}, 1)
	if b.blitEnc == 0 {
		t.Fatal("later mask writes lost GPU ordering")
	}
	b.endCmdBuffer(true)
	var read func(uintptr, uintptr, unsafe.Pointer, uintptr, uploadRegion, uintptr)
	purego.RegisterFunc(&read, uploadMsg)
	out := make([]byte, sz.X*sz.Y)
	read(uintptr(mask.texture), uintptr(objc.RegisterName("getBytes:bytesPerRow:fromRegion:mipmapLevel:")), unsafe.Pointer(&out[0]), uintptr(sz.X), uploadRegion{Width: 256, Height: 256, Depth: 1}, 0)
	src[9*stride+7] = 241
	for y := range sz.Y {
		for x := range sz.X {
			if out[y*sz.X+x] != src[y*stride+x] {
				t.Fatal("R8 partial upload corrupts pixels or stride")
			}
		}
	}
	b.unified = false
	if tex, err := b.NewMaskTexture(2, 2, driver.FilterNearest, driver.FilterNearest); err == nil {
		tex.Release()
		t.Fatal("unsupported mask did not choose portable fallback")
	}
}
