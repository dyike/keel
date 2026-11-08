// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package surface

import (
	"errors"
	"sync"
	"sync/atomic"
)

// DMABuf describes a borrowed Linux image. Each plane includes its DRM
// modifier; a tiled image must never be interpreted as linear memory.
type DMABuf struct {
	Width, Height int
	FourCC        uint32
	Planes        []DMABufPlane
	// EGL YUV hints: 601/709/2020, and full/limited range, respectively.
	ColorSpace uint32
	FullRange  bool
	OnError    func(error)
}

type DMABufPlane struct {
	FD       int
	Offset   uint32
	Pitch    uint32
	Modifier uint64
}

type dmaBuffer struct {
	image   DMABuf
	refs    int
	release func()
}

var dmaBuffers = struct {
	sync.Mutex
	next uintptr
	all  map[uintptr]*dmaBuffer
}{all: make(map[uintptr]*dmaBuffer)}

var dmaDevices atomic.Int32

// RegisterDMABuf takes ownership of release, called after the owner and all
// renderer references are gone. No pixels are mapped or copied.
func RegisterDMABuf(image DMABuf, release func()) (uintptr, error) {
	if image.Width <= 0 || image.Height <= 0 || image.Width > 16384 || image.Height > 16384 || int64(image.Width)*int64(image.Height) > 64*1024*1024 || len(image.Planes) < 1 || len(image.Planes) > 4 {
		return 0, errors.New("surface: invalid DMA-BUF dimensions or plane count")
	}
	for _, p := range image.Planes {
		if p.FD < 0 || p.Pitch == 0 || p.Pitch > 1<<30 || p.Offset > 1<<31-1 {
			return 0, errors.New("surface: invalid DMA-BUF plane")
		}
	}
	image.Planes = append([]DMABufPlane(nil), image.Planes...)
	dmaBuffers.Lock()
	dmaBuffers.next++
	handle := dmaBuffers.next
	dmaBuffers.all[handle] = &dmaBuffer{image: image, refs: 1, release: release}
	dmaBuffers.Unlock()
	return handle, nil
}

func RetainDMABuf(handle uintptr) bool {
	dmaBuffers.Lock()
	defer dmaBuffers.Unlock()
	if b := dmaBuffers.all[handle]; b != nil {
		b.refs++
		return true
	}
	return false
}

func ReleaseDMABuf(handle uintptr) {
	var release func()
	dmaBuffers.Lock()
	if b := dmaBuffers.all[handle]; b != nil {
		b.refs--
		if b.refs == 0 {
			delete(dmaBuffers.all, handle)
			release = b.release
		}
	}
	dmaBuffers.Unlock()
	if release != nil {
		release()
	}
}

// AcquireDMABuf keeps both the FDs and the decoder's buffer alive. The
// renderer releases this reference only after its GPU fence completes.
func AcquireDMABuf(handle uintptr) (DMABuf, func(), bool) {
	dmaBuffers.Lock()
	b := dmaBuffers.all[handle]
	if b == nil {
		dmaBuffers.Unlock()
		return DMABuf{}, nil, false
	}
	b.refs++
	img := b.image
	dmaBuffers.Unlock()
	return img, func() { ReleaseDMABuf(handle) }, true
}

func DMABufSupported() bool { return dmaDevices.Load() > 0 }

// EnableDMABuf is called only by a renderer with the necessary extensions.
func EnableDMABuf() func() {
	dmaDevices.Add(1)
	return func() { dmaDevices.Add(-1) }
}
