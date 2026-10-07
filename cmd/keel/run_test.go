package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRunIconConfiguration(t *testing.T) {
	dir := t.TempDir()
	art := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			art.SetNRGBA(x, y, color.NRGBA{R: 220, A: 255})
		}
	}
	if err := writePNG(filepath.Join(dir, "art.png"), art); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, mask string
		override   bool
		corner     bool
	}{
		{"platform", "platform", false, false},
		{"none", "none", false, true},
		{"override", "platform", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{Icon: "art.png", IconMask: tc.mask}
			if tc.override {
				cfg.Icon = "missing.png"
				cfg.Icons = map[string]string{"darwin": "art.png"}
			}
			path, cleanup, err := prepareRunIcon(dir, cfg, "darwin")
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()
			img, err := decodePNG(path)
			if err != nil {
				t.Fatal(err)
			}
			if img.Bounds() != image.Rect(0, 0, 1024, 1024) {
				t.Fatal(img.Bounds())
			}
			pos := 110
			if tc.override {
				pos = 0
			}
			_, _, _, alpha := img.At(pos, pos).RGBA()
			if (alpha == 65535) != tc.corner {
				t.Fatalf("corner alpha=%d", alpha)
			}
			cleanup()
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatal("temporary icon not removed", err)
			}
		})
	}
	if err := os.WriteFile(filepath.Join(dir, "invalid.png"), []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, cleanup, err := prepareRunIcon(dir, &Config{Icon: "invalid.png"}, "darwin"); err == nil {
		cleanup()
		t.Fatal("invalid icon accepted")
	}
	if _, cleanup, err := prepareRunIcon(dir, &Config{Icon: "absent.png"}, "darwin"); err == nil {
		cleanup()
		t.Fatal("missing configured icon accepted")
	}
	path, cleanup, err := prepareRunIcon(dir, &Config{Icon: "appicon.png"}, "darwin")
	cleanup()
	if err != nil || path != "" {
		t.Fatal("legacy project without artwork", path, err)
	}
}

func TestRunPassesIconAndArguments(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("test go shim uses a POSIX shell")
	}
	dir := t.TempDir()
	cfg := &Config{Name: "Test", AppID: "com.example.test", Version: "1.0.0", Binary: "test", Icon: "art.png", Main: "."}
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	if err := writePNG(filepath.Join(dir, "art.png"), image.NewNRGBA(image.Rect(0, 0, 512, 512))); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	shim := `#!/bin/sh
printf '%s\n' "$KEEL_RUN_ICON"
cp "$KEEL_RUN_ICON" "$KEEL_ICON_TEST_COPY"
printf '%s\n' "$@"
`
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(shim), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	copyPath := filepath.Join(dir, "received.png")
	t.Setenv("KEEL_ICON_TEST_COPY", copyPath)
	c, out := newCLI(dir)
	if err := c.runProject([]string{"-watch=false", "--", "--title", "two words"}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if !strings.Contains(out.String(), "--title\ntwo words") {
		t.Fatal(out.String())
	}
	if _, err := os.Stat(lines[0]); !os.IsNotExist(err) {
		t.Fatal("icon leaked after exit", err)
	}
	data, err := os.ReadFile(copyPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
}
