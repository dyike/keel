package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/image/draw"
)

// A runtime NSApplication icon does not supply Launch Services with an app
// icon. Give development processes a real bundle, like a packaged build,
// while still running the executable directly to retain stdio and signals.
func prepareRunBundle(ctx context.Context, cfg *Config, binary, icon string) (string, func(), error) {
	bundle := binary + ".app"
	cleanup := func() { os.RemoveAll(bundle) }
	fail := func(err error) (string, func(), error) {
		cleanup()
		return binary, func() {}, err
	}
	macOS := filepath.Join(bundle, "Contents", "MacOS")
	resources := filepath.Join(bundle, "Contents", "Resources")
	for _, dir := range []string{macOS, resources} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fail(err)
		}
	}
	executable := filepath.Join(macOS, cfg.Binary)
	if err := os.Rename(binary, executable); err != nil {
		return fail(err)
	}
	if err := (&cli{}).writeInfoPlist(bundle, cfg); err != nil {
		return fail(err)
	}
	if icon != "" {
		if err := writeRunICNS(ctx, icon, resources); err != nil {
			return fail(err)
		}
	}
	// Seal the metadata and resources as well as the Go executable, just as
	// keel build does. No distribution identity is needed for local runs.
	cmd := exec.CommandContext(ctx, "codesign", "--force", "--sign", "-", bundle)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fail(fmt.Errorf("development bundle: codesign: %w\n%s", err, output))
	}
	return executable, cleanup, nil
}

// Resize the finished platform image; applying its mask again would shrink
// or round the artwork twice. iconutil is part of macOS, as in keel build.
func writeRunICNS(ctx context.Context, source, resources string) error {
	art, err := decodePNG(source)
	if err != nil {
		return err
	}
	set := filepath.Join(resources, "icon.iconset")
	defer os.RemoveAll(set)
	for _, size := range []int{16, 32, 128, 256, 512} {
		for _, retina := range []bool{false, true} {
			pixels, suffix := size, ""
			if retina {
				pixels, suffix = size*2, "@2x"
			}
			img := image.NewNRGBA(image.Rect(0, 0, pixels, pixels))
			draw.CatmullRom.Scale(img, img.Bounds(), art, art.Bounds(), draw.Src, nil)
			path := filepath.Join(set, fmt.Sprintf("icon_%dx%d%s.png", size, size, suffix))
			if err := writePNG(path, img); err != nil {
				return err
			}
		}
	}
	cmd := exec.CommandContext(ctx, "iconutil", "-c", "icns", set, "-o", filepath.Join(resources, "icon.icns"))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("development icon: iconutil: %w\n%s", err, output)
	}
	return nil
}
