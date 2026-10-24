// SPDX-License-Identifier: Unlicense OR MIT

//go:build linux && !android && !nowayland && (!cgo || purego)

package vk

// The Wayland surface of vulkan_wayland.go without cgo. Keel patch.

import (
	"fmt"
	"unsafe"
)

var wlFuncs struct {
	vkCreateWaylandSurfaceKHR uintptr
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) uintptr) {
		wlFuncs.vkCreateWaylandSurfaceKHR = dlopen("vkCreateWaylandSurfaceKHR")
	})
}

func CreateWaylandSurface(inst Instance, disp unsafe.Pointer, wlSurf unsafe.Pointer) (Surface, error) {
	inf := vkWaylandSurfaceCreateInfoKHR{
		sType:   _VK_STRUCTURE_TYPE_WAYLAND_SURFACE_CREATE_INFO_KHR,
		display: disp,
		surface: wlSurf,
	}
	var surf Surface
	if err := vkErr(vkCreateWaylandSurfaceKHR(wlFuncs.vkCreateWaylandSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateWaylandSurfaceKHR: %w", err)
	}
	return surf, nil
}

func vkCreateWaylandSurfaceKHR(f uintptr, instance vkInstance, pCreateInfo *vkWaylandSurfaceCreateInfoKHR, pAllocator unsafe.Pointer, pSurface *vkSurfaceKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(instance), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSurface)))))
}
