package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var androidAppIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)

func androidArchitectures(value string) (string, error) {
	if value == "" {
		return "arm64,amd64", nil
	}
	var result []string
	seen := make(map[string]bool)
	for _, arch := range strings.Split(value, ",") {
		arch = strings.TrimSpace(arch)
		switch arch {
		case "arm", "arm64", "386", "amd64":
		default:
			return "", fmt.Errorf("Android arch %q: use arm, arm64, 386 or amd64", arch)
		}
		if seen[arch] {
			return "", fmt.Errorf("duplicate Android architecture %q", arch)
		}
		seen[arch] = true
		result = append(result, arch)
	}
	return strings.Join(result, ","), nil
}

func androidSDK() string {
	if root := os.Getenv("ANDROID_HOME"); root != "" {
		return root
	}
	return os.Getenv("ANDROID_SDK_ROOT")
}

func androidTool(name string) string {
	if name == "adb" {
		candidate := filepath.Join(androidSDK(), "platform-tools", name+androidExeSuffix())
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	if name == "keytool" && os.Getenv("JAVA_HOME") != "" {
		return filepath.Join(os.Getenv("JAVA_HOME"), "bin", name+androidExeSuffix())
	}
	return name
}

func androidExeSuffix() string {
	if os.PathSeparator == '\\' {
		return ".exe"
	}
	return ""
}

// buildAndroid delegates resource packaging and cgo compilation to the pinned
// Gio packager. Keep a project-local development key across builds so adb can
// replace an installed APK without losing its application data.
func (c *cli) buildAndroid(dir string, cfg *Config, icons *iconSet, outDir, main, arches string) error {
	if !androidAppIDPattern.MatchString(cfg.AppID) {
		return fmt.Errorf("Android appid %q: use a Java package name such as com.example.myapp (no hyphens)", cfg.AppID)
	}
	arches, err := androidArchitectures(arches)
	if err != nil {
		return err
	}
	if !c.dryRun && androidSDK() == "" {
		return errors.New("Android SDK missing: set ANDROID_HOME; check keel doctor -target android")
	}
	base := filepath.Join(outDir, "android")
	if err := c.mkdir(base); err != nil {
		return err
	}
	key := filepath.Join(base, "debug.keystore")
	if _, err := os.Stat(key); errors.Is(err, os.ErrNotExist) {
		if err := c.command(dir, nil, androidTool("keytool"), "-genkeypair", "-keystore", key,
			"-storepass", "android", "-keypass", "android", "-alias", "androiddebugkey", "-dname", "CN=Android Debug,O=Android,C=US",
			"-keyalg", "RSA", "-keysize", "2048", "-validity", "10000", "-storetype", "JKS", "-noprompt"); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	icon := filepath.Join(base, "<android-icon>.png")
	if c.dryRun {
		fmt.Fprintln(c.out, "write", icon)
	} else {
		work, err := os.MkdirTemp(base, ".keel-icon-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(work)
		icon = filepath.Join(work, "android.png")
		art, err := icons.icon("android", 1024)
		if err != nil {
			return err
		}
		if err := writePNG(icon, art); err != nil {
			return err
		}
	}
	apk := filepath.Join(base, cfg.Binary+".apk")
	env := append(c.trimEnv(), "ANDROID_HOME="+androidSDK())
	if err := c.runGogio(dir, env, "-target", "android", "-arch", arches,
		"-minsdk", fmt.Sprint(cfg.androidMinimumSDK()), "-targetsdk", fmt.Sprint(cfg.androidTargetSDK()),
		"-appid", cfg.AppID, "-name", cfg.Name, "-version", cfg.fourPart(), "-icon", icon,
		"-signkey", key, "-signpass", "android", "-ldflags", c.ldflags(appIDFlag(cfg)), "-o", apk, main); err != nil {
		return err
	}
	c.done(apk)
	return nil
}

func selectAndroidDevice(data, requested string) (string, error) {
	var ready []string
	for _, line := range strings.Split(data, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] == "List" || strings.HasPrefix(fields[0], "*") {
			continue
		}
		if fields[0] == requested && fields[1] != "device" {
			return "", fmt.Errorf("Android device %s is %s; authorize USB debugging and check adb devices", requested, fields[1])
		}
		if fields[1] == "device" {
			if fields[0] == requested {
				return requested, nil
			}
			ready = append(ready, fields[0])
		}
	}
	if requested != "" {
		return "", fmt.Errorf("Android device %q is unavailable; check adb devices", requested)
	}
	if len(ready) == 0 {
		return "", errors.New("no ready Android device; start an emulator or connect a device with USB debugging enabled (adb devices)")
	}
	if len(ready) > 1 {
		return "", errors.New("multiple Android devices are connected; select one with -serial <serial>")
	}
	return ready[0], nil
}

func androidABIArchitecture(abi string) (string, error) {
	switch strings.TrimSpace(abi) {
	case "arm64-v8a":
		return "arm64", nil
	case "armeabi-v7a":
		return "arm", nil
	case "x86_64":
		return "amd64", nil
	case "x86":
		return "386", nil
	}
	return "", fmt.Errorf("unsupported Android device ABI %q", abi)
}

func (c *cli) runAndroid(dir string, cfg *Config, serial string, args []string) error {
	if len(args) != 0 {
		return errors.New("Android run does not pass command-line arguments to Go; configure the app in code")
	}
	c.release = false
	adb := androidTool("adb")
	arch := "arm64,amd64"
	if c.dryRun {
		if serial == "" {
			serial = "<device-serial>"
		}
		c.command(dir, nil, adb, "devices")
		c.command(dir, nil, adb, "-s", serial, "shell", "getprop", "ro.product.cpu.abi")
	} else {
		inventory, err := toolOutput(dir, adb, "devices")
		if err != nil {
			return err
		}
		serial, err = selectAndroidDevice(inventory, serial)
		if err != nil {
			return err
		}
		abi, err := toolOutput(dir, adb, "-s", serial, "shell", "getprop", "ro.product.cpu.abi")
		if err != nil {
			return err
		}
		arch, err = androidABIArchitecture(abi)
		if err != nil {
			return err
		}
	}
	icons, err := loadIcons(dir, cfg)
	if err != nil {
		return err
	}
	output := filepath.Join(dir, "dist")
	if err := c.buildAndroid(dir, cfg, icons, output, "./"+filepath.ToSlash(filepath.Clean(cfg.Main)), arch); err != nil {
		return err
	}
	if err := c.command(dir, nil, adb, "-s", serial, "install", "-r", filepath.Join(output, "android", cfg.Binary+".apk")); err != nil {
		return err
	}
	// adb shell joins arguments before sending them to the remote shell.
	// AppID is checked by buildAndroid and contains only package-name characters.
	component := cfg.AppID + "/org.gioui.GioActivity"
	if c.dryRun {
		return c.command(dir, nil, adb, "-s", serial, "shell", "am", "start", "-W", "-S", "-n", component)
	}
	cmd := exec.Command(adb, "-s", serial, "shell", "am", "start", "-W", "-S", "-n", component)
	cmd.Dir = dir
	data, err := cmd.CombinedOutput()
	fmt.Fprint(c.out, string(data))
	if err != nil {
		return fmt.Errorf("launch Android app: %w", err)
	}
	if strings.Contains(string(data), "Error:") || !strings.Contains(string(data), "Status: ok") {
		return fmt.Errorf("Android activity failed to start: %s", strings.TrimSpace(string(data)))
	}
	return nil
}
