// SPDX-License-Identifier: Unlicense OR MIT

package metal

import (
	"runtime"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// Textures over IOSurfaces, through purego. Keel patch.

const (
	pixelFormatBGRA8sRGB = 81 // MTLPixelFormatBGRA8Unorm_sRGB
	textureUsageRead     = 1  // MTLTextureUsageShaderRead
)

var surfaceSel uintptr

// surfaceTexture returns a +1 texture sharing the IOSurface's memory, or 0.
func surfaceTexture(dev, surface, width, height uintptr, unified bool) uintptr {
	uploadInit()
	if !uploadReady {
		return 0
	}
	if surfaceSel == 0 {
		surfaceSel = uintptr(objc.RegisterName("newTextureWithDescriptor:iosurface:plane:"))
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool, _, _ := purego.SyscallN(uploadPoolPush)
	defer purego.SyscallN(uploadPoolPop, pool)
	desc := uploadSend(uploadDescriptorClass, "new")
	defer uploadSend(desc, "release")
	uploadSend(desc, "setPixelFormat:", pixelFormatBGRA8sRGB)
	uploadSend(desc, "setWidth:", width)
	uploadSend(desc, "setHeight:", height)
	uploadSend(desc, "setUsage:", textureUsageRead)
	storage := uintptr(1) // managed: the surface's memory is the CPU's
	if unified {
		storage = 0 // shared
	}
	uploadSend(desc, "setStorageMode:", storage)
	uploadSend(desc, "setMipmapLevelCount:", 1)
	tex, _, _ := purego.SyscallN(uploadMsg, dev, surfaceSel, desc, surface, 0)
	return tex
}
