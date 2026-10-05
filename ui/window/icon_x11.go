package window

import (
	"encoding/binary"
	"image"
	"image/color"

	"github.com/dyike/keel/internal/appicon"
)

// x11IconSizes are drawn on GNOME's plate; window managers pick the
// closest and scale.
var x11IconSizes = []int{16, 32, 48, 64, 128}

// netWMIcon encodes _NET_WM_ICON: for each size, width, height, then rows
// of non-premultiplied ARGB, as 32-bit values in the connection's (little
// endian) byte order.
func netWMIcon(art image.Image) []byte {
	var out []byte
	put := func(v uint32) { out = binary.LittleEndian.AppendUint32(out, v) }
	for _, size := range x11IconSizes {
		img := appicon.Linux.Render(art, size, true)
		put(uint32(size))
		put(uint32(size))
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				c := img.NRGBAAt(x, y)
				put(argb(c))
			}
		}
	}
	return out
}

func argb(c color.NRGBA) uint32 {
	return uint32(c.A)<<24 | uint32(c.R)<<16 | uint32(c.G)<<8 | uint32(c.B)
}
