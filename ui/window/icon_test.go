package window

import (
	"bytes"
	"encoding/binary"
	"github.com/dyike/keel/internal/appicon"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
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
	data := netWMIcon(art, false)
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

func TestRunIconKeepsFinishedShape(t *testing.T) {
	old, finished := currentIconState()
	defer func() { appIcon.Lock(); appIcon.art = old; appIcon.finished = finished; appIcon.Unlock() }()
	path := filepath.Join(t.TempDir(), "finished.png")
	if err := os.WriteFile(path, artPNG(t, 64, 64), 0600); err != nil {
		t.Fatal(err)
	}
	if err := readRunIcon(path); err != nil {
		t.Fatal(err)
	}
	art, done := currentIconState()
	if !done {
		t.Fatal("CLI icon was marked as unshaped artwork")
	}
	for _, shape := range []appicon.Shape{appicon.MacOS, appicon.Windows, appicon.Linux} {
		img := renderIcon(art, shape, 64, done)
		if _, _, _, a := img.At(0, 0).RGBA(); a != 65535 {
			t.Fatalf("%s applied a second margin/mask", shape.Name)
		}
	}
	data := netWMIcon(art, true)
	if binary.LittleEndian.Uint32(data[8:])>>24 != 255 {
		t.Fatal("X11 applied a second template")
	}
	if err := SetIcon(artPNG(t, 64, 64)); err != nil {
		t.Fatal(err)
	}
	_, done = currentIconState()
	if done {
		t.Fatal("explicit SetIcon should restore artwork shaping")
	}
}
