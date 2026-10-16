// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"image"
	"math"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

type uploadRegion struct{ X, Y, Z, Width, Height, Depth uintptr }

var uploadOnce sync.Once
var uploadMsg uintptr
var uploadPoolPush, uploadPoolPop, uploadDescriptorClass uintptr
var uploadReady bool
var uploadSelectors map[string]uintptr
var uploadReplace func(uintptr, uintptr, uploadRegion, uintptr, unsafe.Pointer, uintptr)
var uploadReplaceSel uintptr

func uploadInit() {
	uploadOnce.Do(func() {
		var err error
		uploadMsg, err = purego.Dlsym(purego.RTLD_DEFAULT, "objc_msgSend")
		if err != nil {
			return
		}
		uploadPoolPush, err = purego.Dlsym(purego.RTLD_DEFAULT, "objc_autoreleasePoolPush")
		if err != nil {
			return
		}
		uploadPoolPop, err = purego.Dlsym(purego.RTLD_DEFAULT, "objc_autoreleasePoolPop")
		if err != nil {
			return
		}
		uploadDescriptorClass = uintptr(objc.GetClass("MTLTextureDescriptor"))
		if uploadDescriptorClass == 0 {
			return
		}
		purego.RegisterFunc(&uploadReplace, uploadMsg)
		uploadReplaceSel = uintptr(objc.RegisterName("replaceRegion:mipmapLevel:withBytes:bytesPerRow:"))
		uploadSelectors = make(map[string]uintptr)
		for _, name := range []string{"new", "release", "setPixelFormat:", "setWidth:", "setHeight:", "setUsage:", "setStorageMode:", "setMipmapLevelCount:", "newTextureWithDescriptor:", "hasUnifiedMemory", "respondsToSelector:", "setSwizzle:"} {
			uploadSelectors[name] = uintptr(objc.RegisterName(name))
		}
		uploadReady = true
	})
}
func uploadSend(id uintptr, name string, args ...uintptr) uintptr {
	a := []uintptr{id, uploadSelectors[name]}
	a = append(a, args...)
	v, _, _ := purego.SyscallN(uploadMsg, a...)
	return v
}
func uploadTexture(dev uintptr, format, width, height, usage uintptr, mipmap bool) uintptr {
	uploadInit()
	if !uploadReady {
		return 0
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool, _, _ := purego.SyscallN(uploadPoolPush)
	defer purego.SyscallN(uploadPoolPop, pool)
	desc := uploadSend(uploadDescriptorClass, "new")
	defer uploadSend(desc, "release")
	if format == 10 { // MTLPixelFormatR8Unorm
		if uploadSend(desc, "respondsToSelector:", uploadSelectors["setSwizzle:"]) == 0 {
			return 0
		}
		// MTLTextureSwizzleChannels is four uint8_t fields. Replicate the
		// linear red channel to premultiplied white RGBA for every shader.
		uploadSend(desc, "setSwizzle:", 0x02020202)
	}
	uploadSend(desc, "setPixelFormat:", format)
	uploadSend(desc, "setWidth:", width)
	uploadSend(desc, "setHeight:", height)
	uploadSend(desc, "setUsage:", usage)
	uploadSend(desc, "setStorageMode:", 0)
	levels := uintptr(1)
	if mipmap {
		for n := max(width, height); n > 1; n /= 2 {
			levels++
		}
	}
	uploadSend(desc, "setMipmapLevelCount:", levels)
	return uploadSend(dev, "newTextureWithDescriptor:", desc)
}

// Coverage has one linear byte, including in its mip chain. No gamma LUT,
// RGBA expansion or retained upload copy is needed.
func uploadMaskMipmaps(tex uintptr, size image.Point, pixels []byte, stride int) {
	for level := 1; size.X > 1 || size.Y > 1; level++ {
		next := image.Pt(max(1, size.X/2), max(1, size.Y/2))
		dst := make([]byte, next.X*next.Y)
		for y := 0; y < next.Y; y++ {
			fy := (float64(y)+.5)*float64(size.Y)/float64(next.Y) - .5
			y0 := int(fy)
			dy := fy - float64(y0)
			for x := 0; x < next.X; x++ {
				fx := (float64(x)+.5)*float64(size.X)/float64(next.X) - .5
				x0 := int(fx)
				dx := fx - float64(x0)
				v := float64(pixels[y0*stride+x0])*(1-dx)*(1-dy) + float64(pixels[y0*stride+min(x0+1, size.X-1)])*dx*(1-dy) + float64(pixels[min(y0+1, size.Y-1)*stride+x0])*(1-dx)*dy + float64(pixels[min(y0+1, size.Y-1)*stride+min(x0+1, size.X-1)])*dx*dy
				dst[y*next.X+x] = byte(math.RoundToEven(v))
			}
		}
		directUpload(tex, image.Point{}, next, dst, next.X, level)
		pixels, stride, size = dst, next.X, next
	}
}
func directUpload(tex uintptr, offset, size image.Point, pixels []byte, stride int, level int) {
	uploadInit()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool, _, _ := purego.SyscallN(uploadPoolPush)
	defer purego.SyscallN(uploadPoolPop, pool)
	uploadReplace(tex, uploadReplaceSel, uploadRegion{uintptr(offset.X), uintptr(offset.Y), 0, uintptr(size.X), uintptr(size.Y), 1}, uintptr(level), unsafe.Pointer(&pixels[0]), uintptr(stride))
}

var mipGammaOnce sync.Once
var mipDecode [256]float64
var mipEncode [65536]byte

func initMipGamma() {
	mipGammaOnce.Do(func() {
		for i := range mipDecode {
			q := float64(i) / 255
			if q <= .04045 {
				mipDecode[i] = q / 12.92
			} else {
				mipDecode[i] = math.Pow((q+.055)/1.055, 2.4)
			}
		}
		for i := range mipEncode {
			q := float64(i) / 65535
			if q <= .0031308 {
				q *= 12.92
			} else {
				q = 1.055*math.Pow(q, 1/2.4) - .055
			}
			mipEncode[i] = byte(math.RoundToEven(q * 255))
		}
	})
}

// uploadMipmaps preserves linear-light sRGB filtering and every mip level.
// The source is only read during the initial upload; no CPU image cache is retained.
func uploadMipmaps(tex uintptr, size image.Point, pixels []byte, stride int, srgb bool) {
	if srgb {
		initMipGamma()
	}
	for level := 1; size.X > 1 || size.Y > 1; level++ {
		next := image.Pt(max(1, size.X/2), max(1, size.Y/2))
		dst := downsampleMip(size, next, pixels, stride, srgb)
		directUpload(tex, image.Point{}, next, dst, next.X*4, level)
		pixels, stride, size = dst, next.X*4, next
	}
}
func downsampleMip(size, next image.Point, pixels []byte, stride int, srgb bool) []byte {
	dst := make([]byte, next.X*next.Y*4)
	for y := 0; y < next.Y; y++ {
		fy := (float64(y)+.5)*float64(size.Y)/float64(next.Y) - .5
		y0 := int(fy)
		dy := fy - float64(y0)
		for x := 0; x < next.X; x++ {
			fx := (float64(x)+.5)*float64(size.X)/float64(next.X) - .5
			x0 := int(fx)
			dx := fx - float64(x0)
			offsets := [4]int{y0*stride + x0*4, y0*stride + min(x0+1, size.X-1)*4, min(y0+1, size.Y-1)*stride + x0*4, min(y0+1, size.Y-1)*stride + min(x0+1, size.X-1)*4}
			weights := [4]float64{(1 - dx) * (1 - dy), dx * (1 - dy), (1 - dx) * dy, dx * dy}
			for c := 0; c < 4; c++ {
				v := 0.
				for i, off := range offsets {
					q := float64(pixels[off+c]) / 255
					if srgb && c < 3 {
						q = mipDecode[pixels[off+c]]
					}
					v += weights[i] * q
				}
				if srgb && c < 3 {
					dst[(y*next.X+x)*4+c] = mipEncode[min(65535, max(0, int(math.RoundToEven(v*65535))))]
				} else {
					dst[(y*next.X+x)*4+c] = byte(math.RoundToEven(v * 255))
				}
			}
		}
	}
	return dst
}
func unifiedMemory(dev uintptr) bool {
	uploadInit()
	return uploadReady && uploadSend(dev, "respondsToSelector:", uploadSelectors["hasUnifiedMemory"]) != 0 && uploadSend(dev, "hasUnifiedMemory") != 0
}
