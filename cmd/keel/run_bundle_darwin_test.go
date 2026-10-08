package main

import (
	"context"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBundleDockResources(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{Name: "Test & Dock", AppID: "com.example.dock", Version: "1.2.3", Binary: "dock", Icon: "art.png", IconMask: "none", Main: "."}
	if err := cfg.save(dir); err != nil {
		t.Fatal(err)
	}
	art := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	for y := range 512 {
		for x := range 512 {
			art.SetNRGBA(x, y, color.NRGBA{R: 220, G: 30, B: 70, A: 255})
		}
	}
	if err := writePNG(filepath.Join(dir, cfg.Icon), art); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module docktest\n\ngo 1.26.1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c, _ := newCLI(dir)
	build := c.compileRun(context.Background(), dir, t.TempDir(), 1)
	defer build.cleanup()
	if build.err != nil {
		t.Fatal(build.err)
	}
	bundle := filepath.Dir(filepath.Dir(filepath.Dir(build.binary)))
	if !strings.HasSuffix(bundle, ".app") || filepath.Base(build.binary) != cfg.Binary {
		t.Fatal("run did not produce a macOS bundle", build.binary)
	}
	if output, err := exec.Command("codesign", "--verify", "--strict", bundle).CombinedOutput(); err != nil {
		t.Fatalf("invalid development bundle signature: %v %s", err, output)
	}
	plist := filepath.Join(bundle, "Contents", "Info.plist")
	for key, expected := range map[string]string{
		"CFBundleIdentifier": cfg.AppID, "CFBundleName": cfg.Name,
		"CFBundleExecutable": cfg.Binary, "CFBundleIconFile": "icon.icns",
	} {
		value, err := exec.Command("plutil", "-extract", key, "raw", "-o", "-", plist).Output()
		if err != nil || strings.TrimSpace(string(value)) != expected {
			t.Fatalf("%s: %s (%v)", key, value, err)
		}
	}
	// Round-trip with Apple's decoder, ensuring the Dock has all icon sizes.
	decoded := filepath.Join(dir, "decoded.iconset")
	output, err := exec.Command("iconutil", "-c", "iconset", filepath.Join(bundle, "Contents", "Resources", "icon.icns"), "-o", decoded).CombinedOutput()
	if err != nil {
		t.Fatalf("unreadable Dock icon: %v %s", err, output)
	}
	actual, err := decodePNG(filepath.Join(decoded, "icon_512x512@2x.png"))
	if err != nil || actual.Bounds().Dx() != 1024 {
		t.Fatal("missing Retina Dock artwork", err)
	}
	expected, err := decodePNG(build.icon)
	if err != nil {
		t.Fatal(err)
	}
	for _, pos := range []image.Point{{0, 0}, {110, 110}, {512, 512}} {
		if actual.At(pos.X, pos.Y) != expected.At(pos.X, pos.Y) {
			t.Fatal("bundle icon changed the finished artwork", pos)
		}
	}
	build.cleanup()
	for _, path := range []string{bundle, build.icon} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("development resources leaked", path, err)
		}
	}
}

func TestRunBundleWithoutLegacyArtwork(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "legacy-1")
	source := filepath.Join(root, "main.go")
	if err := os.WriteFile(source, []byte("package main\nfunc main() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command("go", "build", "-o", binary, source).CombinedOutput(); err != nil {
		t.Fatalf("compile fixture: %v %s", err, output)
	}
	executable, cleanup, err := prepareRunBundle(context.Background(), &Config{Name: "Legacy", AppID: "com.example.legacy", Version: "1.0.0", Binary: "legacy"}, binary, "")
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(executable); err != nil {
		t.Fatal(err)
	}
}
