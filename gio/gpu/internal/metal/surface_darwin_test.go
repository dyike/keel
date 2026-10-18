// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"bytes"
	"testing"
	"unsafe"

	"github.com/dyike/keel/gio/gpu/internal/driver"
	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// newTestSurface makes a w×h BGRA IOSurface filled by fill.
func newTestSurface(t *testing.T, w, h int, fill func(x, y int) [4]byte) (surface uintptr, release func()) {
	t.Helper()
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW)
	if err != nil {
		t.Fatal(err)
	}
	ios, err := purego.Dlopen("/System/Library/Frameworks/IOSurface.framework/IOSurface", purego.RTLD_NOW)
	if err != nil {
		t.Fatal(err)
	}
	var (
		cfRelease  func(uintptr)
		numCreate  func(uintptr, int, *int32) uintptr
		dictCreate func(uintptr, *uintptr, *uintptr, int, uintptr, uintptr) uintptr
		create     func(uintptr) uintptr
		lock       func(uintptr, uint32, uintptr) int32
		unlock     func(uintptr, uint32, uintptr) int32
		base       func(uintptr) unsafe.Pointer
		rowBytes   func(uintptr) uintptr
	)
	purego.RegisterLibFunc(&cfRelease, cf, "CFRelease")
	purego.RegisterLibFunc(&numCreate, cf, "CFNumberCreate")
	purego.RegisterLibFunc(&dictCreate, cf, "CFDictionaryCreate")
	purego.RegisterLibFunc(&create, ios, "IOSurfaceCreate")
	purego.RegisterLibFunc(&lock, ios, "IOSurfaceLock")
	purego.RegisterLibFunc(&unlock, ios, "IOSurfaceUnlock")
	purego.RegisterLibFunc(&base, ios, "IOSurfaceGetBaseAddress")
	purego.RegisterLibFunc(&rowBytes, ios, "IOSurfaceGetBytesPerRow")
	global := func(lib uintptr, name string) uintptr {
		p, err := purego.Dlsym(lib, name)
		if err != nil {
			t.Fatal(err)
		}
		return **(**uintptr)(unsafe.Pointer(&p))
	}
	kcb, _ := purego.Dlsym(cf, "kCFTypeDictionaryKeyCallBacks")
	vcb, _ := purego.Dlsym(cf, "kCFTypeDictionaryValueCallBacks")
	var keys, vals []uintptr
	for _, kv := range []struct {
		key string
		v   int32
	}{{"kIOSurfaceWidth", int32(w)}, {"kIOSurfaceHeight", int32(h)}, {"kIOSurfaceBytesPerElement", 4}, {"kIOSurfacePixelFormat", 0x42475241}} {
		v := kv.v
		keys = append(keys, global(ios, kv.key))
		vals = append(vals, numCreate(0, 3, &v)) // kCFNumberSInt32Type
	}
	props := dictCreate(0, &keys[0], &vals[0], len(keys), kcb, vcb)
	for _, v := range vals {
		cfRelease(v)
	}
	defer cfRelease(props)
	s := create(props)
	if s == 0 {
		t.Skip("IOSurfaceCreate failed")
	}
	lock(s, 0, 0)
	stride := int(rowBytes(s))
	pix := unsafe.Slice((*byte)(base(s)), stride*h)
	for y := range h {
		for x := range w {
			c := fill(x, y)
			copy(pix[y*stride+x*4:], c[:])
		}
	}
	unlock(s, 0, 0)
	return s, func() { cfRelease(s) }
}

func TestSurfaceTextureSharesPixels(t *testing.T) {
	b := testBackend(t)
	if !b.unified {
		t.Skip("reads back shared storage; this GPU has its own memory")
	}
	const w, h = 37, 21
	fill := func(x, y int) [4]byte { return [4]byte{byte(x * 7), byte(y * 11), byte(x + y), 255} }
	s, release := newTestSurface(t, w, h, fill)
	defer release()
	v, err := b.NewSurfaceTexture(s, w, h, driver.FilterLinear, driver.FilterLinear)
	if err != nil {
		t.Fatal(err)
	}
	tex := v.(*Texture)
	defer tex.Release()
	var read func(uintptr, uintptr, unsafe.Pointer, uintptr, uploadRegion, uintptr)
	purego.RegisterFunc(&read, uploadMsg)
	got := make([]byte, w*h*4)
	read(uintptr(tex.texture), uintptr(objc.RegisterName("getBytes:bytesPerRow:fromRegion:mipmapLevel:")),
		unsafe.Pointer(&got[0]), w*4, uploadRegion{0, 0, 0, w, h, 1}, 0)
	for y := range h {
		for x := range w {
			want := fill(x, y)
			if i := (y*w + x) * 4; !bytes.Equal(got[i:i+4], want[:]) {
				t.Fatalf("texel (%d,%d) = % x, want % x", x, y, got[i:i+4], want)
			}
		}
	}
}
