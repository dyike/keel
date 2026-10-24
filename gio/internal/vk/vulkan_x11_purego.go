// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nox11 && (!cgo || purego)

package vk

// The Xlib surface of vulkan_x11.go without cgo. Keel patch.

import (
	"fmt"
	"unsafe"
)

var x11Funcs struct {
	vkCreateXlibSurfaceKHR uintptr
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) uintptr) {
		x11Funcs.vkCreateXlibSurfaceKHR = dlopen("vkCreateXlibSurfaceKHR")
	})
}

func CreateXlibSurface(inst Instance, dpy unsafe.Pointer, window uintptr) (Surface, error) {
	inf := vkXlibSurfaceCreateInfoKHR{
		sType:  _VK_STRUCTURE_TYPE_XLIB_SURFACE_CREATE_INFO_KHR,
		dpy:    dpy,
		window: uint64(window),
	}
	var surf Surface
	if err := vkErr(vkCreateXlibSurfaceKHR(x11Funcs.vkCreateXlibSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateXlibSurfaceKHR: %w", err)
	}
	return surf, nil
}

func vkCreateXlibSurfaceKHR(f uintptr, instance vkInstance, pCreateInfo *vkXlibSurfaceCreateInfoKHR, pAllocator unsafe.Pointer, pSurface *vkSurfaceKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(instance), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSurface)))))
}
