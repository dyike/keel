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

	"github.com/tc-hib/winres"
	"github.com/tc-hib/winres/version"
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
	icons, err := loadIcons(dir, cfg)
	if err != nil {
		return err
	}
	main := "./" + filepath.ToSlash(filepath.Clean(cfg.Main))
	switch *target {
	case "darwin", "macos":
		return c.buildDarwin(dir, cfg, icons, outDir, main, *arch, *sign)
	case "windows":
		return c.buildWindows(dir, cfg, icons, outDir, main, *arch)
	case "linux":
		return c.buildLinux(dir, cfg, icons, outDir, main, *arch)
	case "js", "web":
		return c.command(dir, nil, "go", "run", gogio, "-target", "js", "-o", filepath.Join(outDir, "web"), main)
	}
	return fmt.Errorf("unknown target %q: use darwin, windows, linux or js", *target)
}

func (c *cli) buildDarwin(dir string, cfg *Config, icons *iconSet, outDir, main, arch, sign string) error {
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
	// gogio makes the .icns sizes from one 1024px image: the macOS one.
	tmp, err := os.MkdirTemp("", "keel-icon")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	iconPath := filepath.Join(tmp, "macos.png")
	mac, err := icons.icon("darwin", 1024)
	if err != nil {
		return err
	}
	if err := writePNG(iconPath, mac); err != nil {
		return err
	}
	if err := c.command(dir, nil, "go", "run", gogio, "-target", "macos", "-arch", arch,
		"-appid", cfg.AppID, "-version", cfg.fourPart(), "-icon", iconPath, "-o", app, main); err != nil {
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

// buildWindows links a .syso with the icon (every size Windows asks for),
// a manifest (per-monitor DPI, common controls 6) and version information,
// then builds a GUI executable. Windows needs no cgo, so any OS will do.
func (c *cli) buildWindows(dir string, cfg *Config, icons *iconSet, outDir, main, arches string) error {
	if arches == "" {
		arches = "amd64"
	}
	mainDir := filepath.Join(dir, filepath.FromSlash(main))
	if others, _ := filepath.Glob(filepath.Join(mainDir, "*.syso")); len(others) > 0 {
		return fmt.Errorf("%w: %s", errOtherSyso, strings.Join(others, ", "))
	}
	if err := c.mkdir(outDir); err != nil {
		return err
	}
	rs, err := windowsResources(cfg, icons)
	if err != nil {
		return err
	}
	list := strings.Split(arches, ",")
	for _, arch := range list {
		arch = strings.TrimSpace(arch)
		if arch != "amd64" && arch != "arm64" && arch != "386" {
			return fmt.Errorf("windows arch %q: use amd64, arm64 or 386", arch)
		}
		exe := filepath.Join(outDir, cfg.Binary+".exe")
		if len(list) > 1 {
			exe = filepath.Join(outDir, cfg.Binary+"-"+arch+".exe")
		}
		syso := filepath.Join(mainDir, "zz_keel_windows_"+arch+".syso")
		if c.dryRun {
			fmt.Fprintln(c.out, "write", syso, "(icon, manifest, version)")
		} else {
			var b bytes.Buffer
			if err := rs.WriteObject(&b, winres.Arch(arch)); err != nil {
				return err
			}
			if err := os.WriteFile(syso, b.Bytes(), 0o644); err != nil {
				return err
			}
		}
		err := c.command(dir, []string{"GOOS=windows", "GOARCH=" + arch, "CGO_ENABLED=0"},
			"go", "build", "-ldflags", "-H=windowsgui", "-o", exe, main)
		os.Remove(syso)
		if err != nil {
			return err
		}
		c.done(exe)
	}
	// The same icon as a file, for installers and shortcuts.
	if c.dryRun {
		fmt.Fprintln(c.out, "write", filepath.Join(outDir, cfg.Binary+".ico"))
		return nil
	}
	return writeICO(icons, filepath.Join(outDir, cfg.Binary+".ico"))
}

func windowsResources(cfg *Config, icons *iconSet) (*winres.ResourceSet, error) {
	rs := &winres.ResourceSet{}
	ico, err := icons.windowsIcon()
	if err != nil {
		return nil, err
	}
	if err := rs.SetIcon(winres.Name("APPICON"), ico); err != nil {
		return nil, err
	}
	rs.SetManifest(winres.AppManifest{
		Description:         cfg.Name,
		Compatibility:       winres.Win10AndAbove,
		ExecutionLevel:      winres.AsInvoker,
		DPIAwareness:        winres.DPIPerMonitorV2,
		UseCommonControlsV6: true,
		LongPathAware:       true,
	})
	// One string table (en-US), and the numeric versions set directly:
	// SetFileVersion would add a second, language-neutral table.
	vi := version.Info{}
	vi.FileVersion, vi.ProductVersion = cfg.versionWords(), cfg.versionWords()
	for key, value := range map[string]string{
		version.ProductName: cfg.Name, version.FileDescription: cfg.Name,
		version.OriginalFilename: cfg.Binary + ".exe", version.InternalName: cfg.Binary,
		version.ProductVersion: cfg.Version, version.FileVersion: cfg.Version,
	} {
		if err := vi.Set(version.LangDefault, key, value); err != nil {
			return nil, err
		}
	}
	rs.SetVersionInfo(vi)
	return rs, nil
}

// buildLinux writes the binary with its app ID, a .desktop entry named
// after the ID, the icon in the hicolor sizes, and an install script.
func (c *cli) buildLinux(dir string, cfg *Config, icons *iconSet, outDir, main, arch string) error {
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
	for _, size := range linuxIconSizes {
		img, err := icons.icon("linux", size)
		if err != nil {
			return err
		}
		path := filepath.Join(dist, "icons", "hicolor", fmt.Sprintf("%dx%d", size, size), "apps", cfg.AppID+".png")
		if err := writePNG(path, img); err != nil {
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
