package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/dyike/keel/internal/appicon"
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

func TestIconCommandAndOverrides(t *testing.T) {
	root := t.TempDir()
	c, out := newCLI(root)
	if code := c.main([]string{"new", "app", "-offline"}); code != 0 {
		t.Fatal(out.String())
	}
	dir := filepath.Join(root, "app")
	c, out = newCLI(dir)
	if code := c.main([]string{"icon"}); code != 0 {
		t.Fatal(out.String())
	}
	icons := filepath.Join(dir, "dist", "icons")
	for _, f := range []string{"macos.png", "windows.ico", "windows/16.png", "windows/256.png", "linux/48.png", "linux/512.png"} {
		if _, err := os.Stat(filepath.Join(icons, f)); err != nil {
			t.Fatal("missing", f)
		}
	}
	ico, _ := os.ReadFile(filepath.Join(icons, "windows.ico"))
	var header struct{ Reserved, Type, Count uint16 }
	binary.Read(bytes.NewReader(ico), binary.LittleEndian, &header)
	if header.Type != 1 || int(header.Count) != len(appicon.WindowsSizes) {
		t.Fatalf("ico holds %d images, want %d", header.Count, len(appicon.WindowsSizes))
	}

	// A finished per-platform icon is used as it is.
	cfg, _ := loadConfig(dir)
	if err := writePNG(filepath.Join(dir, "mac-final.png"), solid(64)); err != nil {
		t.Fatal(err)
	}
	cfg.Icons = map[string]string{"darwin": "mac-final.png"}
	set, err := loadIcons(dir, cfg)
	if err != nil {
		t.Fatal(err)
	}
	img, _ := set.icon("darwin", 1024)
	if alphaAt(img, 2, 2) != 255 {
		t.Fatal("override not used as is")
	}
	cfg.Icons = map[string]string{"beos": "x.png"}
	if cfg.validate() == nil {
		t.Fatal("unknown platform in icons accepted")
	}
}
