package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func iosArchitecture(arch string, device bool) (string, error) {
	if arch == "" {
		arch = runtime.GOARCH
		if device {
			arch = "arm64"
		}
	}
	if arch != "arm64" && (device || arch != "amd64") {
		return "", fmt.Errorf("iOS arch %q: use arm64 for devices, arm64 or amd64 for simulators (one architecture per build)", arch)
	}
	return arch, nil
}

func (c *cli) buildIOS(dir string, cfg *Config, icons *iconSet, outDir, main, arch, sign, provision string, device bool) error {
	if runtime.GOOS != "darwin" && !c.dryRun {
		return errors.New("iOS apps are built on a Mac with full Xcode and its iOS SDKs")
	}
	arch, err := iosArchitecture(arch, device)
	if err != nil {
		return err
	}
	if device && (sign == "" || sign == "-" || provision == "") {
		return errors.New("iOS devices require -sign \"Apple Development/Distribution: …\" and -provision path.mobileprovision")
	}
	if !device && (sign != "" || provision != "") {
		return errors.New("simulators use an ad hoc signature; use -device for -sign and -provision")
	}
	sdk, platform := "iphonesimulator", "iPhoneSimulator"
	tripleArch := arch
	if arch == "amd64" {
		tripleArch = "x86_64"
	}
	triple := tripleArch + "-apple-ios" + cfg.iosMinimumVersion() + "-simulator"
	if device {
		sdk, platform = "iphoneos", "iPhoneOS"
		triple = "arm64-apple-ios" + cfg.iosMinimumVersion()
	}
	sdkPath, compiler := "<"+sdk+"-sdk>", "clang"
	if c.dryRun {
		c.command(dir, nil, "xcrun", "--sdk", sdk, "--show-sdk-path")
		c.command(dir, nil, "xcrun", "--sdk", sdk, "--find", "clang")
	} else {
		if sdkPath, err = toolOutput(dir, "xcrun", "--sdk", sdk, "--show-sdk-path"); err != nil {
			return fmt.Errorf("install full Xcode and select its developer directory: %w", err)
		}
		if compiler, err = toolOutput(dir, "xcrun", "--sdk", sdk, "--find", "clang"); err != nil {
			return err
		}
	}
	base := filepath.Join(outDir, "ios")
	if err := c.mkdir(base); err != nil {
		return err
	}
	work := filepath.Join(base, "<ios-build>")
	if !c.dryRun {
		work, err = os.MkdirTemp(base, ".keel-build-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(work)
	}
	app := filepath.Join(work, "Payload", cfg.Binary+".app")
	if err := c.mkdir(app); err != nil {
		return err
	}
	goArgs := []string{"build", c.trimFlag(), "-ldflags", c.ldflags(appIDFlag(cfg))}
	if !device {
		modfile := filepath.Join(work, "go.mod")
		if c.dryRun {
			fmt.Fprintln(c.out, "prepare simulator Metal shaders in", modfile)
		} else if err := prepareIOSSimulatorModule(dir, work); err != nil {
			return err
		}
		goArgs = append(goArgs, "-modfile", modfile)
	}
	goArgs = append(goArgs, "-o", filepath.Join(app, cfg.Binary), main)
	// Quote the SDK path for cgo's flag parser as Xcode may live in a path
	// containing spaces. The compiler path itself is a single CC executable.
	flags := fmt.Sprintf("-target %s -isysroot %q -fobjc-arc", triple, sdkPath)
	env := []string{"GOOS=ios", "GOARCH=" + arch, "CGO_ENABLED=1", "CC=" + strconv.Quote(compiler),
		"CGO_CFLAGS=" + flags, "CGO_LDFLAGS=" + flags}
	if !device {
		env = append(env, "GOWORK=off") // Go does not permit -modfile in workspace mode.
	}
	if err := c.command(dir, env, "go", goArgs...); err != nil {
		return err
	}
	if c.dryRun {
		fmt.Fprintln(c.out, "write", filepath.Join(app, "Info.plist"), "(name, bundle ID, version, KeelSceneDelegate)")
	} else if err := os.WriteFile(filepath.Join(app, "Info.plist"), []byte(iosInfoPlist(cfg, platform)), 0o644); err != nil {
		return err
	}
	if err := c.iosIcons(dir, work, app, sdk, cfg, icons); err != nil {
		return err
	}
	if device {
		if !filepath.IsAbs(provision) {
			provision = filepath.Join(dir, provision)
		}
		if err := c.signIOSDevice(dir, work, app, cfg, sign, provision); err != nil {
			return err
		}
	} else if err := c.command(dir, nil, "codesign", "--force", "--sign", "-", app); err != nil {
		return err
	}
	out := filepath.Join(base, cfg.Binary+".app")
	if device {
		out = filepath.Join(base, cfg.Binary+".ipa")
		if c.dryRun {
			fmt.Fprintln(c.out, "archive", filepath.Join(work, "Payload"), "to", out)
		} else if err := zipIOSPayload(work, out); err != nil {
			return err
		}
	} else if c.dryRun {
		fmt.Fprintln(c.out, "move", app, "to", out)
	} else {
		// Publish only after compilation, assets and signing have succeeded.
		if err := os.RemoveAll(out); err != nil {
			return err
		}
		if err := os.Rename(app, out); err != nil {
			return err
		}
	}
	c.done(out)
	return nil
}

func xmlText(s string) string {
	var b bytes.Buffer
	xml.EscapeText(&b, []byte(s))
	return b.String()
}

func iosInfoPlist(cfg *Config, platform string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleDisplayName</key><string>%[1]s</string>
<key>CFBundleName</key><string>%[1]s</string>
<key>CFBundleExecutable</key><string>%[2]s</string>
<key>CFBundleIdentifier</key><string>%[3]s</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleShortVersionString</key><string>%[4]s</string>
<key>CFBundleVersion</key><string>%[5]d</string>
<key>CFBundleSupportedPlatforms</key><array><string>%[6]s</string></array>
<key>MinimumOSVersion</key><string>%[7]s</string>
<key>UIDeviceFamily</key><array><integer>1</integer><integer>2</integer></array>
<key>UIApplicationSceneManifest</key><dict>
<key>UIApplicationSupportsMultipleScenes</key><false/>
<key>UISceneConfigurations</key><dict>
<key>UIWindowSceneSessionRoleApplication</key><array><dict>
<key>UISceneConfigurationName</key><string>Keel</string>
<key>UISceneClassName</key><string>UIWindowScene</string>
<key>UISceneDelegateClassName</key><string>KeelSceneDelegate</string>
</dict></array></dict></dict>
<key>UILaunchScreen</key><dict/>
<key>UISupportedInterfaceOrientations</key><array>
<string>UIInterfaceOrientationPortrait</string>
<string>UIInterfaceOrientationLandscapeLeft</string>
<string>UIInterfaceOrientationLandscapeRight</string>
</array>
<key>UISupportedInterfaceOrientations~ipad</key><array>
<string>UIInterfaceOrientationPortrait</string>
<string>UIInterfaceOrientationPortraitUpsideDown</string>
<string>UIInterfaceOrientationLandscapeLeft</string>
<string>UIInterfaceOrientationLandscapeRight</string>
</array>
</dict></plist>
`, xmlText(cfg.Name), xmlText(cfg.Binary), xmlText(cfg.AppID), cfg.Version, max(cfg.Build, 1), platform, cfg.iosMinimumVersion())
}

// toolOutput captures machine-readable output without mixing it with diagnostics.
func toolOutput(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(diagnostic.String()))
	}
	return strings.TrimSpace(string(out)), nil
}

// The simulator needs a temporary copy of shader v1.0.9, which identifies
// simulators by amd64 alone. The real module and cache are never edited.
func prepareIOSSimulatorModule(dir, work string) error {
	data, err := toolOutput(dir, "go", "list", "-m", "-json", "gioui.org/shader")
	if err != nil {
		return err
	}
	var module struct{ Dir string }
	if err := json.Unmarshal([]byte(data), &module); err != nil {
		return err
	}
	if module.Dir == "" {
		return errors.New("shader sources are unavailable; run go mod download first")
	}
	copy := filepath.Join(work, "shader")
	if err := copyDirectory(module.Dir, copy); err != nil {
		return err
	}
	for _, pkg := range []string{"gio", "piet"} {
		file := filepath.Join(copy, pkg, "shaders.go")
		original, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		condition := `runtime.GOARCH == "amd64"`
		if !bytes.Contains(original, []byte(condition)) {
			return errors.New("shader simulator selection changed; update Keel's iOS compatibility adapter")
		}
		patched := strings.ReplaceAll(string(original), condition, "("+condition+` || runtime.GOARCH == "arm64")`)
		if err := os.WriteFile(file, []byte(patched), 0o644); err != nil {
			return err
		}
	}
	for _, name := range []string{"go.mod", "go.sum"} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if name == "go.sum" && errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(work, name), data, 0o644); err != nil {
			return err
		}
	}
	// Resolve relative local replacements before moving the module file.
	data, err = toolOutput(dir, "go", "mod", "edit", "-json")
	if err != nil {
		return err
	}
	var mod struct {
		Replace []struct {
			Old, New struct{ Path, Version string }
		}
	}
	if err := json.Unmarshal([]byte(data), &mod); err != nil {
		return err
	}
	args := []string{"mod", "edit", "-modfile", filepath.Join(work, "go.mod")}
	for _, replacement := range mod.Replace {
		old := replacement.Old.Path
		if replacement.Old.Version != "" {
			old += "@" + replacement.Old.Version
		}
		if replacement.Old.Path == "gioui.org/shader" {
			args = append(args, "-dropreplace", old)
			continue
		}
		if replacement.New.Version != "" || filepath.IsAbs(replacement.New.Path) {
			continue
		}
		args = append(args, "-replace", old+"="+filepath.Join(dir, replacement.New.Path))
	}
	args = append(args, "-replace", "gioui.org/shader="+copy)
	_, err = toolOutput(dir, "go", args...)
	return err
}

func copyDirectory(source, target string) error {
	return filepath.WalkDir(source, func(file string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, file)
		if err != nil {
			return err
		}
		dest := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("unsupported file in shader module: %s", file)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0o644)
	})
}

func (c *cli) iosIcons(dir, work, app, sdk string, cfg *Config, icons *iconSet) error {
	assets := filepath.Join(work, "Assets.xcassets")
	set := filepath.Join(assets, "AppIcon.appiconset")
	partial := filepath.Join(work, "assets.plist")
	if c.dryRun {
		fmt.Fprintln(c.out, "write", set, "(opaque iPhone/iPad icons)")
	} else {
		if err := os.MkdirAll(set, 0o755); err != nil {
			return err
		}
		var images []map[string]string
		for _, item := range []struct {
			idiom, size, scale string
			pixels             int
		}{
			{"iphone", "20x20", "2x", 40}, {"iphone", "20x20", "3x", 60},
			{"iphone", "29x29", "2x", 58}, {"iphone", "29x29", "3x", 87},
			{"iphone", "40x40", "2x", 80}, {"iphone", "40x40", "3x", 120},
			{"iphone", "60x60", "2x", 120}, {"iphone", "60x60", "3x", 180},
			{"ipad", "20x20", "1x", 20}, {"ipad", "20x20", "2x", 40},
			{"ipad", "29x29", "1x", 29}, {"ipad", "29x29", "2x", 58},
			{"ipad", "40x40", "1x", 40}, {"ipad", "40x40", "2x", 80},
			{"ipad", "76x76", "1x", 76}, {"ipad", "76x76", "2x", 152},
			{"ipad", "83.5x83.5", "2x", 167}, {"ios-marketing", "1024x1024", "1x", 1024},
		} {
			filename := item.idiom + "-" + item.size + "-" + item.scale + ".png"
			art, err := icons.icon("ios", item.pixels)
			if err != nil {
				return err
			}
			if err := writePNG(filepath.Join(set, filename), opaqueIOSIcon(art)); err != nil {
				return err
			}
			images = append(images, map[string]string{"idiom": item.idiom, "size": item.size, "scale": item.scale, "filename": filename})
		}
		data, err := json.Marshal(map[string]any{"images": images, "info": map[string]any{"version": 1, "author": "keel"}})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(set, "Contents.json"), data, 0o644); err != nil {
			return err
		}
	}
	if err := c.command(dir, nil, "xcrun", "actool", "--compile", app, "--platform", sdk,
		"--output-format", "human-readable-text",
		"--minimum-deployment-target", cfg.iosMinimumVersion(), "--target-device", "iphone", "--target-device", "ipad",
		"--app-icon", "AppIcon", "--output-partial-info-plist", partial, assets); err != nil {
		return err
	}
	return c.command(dir, nil, "/usr/libexec/PlistBuddy", "-c", "Merge "+strconv.Quote(partial), filepath.Join(app, "Info.plist"))
}

func (c *cli) signIOSDevice(dir, work, app string, cfg *Config, identity, profile string) error {
	decoded, entitlements := filepath.Join(work, "profile.plist"), filepath.Join(work, "entitlements.plist")
	if err := c.command(dir, nil, "security", "cms", "-D", "-i", profile, "-o", decoded); err != nil {
		return err
	}
	if c.dryRun {
		fmt.Fprintln(c.out, "extract matching profile entitlements to", entitlements)
		fmt.Fprintln(c.out, "copy", profile, "to", filepath.Join(app, "embedded.mobileprovision"))
	} else {
		appID, err := toolOutput(dir, "/usr/libexec/PlistBuddy", "-c", "Print :Entitlements:application-identifier", decoded)
		if err != nil {
			return err
		}
		concrete, err := iosProfileAppID(appID, cfg.AppID)
		if err != nil {
			return err
		}
		data, err := toolOutput(dir, "/usr/libexec/PlistBuddy", "-x", "-c", "Print :Entitlements", decoded)
		if err != nil {
			return err
		}
		if err := os.WriteFile(entitlements, []byte(data), 0o600); err != nil {
			return err
		}
		if _, err := toolOutput(dir, "/usr/libexec/PlistBuddy", "-c", "Set :application-identifier "+concrete, entitlements); err != nil {
			return err
		}
		// Wildcard profiles need concrete keychain groups in the signature.
		groups, err := toolOutput(dir, "plutil", "-extract", "keychain-access-groups", "json", "-o", "-", entitlements)
		if err == nil {
			var values []string
			if err := json.Unmarshal([]byte(groups), &values); err != nil {
				return err
			}
			for i, value := range values {
				if strings.Contains(value, "*") {
					values[i] = concrete
				}
			}
			data, _ := json.Marshal(values)
			if _, err := toolOutput(dir, "plutil", "-replace", "keychain-access-groups", "-json", string(data), entitlements); err != nil {
				return err
			}
		}
		dataProfile, err := os.ReadFile(profile)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(app, "embedded.mobileprovision"), dataProfile, 0o644); err != nil {
			return err
		}
	}
	return c.command(dir, nil, "codesign", "--force", "--sign", identity, "--entitlements", entitlements, app)
}

func iosProfileAppID(profileID, appID string) (string, error) {
	prefix, pattern, found := strings.Cut(profileID, ".")
	match, err := path.Match(pattern, appID)
	if !found || prefix == "" || err != nil || !match {
		return "", fmt.Errorf("provisioning profile does not match bundle ID %q", appID)
	}
	return prefix + "." + appID, nil
}

func zipIOSPayload(work, output string) error {
	file, err := os.CreateTemp(filepath.Dir(output), ".keel-ipa-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	archive := zip.NewWriter(file)
	err = filepath.WalkDir(filepath.Join(work, "Payload"), func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(work, file)
		header.Name, header.Method = filepath.ToSlash(rel), zip.Deflate
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		reader, err := os.Open(file)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, reader)
		reader.Close()
		return err
	})
	err = errors.Join(err, archive.Close(), file.Close())
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), output)
}
