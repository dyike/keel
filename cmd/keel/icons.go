package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"

	"github.com/tc-hib/winres"
	"golang.org/x/image/draw"
)

// iconSet makes a project's icons, per platform and size.
type iconSet struct {
	dir string
	cfg *Config
	art image.Image
}

func loadIcons(dir string, cfg *Config) (*iconSet, error) {
	art, err := decodePNG(filepath.Join(dir, cfg.Icon))
	if err != nil {
		return nil, err
	}
	b := art.Bounds()
	if b.Dx() != b.Dy() {
		return nil, fmt.Errorf("icon %s is %dx%d; it must be square", cfg.Icon, b.Dx(), b.Dy())
	}
	if b.Dx() < 512 {
		fmt.Fprintf(os.Stderr, "keel: icon %s is %dpx; give 1024px so every size stays sharp\n", cfg.Icon, b.Dx())
	}
	return &iconSet{dir: dir, cfg: cfg, art: art}, nil
}

// icon is the platform's icon at size: the finished override if keel.json
// names one, else the artwork cut to the platform's shape.
func (s *iconSet) icon(platform string, size int) (image.Image, error) {
	if path := s.cfg.Icons[platform]; path != "" {
		img, err := decodePNG(filepath.Join(s.dir, path))
		if err != nil {
			return nil, err
		}
		if b := img.Bounds(); b.Dx() != b.Dy() {
			return nil, fmt.Errorf("icons.%s %s is not square", platform, path)
		}
		dst := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
		return dst, nil
	}
	shape := map[string]iconShape{"darwin": macShape, "windows": winShape, "linux": linuxShape}[platform]
	return shape.render(s.art, size, s.cfg.IconMask != "none"), nil
}

// windowsIcon holds every Windows size, each drawn at its own size so small
// ones stay crisp.
func (s *iconSet) windowsIcon() (*winres.Icon, error) {
	var imgs []image.Image
	for _, size := range windowsIconSizes {
		img, err := s.icon("windows", size)
		if err != nil {
			return nil, err
		}
		imgs = append(imgs, img)
	}
	return winres.NewIconFromImages(imgs)
}

func writePNG(path string, img image.Image) error {
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// iconCommand writes every platform's icons for a look before building.
func (c *cli) iconCommand(args []string) error {
	fs := flag.NewFlagSet("icon", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	out := fs.String("o", filepath.Join("dist", "icons"), "output directory")
	fs.Usage = func() {
		fmt.Fprintln(c.errw, "Usage: keel icon [-o dir]\n\nWrites the macOS, Windows and Linux icons made from keel.json's icon.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	dir := c.wd()
	cfg, err := loadConfig(dir)
	if err != nil {
		return err
	}
	set, err := loadIcons(dir, cfg)
	if err != nil {
		return err
	}
	outDir := *out
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(dir, outDir)
	}
	mac, err := set.icon("darwin", 1024)
	if err != nil {
		return err
	}
	if err := writePNG(filepath.Join(outDir, "macos.png"), mac); err != nil {
		return err
	}
	for _, size := range windowsIconSizes {
		img, err := set.icon("windows", size)
		if err != nil {
			return err
		}
		if err := writePNG(filepath.Join(outDir, "windows", strconv.Itoa(size)+".png"), img); err != nil {
			return err
		}
	}
	if err := writeICO(set, filepath.Join(outDir, "windows.ico")); err != nil {
		return err
	}
	for _, size := range linuxIconSizes {
		img, err := set.icon("linux", size)
		if err != nil {
			return err
		}
		if err := writePNG(filepath.Join(outDir, "linux", strconv.Itoa(size)+".png"), img); err != nil {
			return err
		}
	}
	fmt.Fprintln(c.out, "Wrote macOS, Windows and Linux icons to", outDir)
	return nil
}

func writeICO(set *iconSet, path string) error {
	ico, err := set.windowsIcon()
	if err != nil {
		return err
	}
	var b bytes.Buffer
	if err := ico.SaveICO(&b); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

var errOtherSyso = errors.New("the main package already has .syso resources; remove them, keel build writes the icon, manifest and version itself")
