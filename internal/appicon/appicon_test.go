package appicon

import (
	"image"
	"image/color"
	"testing"
)

// solid is full-bleed artwork of one opaque color.
func solid(size int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 30, 120, 220, 255
	}
	return img
}

func alphaAt(img image.Image, x, y int) uint8 {
	return color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA).A
}

func TestIconShapesFollowEachPlatform(t *testing.T) {
	art := solid(1024)
	// macOS: an 824px plate at 100px from each edge.
	mac := MacOS.Render(art, 1024, true)
	if alphaAt(mac, 512, 102) != 255 || alphaAt(mac, 102, 512) != 255 {
		t.Fatal("macOS plate edges")
	}
	if alphaAt(mac, 512, 80) > 40 || alphaAt(mac, 80, 512) > 60 {
		t.Fatal("macOS keeps its 100px gutter (shadow aside)")
	}
	if alphaAt(mac, 512, 940) == 0 || alphaAt(mac, 512, 940) > 160 {
		t.Fatal("macOS shadow falls below the plate", alphaAt(mac, 512, 940))
	}
	// The corner is cut: continuous corners reach further in than the
	// plate's corner point.
	if alphaAt(mac, 110, 110) > 128 {
		t.Fatal("macOS corner not rounded")
	}
	// Continuous corners share the circular arc at the diagonal but ease
	// out of the straight edge earlier and more gently: one pixel inside
	// the left edge, the plate starts lower than with a circular corner.
	circ := Shape{body: MacOS.body, radius: MacOS.radius}.Render(art, 1024, true)
	firstOpaque := func(img image.Image, x int) int {
		for y := 100; y < 512; y++ {
			if alphaAt(img, x, y) >= 128 {
				return y
			}
		}
		return -1
	}
	if c, k := firstOpaque(mac, 101), firstOpaque(circ, 101); c <= k+10 {
		t.Fatalf("continuous corner should ease out of the edge further down: %d vs circular %d", c, k)
	}
	if d1, d2 := diagonalEdge(mac), diagonalEdge(circ); d1-d2 > 2 || d2-d1 > 2 {
		t.Fatalf("diagonal %d vs %d", d1, d2)
	}

	// Windows: a 42/48 plate with 2/48 corners.
	win := Windows.Render(art, 48, true)
	if alphaAt(win, 24, 3) < 200 || alphaAt(win, 3, 24) < 200 {
		t.Fatal("Windows plate edges at 3px")
	}
	if alphaAt(win, 1, 24) != 0 || alphaAt(win, 24, 46) != 0 {
		t.Fatal("Windows margin")
	}
	if alphaAt(win, 3, 3) > 128 {
		t.Fatal("Windows corner not rounded")
	}
	// GNOME: a 104/128 plate with 8px corners.
	lin := Linux.Render(art, 128, true)
	if alphaAt(lin, 64, 12) < 200 || alphaAt(lin, 64, 11) != 0 || alphaAt(lin, 64, 116) != 0 {
		t.Fatal("GNOME plate keyline")
	}
	if alphaAt(lin, 12, 12) > 64 {
		t.Fatal("GNOME corner not rounded")
	}

	// icon_mask none keeps the artwork's own transparency.
	dot := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	dot.Set(50, 50, color.NRGBA{255, 0, 0, 255})
	free := Linux.Render(dot, 128, false)
	if alphaAt(free, 12, 12) != 0 || alphaAt(free, 64, 64) == 0 {
		t.Fatal("unmasked artwork keeps its own outline")
	}
}

func diagonalEdge(img image.Image) int {
	for i := 100; i < 512; i++ {
		if alphaAt(img, i, i) >= 128 {
			return i
		}
	}
	return -1
}
