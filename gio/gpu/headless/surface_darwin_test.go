// SPDX-License-Identifier: Unlicense OR MIT

//go:build !nometal

package headless

import (
	"bytes"
	"image"
	"image/color"
	"testing"
	"unsafe"

	"github.com/dyike/keel/gio/f32"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/ebitengine/purego"
)

// bgraSurface makes an IOSurface holding img's pixels in BGRA order.
func bgraSurface(t *testing.T, img *image.RGBA) (uintptr, func()) {
	t.Helper()
	cf, err := purego.Dlopen("/System/Library/Frameworks/CoreFoundation.framework/CoreFoundation", purego.RTLD_NOW)
	if err != nil {
		t.Skip(err)
	}
	ios, err := purego.Dlopen("/System/Library/Frameworks/IOSurface.framework/IOSurface", purego.RTLD_NOW)
	if err != nil {
		t.Skip(err)
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
	kcb, _ := purego.Dlsym(cf, "kCFTypeDictionaryKeyCallBacks")
	vcb, _ := purego.Dlsym(cf, "kCFTypeDictionaryValueCallBacks")
	w, h := img.Rect.Dx(), img.Rect.Dy()
	var keys, vals []uintptr
	for _, kv := range []struct {
		key string
		v   int32
	}{{"kIOSurfaceWidth", int32(w)}, {"kIOSurfaceHeight", int32(h)}, {"kIOSurfaceBytesPerElement", 4}, {"kIOSurfacePixelFormat", 0x42475241}} {
		p, err := purego.Dlsym(ios, kv.key)
		if err != nil {
			t.Skip(err)
		}
		v := kv.v
		keys = append(keys, **(**uintptr)(unsafe.Pointer(&p)))
		vals = append(vals, numCreate(0, 3, &v))
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
			c := img.RGBAAt(x, y)
			copy(pix[y*stride+x*4:], []byte{c.B, c.G, c.R, c.A})
		}
	}
	unlock(s, 0, 0)
	return s, func() { cfRelease(s) }
}

// A surface draws exactly like the same pixels in an ImageOp: channel
// order, sRGB decoding, orientation and filtering all match.
func TestSurfaceDrawsLikeImage(t *testing.T) {
	w, err := NewWindow(160, 100)
	if err != nil {
		t.Skip(err)
	}
	defer w.Release()
	img := image.NewRGBA(image.Rect(0, 0, 61, 37))
	for y := range 37 {
		for x := range 61 {
			img.SetRGBA(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 6), B: uint8(255 - x*3), A: 255})
		}
	}
	img.SetRGBA(0, 0, color.RGBA{R: 255, A: 255}) // top-left marker
	surf, release := bgraSurface(t, img)
	defer release()
	var s paint.Surface
	s.Set(surf, img.Rect.Size())

	for _, scale := range []float32{1, 1.6, .7} {
		render := func(im paint.ImageOp) []byte {
			var o op.Ops
			paint.Fill(&o, color.NRGBA{R: 9, G: 9, B: 9, A: 255})
			tr := op.Affine(f32.AffineId().Scale(f32.Point{}, f32.Pt(scale, scale)).Offset(f32.Pt(11, 7))).Push(&o)
			im.Filter = paint.FilterLinearNoMipmap
			im.Add(&o)
			paint.PaintOp{}.Add(&o)
			tr.Pop()
			if err := w.Frame(&o); err != nil {
				t.Fatal(err)
			}
			shot := image.NewRGBA(image.Rect(0, 0, 160, 100))
			if err := w.Screenshot(shot); err != nil {
				t.Fatal(err)
			}
			return shot.Pix
		}
		want := render(paint.NewImageOp(img))
		got := render(paint.NewSurfaceOp(&s))
		if !bytes.Equal(got, want) {
			diff := 0
			for i := range got {
				if got[i] != want[i] {
					diff++
				}
			}
			t.Errorf("scale %v: surface differs from image in %d of %d bytes", scale, diff, len(got))
		}
	}
}
