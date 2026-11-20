// SPDX-License-Identifier: Unlicense OR MIT

//go:build android && !nowayland

package vk

import (
	"fmt"
	"unsafe"
)

// The Android surface of Vulkan without cgo. Keel patch.

const _VK_STRUCTURE_TYPE_ANDROID_SURFACE_CREATE_INFO_KHR vkStructureType = 1000008000

type vkAndroidSurfaceCreateInfoKHR struct {
	sType  vkStructureType
	pNext  unsafe.Pointer
	flags  uint32
	window unsafe.Pointer
}

var androidFuncs struct {
	vkCreateAndroidSurfaceKHR uintptr
}

func init() {
	loadFuncs = append(loadFuncs, func(dlopen func(name string) uintptr) {
		androidFuncs.vkCreateAndroidSurfaceKHR = dlopen("vkCreateAndroidSurfaceKHR")
	})
}

func CreateAndroidSurface(inst Instance, window unsafe.Pointer) (Surface, error) {
	inf := vkAndroidSurfaceCreateInfoKHR{
		sType:  _VK_STRUCTURE_TYPE_ANDROID_SURFACE_CREATE_INFO_KHR,
		window: window,
	}
	var surf Surface
	if err := vkErr(vkCreateAndroidSurfaceKHR(androidFuncs.vkCreateAndroidSurfaceKHR, inst, &inf, nil, &surf)); err != nil {
		return 0, fmt.Errorf("vulkan: vkCreateAndroidSurfaceKHR: %w", err)
	}
	return surf, nil
}

func vkCreateAndroidSurfaceKHR(f uintptr, instance vkInstance, pCreateInfo *vkAndroidSurfaceCreateInfoKHR, pAllocator unsafe.Pointer, pSurface *vkSurfaceKHR) vkResult {
	return vkResult(int32(vkCall(f, uintptr(instance), uintptr(unsafe.Pointer(pCreateInfo)), uintptr(pAllocator), uintptr(unsafe.Pointer(pSurface)))))
}
