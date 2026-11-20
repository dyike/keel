// SPDX-License-Identifier: Unlicense OR MIT

//go:build android && !nowayland

package vk

import (
	"testing"
	"unsafe"
)

func TestAndroidSurfaceCreateInfoLayout(t *testing.T) {
	var info vkAndroidSurfaceCreateInfoKHR
	pointerSize := unsafe.Sizeof(uintptr(0))
	if unsafe.Sizeof(info.sType) != 4 || unsafe.Sizeof(info.flags) != 4 {
		t.Fatal("Vulkan structure type and flags must be 32-bit")
	}
	if unsafe.Offsetof(info.pNext) != pointerSize || unsafe.Offsetof(info.flags) != 2*pointerSize || unsafe.Offsetof(info.window) != 3*pointerSize || unsafe.Sizeof(info) != 4*pointerSize {
		t.Fatalf("invalid Android surface ABI: pNext=%d flags=%d window=%d size=%d", unsafe.Offsetof(info.pNext), unsafe.Offsetof(info.flags), unsafe.Offsetof(info.window), unsafe.Sizeof(info))
	}
}
