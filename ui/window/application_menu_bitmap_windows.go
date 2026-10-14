//go:build windows

package window

import (
	"bytes"
	"fmt"
	"image/png"
	"math"
	"unsafe"

	"golang.org/x/sys/windows"
)

var menuGDI = windows.NewLazySystemDLL("gdi32.dll")

func createMenuBitmap(data []byte, side int) (uintptr, error) {
	art, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	size := art.Bounds().Size()
	if size.X < 1 || size.Y < 1 || size.X > 256 || size.Y > 256 {
		return 0, fmt.Errorf("window: invalid menu bitmap dimensions")
	}
	// Top-down, premultiplied BGRA, retained as an HBITMAP by the menu owner.
	info := struct {
		Size                   uint32
		Width, Height          int32
		Planes, Bits           uint16
		Compression, ImageSize uint32
		X, Y                   int32
		Colors, Important      uint32
	}{Size: 40, Width: int32(size.X), Height: -int32(size.Y), Planes: 1, Bits: 32}
	sourceSize := size
	ratio := math.Min(float64(side)/float64(size.X), float64(side)/float64(size.Y))
	width, height := max(1, int(float64(size.X)*ratio)), max(1, int(float64(size.Y)*ratio))
	info.Width, info.Height = int32(side), -int32(side)
	var pixels unsafe.Pointer // DIB memory owned by the bitmap, not by Go
	bitmap, _, err := menuGDI.NewProc("CreateDIBSection").Call(0, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&pixels)), 0, 0)
	if bitmap == 0 {
		return 0, fmt.Errorf("window: CreateDIBSection: %v", err)
	}
	dst := unsafe.Slice((*uint32)(pixels), side*side)
	clear(dst) // Transparent padding around nonsquare artwork.
	for y := range side {
		for x := range side {
			if x < (side-width)/2 || x >= (side-width)/2+width || y < (side-height)/2 || y >= (side-height)/2+height {
				continue
			}
			sx := (x - (side-width)/2) * sourceSize.X / width
			sy := (y - (side-height)/2) * sourceSize.Y / height
			r, g, b, a := art.At(art.Bounds().Min.X+sx, art.Bounds().Min.Y+sy).RGBA()
			dst[y*side+x] = a>>8<<24 | r>>8<<16 | g>>8<<8 | b>>8
		}
	}
	return bitmap, nil
}

func setMenuBitmap(menu uintptr, position uint32, bitmap uintptr) bool {
	info := struct {
		Size, Mask, Type, State, ID       uint32
		Submenu, Checked, Unchecked, Data uintptr
		Text                              *uint16
		Length                            uint32
		Bitmap                            uintptr
	}{Mask: 0x80, Bitmap: bitmap} // MIIM_BITMAP; preserves the checkmark column.
	info.Size = uint32(unsafe.Sizeof(info))
	result, _, _ := user32.NewProc("SetMenuItemInfoW").Call(menu, uintptr(position), 1, uintptr(unsafe.Pointer(&info)))
	return result != 0
}

func freeMenuBitmaps(bitmaps []uintptr) {
	for _, bitmap := range bitmaps {
		menuGDI.NewProc("DeleteObject").Call(bitmap)
	}
}
