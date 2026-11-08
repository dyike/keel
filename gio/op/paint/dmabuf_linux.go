// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android

package paint

import "github.com/dyike/keel/gio/internal/surface"

// DMABuf and DMABufPlane describe a Linux GPU image and its DRM planes.
type DMABuf = surface.DMABuf
type DMABufPlane = surface.DMABufPlane

// RegisterDMABuf returns a Surface.Set handle. release must keep the file
// descriptors and their producer buffer valid until the last GPU consumer
// completes. Ownership passes to the registry only on success.
func RegisterDMABuf(image DMABuf, release func()) (uintptr, error) {
	return surface.RegisterDMABuf(image, release)
}

func RetainDMABuf(handle uintptr) bool { return surface.RetainDMABuf(handle) }
func ReleaseDMABuf(handle uintptr)     { surface.ReleaseDMABuf(handle) }

// DMABufSupported reports whether a live renderer can import Linux GPU images.
func DMABufSupported() bool { return surface.DMABufSupported() }
