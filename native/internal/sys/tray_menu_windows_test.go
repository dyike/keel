//go:build windows

package sys

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"testing"
	"unsafe"
)

func TestTrayMenuBitmapPixelsAndLifetime(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	art := image.NewNRGBA(image.Rect(0, 0, 32, 16))
	for y := range 16 {
		for x := range 32 {
			art.SetNRGBA(x, y, color.NRGBA{R: 200, G: 100, B: 50, A: 128})
		}
	}
	var data bytes.Buffer
	if err := png.Encode(&data, art); err != nil {
		t.Fatal(err)
	}
	bitmap, err := createTrayMenuBitmap(data.Bytes(), 16)
	if err != nil {
		t.Fatal(err)
	}
	// DIBSECTION starts with BITMAP; its bits expose the exact BGRA mask.
	var object struct {
		Type, Width, Height, WidthBytes int32
		Planes, Bits                    uint16
		Pixels                          unsafe.Pointer
	}
	if n, _, _ := gdi32.NewProc("GetObjectW").Call(bitmap, unsafe.Sizeof(object), uintptr(unsafe.Pointer(&object))); n == 0 {
		freeTrayMenuBitmaps([]uintptr{bitmap})
		t.Fatal("GetObjectW")
	}
	pixels := unsafe.Slice((*uint32)(object.Pixels), 16*16)
	if object.Width != 16 || object.Height != 16 || pixels[0] != 0 || pixels[8*16+8] != 0x80643219 {
		freeTrayMenuBitmaps([]uintptr{bitmap})
		t.Fatalf("bitmap size, aspect ratio or premultiplication: %+v", object)
	}
	freeTrayMenuBitmaps([]uintptr{bitmap})
	if kind, _, _ := gdi32.NewProc("GetObjectType").Call(bitmap); kind != 0 {
		t.Fatal("bitmap retained after deletion")
	}
	menu, bitmaps, err := makeWindowsTrayMenu([]TrayMenuItem{{ID: 1, Title: "Checked icon", PNG: data.Bytes(), Checked: true}})
	if err != nil || menu == 0 || len(bitmaps) != 1 {
		t.Fatal("menu bitmap allocation", err)
	}
	trayDestroyMenu.Call(menu)
	freeTrayMenuBitmaps(bitmaps)
	// A later malformed icon must release already-created menus and bitmaps.
	resources := user32.NewProc("GetGuiResources")
	before, _, _ := resources.Call(^uintptr(0), 0)
	for range 20 {
		menu, bitmaps, err = makeWindowsTrayMenu([]TrayMenuItem{{ID: 1, Title: "Valid", PNG: data.Bytes()}, {ID: 2, Title: "Invalid", PNG: []byte("invalid")}})
		if err == nil || menu != 0 || len(bitmaps) != 0 {
			t.Fatal("partial menu escaped failed build")
		}
	}
	if after, _, _ := resources.Call(^uintptr(0), 0); after != before {
		t.Fatalf("failed builds leaked GDI resources: %d -> %d", before, after)
	}
}
