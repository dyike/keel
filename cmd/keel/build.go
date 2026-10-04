package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/image/draw"
)

// gogio is Gio's packager, pinned to the release Keel builds on.
const gogio = "gioui.org/cmd/gogio@v0.10.0"

func (c *cli) build(args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(c.errw)
	target := fs.String("target", runtime.GOOS, "darwin, windows, linux or js")
	arch := fs.String("arch", "", "architectures, comma-separated (default: this machine's; amd64 for windows)")
	out := fs.String("o", "dist", "output directory")
	sign := fs.String("sign", "", "darwin: codesign identity, e.g. \"Developer ID Application: Name (TEAMID)\"")
	fs.BoolVar(&c.dryRun, "n", false, "print the commands instead of running them")
	fs.Usage = func() {
		fmt.Fprintln(c.errw, "Usage: keel build [flags]")
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
	outDir := *out
	if !filepath.IsAbs(outDir) {
		outDir = filepath.Join(dir, outDir)
	}
	icon := filepath.Join(dir, cfg.Icon)
	if err := checkIcon(icon); err != nil {
		return err
	}
	main := "./" + filepath.ToSlash(filepath.Clean(cfg.Main))
	switch *target {
	case "darwin", "macos":
		return c.buildDarwin(dir, cfg, outDir, icon, main, *arch, *sign)
	case "windows":
		return c.buildWindows(dir, cfg, outDir, icon, main, *arch)
	case "linux":
		return c.buildLinux(dir, cfg, outDir, icon, main, *arch)
	case "js", "web":
		return c.command(dir, nil, "go", "run", gogio, "-target", "js", "-o", filepath.Join(outDir, "web"), main)
	}
	return fmt.Errorf("unknown target %q: use darwin, windows, linux or js", *target)
}

func (c *cli) buildDarwin(dir string, cfg *Config, outDir, icon, main, arch, sign string) error {
	if runtime.GOOS != "darwin" && !c.dryRun {
		return errors.New("macOS apps are built on a Mac: the toolchain needs Xcode's cgo and iconutil")
	}
	if arch == "" {
		arch = runtime.GOARCH
	}
	app := filepath.Join(outDir, cfg.Name+".app")
	if err := c.mkdir(outDir); err != nil {
		return err
	}
	if err := c.command(dir, nil, "go", "run", gogio, "-target", "macos", "-arch", arch,
		"-appid", cfg.AppID, "-version", cfg.fourPart(), "-icon", icon, "-o", app, main); err != nil {
		return err
	}
	if err := c.writeInfoPlist(app, cfg); err != nil {
		return err
	}
	// Sign the whole bundle, so its signature covers the Info.plist just
	// written; without an identity, ad hoc ("-") for this Mac.
	signArgs := []string{"--force", "--deep", "--sign", "-", app}
	if sign != "" {
		signArgs = []string{"--force", "--deep", "--options", "runtime", "--timestamp", "--sign", sign, app}
	}
	if err := c.command(dir, nil, "codesign", signArgs...); err != nil {
		return err
	}
	c.done(app)
	if sign == "" && !c.dryRun {
		fmt.Fprintln(c.out, "Signed ad hoc, for this Mac. To share it, sign with -sign \"Developer ID Application: …\" and notarize (xcrun notarytool).")
	}
	return nil
}

func (c *cli) buildWindows(dir string, cfg *Config, outDir, icon, main, arch string) error {
	if arch == "" {
		arch = "amd64"
	}
	exe := filepath.Join(outDir, cfg.Binary+".exe")
	if err := c.mkdir(outDir); err != nil {
		return err
	}
	// gogio leaves the icon and manifest as .syso files beside the code;
	// keep only those that were already there.
	mainDir := filepath.Join(dir, filepath.FromSlash(main))
	before, _ := filepath.Glob(filepath.Join(mainDir, "*_windows_*.syso"))
	err := c.command(dir, nil, "go", "run", gogio, "-target", "windows", "-arch", arch,
		"-version", cfg.fourPart(), "-icon", icon, "-o", exe, main)
	after, _ := filepath.Glob(filepath.Join(mainDir, "*_windows_*.syso"))
	for _, f := range after {
		if !contains(before, f) {
			os.Remove(f)
		}
	}
	if err != nil {
		return err
	}
	c.done(exe)
	return nil
}

// buildLinux writes the binary with its app ID, a .desktop entry named
// after the ID, the icon in the hicolor sizes, and an install script.
func (c *cli) buildLinux(dir string, cfg *Config, outDir, icon, main, arch string) error {
	if runtime.GOOS != "linux" && !c.dryRun {
		return errors.New("Linux apps are built on Linux: Gio's window code needs cgo against Wayland and X11 headers")
	}
	dist := filepath.Join(outDir, "linux")
	if err := c.mkdir(dist); err != nil {
		return err
	}
	var env []string
	if arch != "" {
		env = []string{"GOARCH=" + arch}
	}
	if err := c.command(dir, env, "go", "build", "-ldflags", appIDFlag(cfg), "-o", filepath.Join(dist, cfg.Binary), main); err != nil {
		return err
	}
	if c.dryRun {
		fmt.Fprintf(c.out, "write %s.desktop, icons and install.sh in %s\n", cfg.AppID, dist)
		return nil
	}
	desktop := fmt.Sprintf("[Desktop Entry]\nType=Application\nName=%s\nExec=%s\nIcon=%s\nStartupWMClass=%s\nTerminal=false\nCategories=Utility;\n",
		cfg.Name, cfg.Binary, cfg.AppID, cfg.AppID)
	if err := os.WriteFile(filepath.Join(dist, cfg.AppID+".desktop"), []byte(desktop), 0o644); err != nil {
		return err
	}
	src, err := decodePNG(icon)
	if err != nil {
		return err
	}
	for _, size := range []int{128, 256, 512} {
		dst := image.NewNRGBA(image.Rect(0, 0, size, size))
		draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Src, nil)
		var b bytes.Buffer
		if err := png.Encode(&b, dst); err != nil {
			return err
		}
		path := filepath.Join(dist, "icons", "hicolor", fmt.Sprintf("%dx%d", size, size), "apps", cfg.AppID+".png")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, b.Bytes(), 0o644); err != nil {
			return err
		}
	}
	install := strings.NewReplacer("{{binary}}", cfg.Binary, "{{appid}}", cfg.AppID).Replace(linuxInstall)
	if err := os.WriteFile(filepath.Join(dist, "install.sh"), []byte(install), 0o755); err != nil {
		return err
	}
	c.done(dist)
	fmt.Fprintln(c.out, "Install for this user with", filepath.Join(dist, "install.sh"))
	return nil
}

const linuxInstall = `#!/bin/sh
# Installs {{binary}} for the current user, or under PREFIX (e.g. /usr/local).
set -e
here=$(cd "$(dirname "$0")" && pwd)
prefix=${PREFIX:-$HOME/.local}
install -Dm755 "$here/{{binary}}" "$prefix/bin/{{binary}}"
mkdir -p "$prefix/share/applications"
sed "s|^Exec=.*|Exec=$prefix/bin/{{binary}}|" "$here/{{appid}}.desktop" > "$prefix/share/applications/{{appid}}.desktop"
mkdir -p "$prefix/share/icons"
cp -R "$here/icons/hicolor" "$prefix/share/icons/"
command -v update-desktop-database >/dev/null && update-desktop-database "$prefix/share/applications" || true
command -v gtk-update-icon-cache >/dev/null && gtk-update-icon-cache -q "$prefix/share/icons/hicolor" || true
echo "Installed {{binary}} in $prefix"
`

func (c *cli) mkdir(dir string) error {
	if c.dryRun {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func (c *cli) done(path string) {
	if !c.dryRun {
		fmt.Fprintln(c.out, "Built", path)
	}
}

// checkIcon wants a square PNG; under 512 pixels is allowed but warned.
func checkIcon(path string) error {
	img, err := decodePNG(path)
	if err != nil {
		return err
	}
	b := img.Bounds()
	if b.Dx() != b.Dy() {
		return fmt.Errorf("icon %s is %dx%d; it must be square", filepath.Base(path), b.Dx(), b.Dy())
	}
	if b.Dx() < 512 {
		fmt.Fprintf(os.Stderr, "keel: icon %s is %dpx; 1024px looks sharp on every screen\n", filepath.Base(path), b.Dx())
	}
	return nil
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("icon: %w", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("icon %s is not a PNG: %w", filepath.Base(path), err)
	}
	return img, nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// writeInfoPlist replaces gogio's minimal Info.plist with a complete one:
// an application (not a bundle) with its name, versions and minimum macOS.
func (c *cli) writeInfoPlist(app string, cfg *Config) error {
	if c.dryRun {
		fmt.Fprintln(c.out, "write", filepath.Join(app, "Contents", "Info.plist"))
		return nil
	}
	exe, err := os.ReadDir(filepath.Join(app, "Contents", "MacOS"))
	if err != nil || len(exe) != 1 {
		return fmt.Errorf("unexpected bundle layout in %s", app)
	}
	esc := func(s string) string {
		var b bytes.Buffer
		xml.EscapeText(&b, []byte(s))
		return b.String()
	}
	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key><string>zh_CN</string>
	<key>CFBundleDisplayName</key><string>%[1]s</string>
	<key>CFBundleExecutable</key><string>%[2]s</string>
	<key>CFBundleIconFile</key><string>icon.icns</string>
	<key>CFBundleIdentifier</key><string>%[3]s</string>
	<key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
	<key>CFBundleName</key><string>%[1]s</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>%[4]s</string>
	<key>CFBundleVersion</key><string>%[5]d</string>
	<key>LSApplicationCategoryType</key><string>public.app-category.productivity</string>
	<key>LSMinimumSystemVersion</key><string>14.0</string>
	<key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
`, esc(cfg.Name), esc(exe[0].Name()), esc(cfg.AppID), cfg.Version, max(cfg.Build, 1))
	return os.WriteFile(filepath.Join(app, "Contents", "Info.plist"), []byte(plist), 0o644)
}
