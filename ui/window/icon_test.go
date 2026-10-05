package window

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func artPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = 20, 140, 220, 255
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestSetIconValidatesAndKeepsArtwork(t *testing.T) {
	defer func() { appIcon.Lock(); appIcon.art = nil; appIcon.Unlock() }()
	if SetIcon([]byte("not a png")) == nil {
		t.Fatal("accepted non-PNG data")
	}
	if SetIcon(artPNG(t, 64, 32)) == nil {
		t.Fatal("accepted non-square artwork")
	}
	t.Setenv("KEEL_AUTOMATION", "1")
	t.Setenv("KEEL_HEADLESS", "1") // off screen: store only, no platform call
	if err := SetIcon(artPNG(t, 256, 256)); err != nil {
		t.Fatal(err)
	}
	if art := currentIcon(); art == nil || art.Bounds().Dx() != 256 {
		t.Fatal("artwork not kept for windows opened later")
	}
}

func TestNetWMIconLayout(t *testing.T) {
	art, _ := png.Decode(bytes.NewReader(artPNG(t, 512, 512)))
	data := netWMIcon(art)
	want := 0
	for _, s := range x11IconSizes {
		want += 2 + s*s
	}
	if len(data) != want*4 {
		t.Fatalf("%d bytes, want %d", len(data), want*4)
	}
	word := func(i int) uint32 { return binary.LittleEndian.Uint32(data[i*4:]) }
	if word(0) != 16 || word(1) != 16 {
		t.Fatal("first image header", word(0), word(1))
	}
	// 16px on GNOME's plate: corners transparent, the middle opaque blue.
	corner, middle := word(2), word(2+8*16+8)
	if corner>>24 != 0 {
		t.Fatalf("corner pixel %08x should be transparent", corner)
	}
	if middle>>24 != 0xff || argb(color.NRGBA{20, 140, 220, 255}) != middle {
		t.Fatalf("middle pixel %08x", middle)
	}
	next := 2 + 16*16
	if word(next) != 32 || word(next+1) != 32 {
		t.Fatal("second image header")
	}
}
