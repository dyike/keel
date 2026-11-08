// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package surface

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestDMABufProducerAndGPUOwnership(t *testing.T) {
	var frees atomic.Int32
	h, err := RegisterDMABuf(DMABuf{Width: 64, Height: 48, FourCC: 0x3231564e, Planes: []DMABufPlane{{FD: 10, Pitch: 256, Modifier: 0x0200000018601b04}}}, func() {
		frees.Add(1)
		// Callbacks run outside the registry lock.
		if RetainDMABuf(hypotheticalExpiredHandle()) {
			t.Error("expired handle was retained")
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	_, gpuDone, ok := AcquireDMABuf(h)
	if !ok {
		t.Fatal("cannot acquire live surface")
	}
	ReleaseDMABuf(h)
	if frees.Load() != 0 {
		t.Fatal("producer buffer was returned before the GPU completed")
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 100 {
				if RetainDMABuf(h) {
					ReleaseDMABuf(h)
				}
			}
		}()
	}
	wg.Wait()
	gpuDone()
	if frees.Load() != 1 || RetainDMABuf(h) {
		t.Fatal("final GPU release did not retire the buffer exactly once")
	}
}

func hypotheticalExpiredHandle() uintptr { return ^uintptr(0) }

func TestDMABufRejectsBadMetadataWithoutTakingOwnership(t *testing.T) {
	for _, img := range []DMABuf{{}, {Width: 1, Height: 1}, {Width: 1, Height: 1, Planes: []DMABufPlane{{FD: -1, Pitch: 1}}}} {
		if _, err := RegisterDMABuf(img, func() { t.Error("invalid image consumed producer ownership") }); err == nil {
			t.Fatal("invalid image registered")
		}
	}
}
